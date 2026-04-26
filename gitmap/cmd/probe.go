package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/model"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/probe"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/store"
)

// probeJSONEntry is a single repo-level result emitted under `--json`.
// Embeds the result + repo identity so a CI consumer can join on either.
type probeJSONEntry struct {
	RepoID         int64  `json:"repoId"`
	Slug           string `json:"slug"`
	AbsolutePath   string `json:"absolutePath"`
	NextVersionTag string `json:"nextVersionTag"`
	NextVersionNum int64  `json:"nextVersionNum"`
	Method         string `json:"method"`
	IsAvailable    bool   `json:"isAvailable"`
	Error          string `json:"error,omitempty"`
}

// probeOptions captures the parsed CLI flags for `gitmap probe`.
// Workers is already clamped into [1, ProbeMaxWorkers] by parseProbeArgs.
type probeOptions struct {
	jsonOut bool
	workers int
	rest    []string
}

// runProbe dispatches `gitmap probe [<repo-path>|--all] [--json] [--workers N]`.
// The probe pool is capped at constants.ProbeMaxWorkers (default 2) to stay
// under provider rate limits.
func runProbe(args []string) {
	checkHelp("probe", args)

	opts, err := parseProbeArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	db := openSfDB()
	defer db.Close()

	targets, err := resolveProbeTargets(db, opts.rest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if len(targets) == 0 {
		emitProbeEmpty(opts.jsonOut)
		return
	}

	probeAndReport(db, targets, opts)
}

// parseProbeArgs walks the arg list, peeling off recognised flags and
// returning everything else as positional args. Order-agnostic; supports
// both `--workers N` and `--workers=N`.
func parseProbeArgs(args []string) (probeOptions, error) {
	opts := probeOptions{workers: constants.ProbeDefaultWorkers, rest: make([]string, 0, len(args))}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == constants.ProbeFlagJSON:
			opts.jsonOut = true
		case a == constants.ProbeFlagWorkers:
			if i+1 >= len(args) {
				return opts, fmt.Errorf(constants.ErrProbeWorkersMissing)
			}
			n, err := parseWorkersValue(args[i+1])
			if err != nil {
				return opts, err
			}
			opts.workers = clampProbeWorkers(n)
			i++
		case len(a) > len(constants.ProbeFlagWorkers)+1 && a[:len(constants.ProbeFlagWorkers)+1] == constants.ProbeFlagWorkers+"=":
			n, err := parseWorkersValue(a[len(constants.ProbeFlagWorkers)+1:])
			if err != nil {
				return opts, err
			}
			opts.workers = clampProbeWorkers(n)
		default:
			opts.rest = append(opts.rest, a)
		}
	}

	return opts, nil
}

// parseWorkersValue validates the `--workers` argument as a positive int.
func parseWorkersValue(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 0, fmt.Errorf(constants.ErrProbeWorkersValue, s)
	}

	return n, nil
}

// clampProbeWorkers enforces the [1, ProbeMaxWorkers] cap, printing a
// notice to stderr when the user asked for more than we'll grant.
func clampProbeWorkers(n int) int {
	if n > constants.ProbeMaxWorkers {
		fmt.Fprintf(os.Stderr, constants.MsgProbeWorkersClamped, n, constants.ProbeMaxWorkers)
		return constants.ProbeMaxWorkers
	}

	return n
}

// emitProbeEmpty handles the "no targets" case in either output mode.
func emitProbeEmpty(jsonOut bool) {
	if jsonOut {
		fmt.Println("[]")
		return
	}
	fmt.Print(constants.MsgProbeNoTargets)
}

// resolveProbeTargets converts CLI args into a list of repos to probe.
func resolveProbeTargets(db *store.DB, args []string) ([]model.ScanRecord, error) {
	if len(args) == 0 || args[0] == constants.ProbeFlagAll {
		return db.ListRepos()
	}

	absPath, err := filepath.Abs(args[0])
	if err != nil {
		return nil, fmt.Errorf(constants.ErrSFAbsResolve, args[0], err)
	}

	matches, err := db.FindByPath(absPath)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf(constants.ErrProbeNoRepo, absPath)
	}

	return matches, nil
}

// probeAndReport executes RunOne for every target via a capped worker
// pool, persists results, and emits either the human summary or a JSON
// array depending on opts.jsonOut. Workers always >= 1.
func probeAndReport(db *store.DB, targets []model.ScanRecord, opts probeOptions) {
	if !opts.jsonOut {
		fmt.Printf(constants.MsgProbeStartFmt, len(targets))
	}

	entries, available, unchanged, failed := runProbePool(db, targets, opts)

	if opts.jsonOut {
		emitProbeJSON(entries)
		return
	}
	fmt.Printf(constants.MsgProbeDoneFmt, available, unchanged, failed)
}

// runProbePool fans the targets across opts.workers goroutines. Entries
// are slotted back into input order so the JSON output is deterministic
// regardless of completion order. Per-repo human progress lines print
// as workers finish (matches the cloner pattern); only the trailing
// summary depends on the totals.
func runProbePool(db *store.DB, targets []model.ScanRecord, opts probeOptions) ([]probeJSONEntry, int, int, int) {
	type job struct {
		idx  int
		repo model.ScanRecord
	}
	jobs := make(chan job, len(targets))
	entries := make([]probeJSONEntry, len(targets))
	var counterMu sync.Mutex
	available, unchanged, failed := 0, 0, 0

	var wg sync.WaitGroup
	for w := 0; w < opts.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				result := executeOneProbe(db, j.repo)
				entries[j.idx] = makeProbeEntry(j.repo, result)
				counterMu.Lock()
				available, unchanged, failed = tallyProbe(j.repo, result, available, unchanged, failed, opts.jsonOut)
				counterMu.Unlock()
			}
		}()
	}
	for i, repo := range targets {
		jobs <- job{idx: i, repo: repo}
	}
	close(jobs)
	wg.Wait()

	return entries, available, unchanged, failed
}

// executeOneProbe runs a single probe and persists it, mirroring the
// missing-URL handling that the sequential loop used. Lives outside
// the worker closure so it stays under the 15-line function limit.
func executeOneProbe(db *store.DB, repo model.ScanRecord) probe.Result {
	url := pickProbeURL(repo)
	if url == "" {
		result := probe.Result{Method: constants.ProbeMethodNone, Error: fmt.Sprintf(constants.ErrProbeMissingURL, repo.Slug)}
		recordProbeResult(db, repo, result)

		return result
	}
	result := probe.RunOne(url)
	recordProbeResult(db, repo, result)

	return result
}

// makeProbeEntry converts a probe.Result + repo into a JSON-friendly row.
func makeProbeEntry(repo model.ScanRecord, r probe.Result) probeJSONEntry {
	return probeJSONEntry{
		RepoID:         repo.ID,
		Slug:           repo.Slug,
		AbsolutePath:   repo.AbsolutePath,
		NextVersionTag: r.NextVersionTag,
		NextVersionNum: r.NextVersionNum,
		Method:         r.Method,
		IsAvailable:    r.IsAvailable,
		Error:          r.Error,
	}
}

// emitProbeJSON dumps the collected entries as indented JSON to stdout.
func emitProbeJSON(entries []probeJSONEntry) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// pickProbeURL prefers HTTPS (less auth friction in CI), falls back to SSH.
func pickProbeURL(r model.ScanRecord) string {
	if r.HTTPSUrl != "" {
		return r.HTTPSUrl
	}

	return r.SSHUrl
}

// recordProbeResult persists the probe row, logging-but-not-exiting on error.
func recordProbeResult(db *store.DB, repo model.ScanRecord, result probe.Result) {
	if err := db.RecordVersionProbe(result.AsModel(repo.ID)); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

// tallyProbe updates the running counters and (unless jsonOut) prints the
// per-repo summary line. Caller is responsible for serialising access to
// the counters; with the worker pool that's `counterMu` in runProbePool.
func tallyProbe(repo model.ScanRecord, r probe.Result, ok, none, fail int, jsonOut bool) (int, int, int) {
	if r.Error != "" {
		if !jsonOut {
			fmt.Printf(constants.MsgProbeFailFmt, repo.Slug, r.Error)
		}
		return ok, none, fail + 1
	}
	if r.IsAvailable {
		if !jsonOut {
			fmt.Printf(constants.MsgProbeOkFmt, repo.Slug, r.NextVersionTag, r.Method)
		}
		return ok + 1, none, fail
	}
	if !jsonOut {
		fmt.Printf(constants.MsgProbeNoneFmt, repo.Slug, r.Method)
	}

	return ok, none + 1, fail
}
