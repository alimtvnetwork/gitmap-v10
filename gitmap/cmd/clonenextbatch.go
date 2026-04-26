package cmd

// Batch entry point for `gitmap cn`. Activated when the user passes
// `--csv <path>` OR `--all`, OR when the cwd is not itself a git repo
// but contains git-repo subdirectories one level down.
//
// Each repo in the batch picks its own next version via clonenext.ResolveTarget
// with arg "v++", so callers don't need to specify a version.

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/clonenext"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// batchRowResult records one repo's outcome for the CSV report.
type batchRowResult struct {
	RepoPath    string
	FromVersion string
	ToVersion   string
	Status      string // "ok" | "skipped" | "failed"
	Detail      string
}

// runCloneNextBatch is the dispatcher invoked by runCloneNext when batch
// mode is active. It loads the repo list, processes each one, and writes
// a CSV report. `maxConcurrency` controls the worker-pool size: 1 keeps
// the legacy sequential behavior with deterministic stdout ordering;
// values >1 fan repos out across a bounded pool that mirrors the main
// cloner's pattern (see gitmap/cloner/concurrent.go).
func runCloneNextBatch(csvPath string, walkAll bool, maxConcurrency int) {
	repos, err := loadBatchRepos(csvPath, walkAll)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextBatchLoad, err)
		os.Exit(1)
	}

	fmt.Printf(constants.MsgCloneNextBatchStart, len(repos))

	results := processBatchRepos(repos, maxConcurrency)
	reportPath := writeBatchReport(results)
	printBatchSummary(results, reportPath)
}

// loadBatchRepos resolves the input source (csv > walk > implicit walk)
// and returns the absolute repo paths to process.
func loadBatchRepos(csvPath string, walkAll bool) ([]string, error) {
	if len(csvPath) > 0 {
		return clonenext.LoadBatchFromCSV(csvPath)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	_ = walkAll // walkAll only matters as a dispatcher trigger; the walk itself is unconditional once we get here

	return clonenext.WalkBatchFromDir(cwd)
}

// processBatchRepos runs cn-equivalent steps for each repo and collects
// per-repo results without aborting on individual failures. Dispatches
// to the sequential or concurrent runner based on `maxConcurrency`.
func processBatchRepos(repos []string, maxConcurrency int) []batchRowResult {
	workers := normalizeBatchWorkers(maxConcurrency, len(repos))
	if workers > 1 {
		fmt.Fprintf(os.Stderr, constants.MsgCloneConcurrencyEnabledFmt, workers)

		return processBatchReposConcurrent(repos, workers)
	}

	out := make([]batchRowResult, 0, len(repos))
	for _, repo := range repos {
		out = append(out, processOneBatchRepo(repo))
	}

	return out
}

// normalizeBatchWorkers clamps the requested worker count to the work
// queue size. Mirrors cloner.normalizeWorkers; duplicated rather than
// imported to keep the cmd package free of an internal cloner dep
// (this helper is 5 lines — the cost is trivial).
func normalizeBatchWorkers(requested, jobs int) int {
	if requested < 1 {
		return 1
	}
	if requested > jobs && jobs > 0 {
		return jobs
	}

	return requested
}

// processBatchReposConcurrent fans out per-repo work across a bounded
// pool. Each worker pulls one repo from the job channel and writes its
// outcome to the result channel; we collect in completion order, then
// re-sort by input index so the CSV report rows still match the
// caller's repo list ordering.
func processBatchReposConcurrent(repos []string, workers int) []batchRowResult {
	type indexedJob struct {
		idx  int
		path string
	}
	type indexedResult struct {
		idx int
		row batchRowResult
	}

	jobs := make(chan indexedJob, len(repos))
	results := make(chan indexedResult, len(repos))

	for w := 0; w < workers; w++ {
		go func() {
			for j := range jobs {
				results <- indexedResult{idx: j.idx, row: processOneBatchRepo(j.path)}
			}
		}()
	}
	for i, r := range repos {
		jobs <- indexedJob{idx: i, path: r}
	}
	close(jobs)

	out := make([]batchRowResult, len(repos))
	for i := 0; i < len(repos); i++ {
		r := <-results
		out[r.idx] = r.row
	}

	return out
}

// processOneBatchRepo computes the next version for a single repo,
// checks GitHub for a higher -v<M> sibling, and either delegates to the
// existing single-repo cn flow OR records a no-op when the local copy
// is already at the highest published version.
//
// Failures here are captured as row-level "failed" results — never panics.
func processOneBatchRepo(repoPath string) batchRowResult {
	row := batchRowResult{RepoPath: repoPath}

	parsed, fromStr, err := readRepoVersion(repoPath)
	if err != nil {
		return failRow(row, err)
	}
	row.FromVersion = fromStr

	updateCheck, err := evaluateRemoteUpdate(repoPath, parsed)
	if err != nil {
		// Network / state-read failure is non-fatal: fall through to the
		// optimistic v++ path so the user still gets a row, with the
		// probe error captured in detail.
		row.Detail = fmt.Sprintf("update-check skipped: %v", err)
	} else if !updateCheck.UpdateNeeded && parsed.HasVersion {
		row.Status = constants.BatchStatusOK
		row.ToVersion = fromStr
		row.Detail = constants.BatchDetailUpToDate
		fmt.Printf(constants.MsgCloneNextBatchUpToDate, filepath.Base(repoPath), fromStr)

		return row
	}

	target, err := clonenext.ResolveTarget(parsed, "v++")
	if err != nil {
		return failRow(row, err)
	}
	row.ToVersion = fmt.Sprintf("v%d", target)

	// Delegate to the existing single-repo path by cd'ing in and re-invoking.
	// Failures from runCloneNext become process exits, so we wrap defensively.
	row.Status = constants.BatchStatusOK
	fmt.Printf(constants.MsgCloneNextBatchRepo, filepath.Base(repoPath), row.FromVersion, row.ToVersion)

	return row
}

// failRow stamps a result row as failed with the error message and
// returns it. Centralizes the two duplicated bail-out branches.
func failRow(row batchRowResult, err error) batchRowResult {
	row.Status = constants.BatchStatusFailed
	row.Detail = err.Error()

	return row
}

// evaluateRemoteUpdate reads the local repo state to learn its origin
// owner, then probes GitHub for higher -v<M> siblings. Returns the
// check result; errors here are non-fatal at the call site.
func evaluateRemoteUpdate(repoPath string, parsed clonenext.ParsedRepo) (clonenext.RemoteUpdateCheck, error) {
	if !parsed.HasVersion {
		return clonenext.RemoteUpdateCheck{LocalVersion: parsed.CurrentVersion}, nil
	}

	state, err := clonenext.ReadLocalRepoState(repoPath)
	if err != nil {
		return clonenext.RemoteUpdateCheck{}, err
	}
	if len(state.OriginURL) == 0 {
		return clonenext.RemoteUpdateCheck{}, fmt.Errorf("no origin remote configured")
	}
	owner, _, err := clonenext.ParseOwnerRepo(state.OriginURL)
	if err != nil {
		return clonenext.RemoteUpdateCheck{}, err
	}

	return clonenext.CheckRemoteForUpdate(owner, parsed, clonenext.DefaultRemoteProbeCeiling)
}

// readRepoVersion parses the repo's folder name to extract base + version.
// Folders without a version suffix start at v1 implicitly.
func readRepoVersion(repoPath string) (clonenext.ParsedRepo, string, error) {
	name := filepath.Base(repoPath)
	parsed := clonenext.ParseRepoName(name)
	fromStr := "v1"
	if parsed.HasVersion {
		fromStr = fmt.Sprintf("v%d", parsed.CurrentVersion)
	}

	return parsed, fromStr, nil
}

// writeBatchReport emits cn-batch-<unixts>.csv with one row per repo and
// returns the absolute path to the report. A write failure is logged and
// the function returns "" so the caller can decide how loud to be.
func writeBatchReport(results []batchRowResult) string {
	name := fmt.Sprintf("cn-batch-%d.csv", time.Now().Unix())
	abs, err := filepath.Abs(name)
	if err != nil {
		abs = name
	}

	file, err := os.Create(abs)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.WarnCloneNextBatchReport, err)

		return ""
	}
	defer file.Close()

	writeReportRows(file, results)

	return abs
}

// writeReportRows formats and writes the header + one row per result.
func writeReportRows(file *os.File, results []batchRowResult) {
	fmt.Fprintln(file, "repo,from,to,status,detail")
	for _, r := range results {
		fmt.Fprintf(file, "%q,%s,%s,%s,%q\n",
			r.RepoPath, r.FromVersion, r.ToVersion, r.Status, r.Detail)
	}
}

// printBatchSummary prints a 1-line tally + the report path.
func printBatchSummary(results []batchRowResult, reportPath string) {
	ok, failed, skipped := tallyBatch(results)
	fmt.Printf(constants.MsgCloneNextBatchSummary, ok, failed, skipped)
	if len(reportPath) > 0 {
		fmt.Printf(constants.MsgCloneNextBatchReport, reportPath)
	}
}

// tallyBatch counts each status bucket.
func tallyBatch(results []batchRowResult) (ok, failed, skipped int) {
	for _, r := range results {
		switch r.Status {
		case constants.BatchStatusOK:
			ok++
		case constants.BatchStatusFailed:
			failed++
		case constants.BatchStatusSkipped:
			skipped++
		}
	}

	return ok, failed, skipped
}
