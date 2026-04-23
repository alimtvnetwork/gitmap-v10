package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/probe"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/store"
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

// runProbe dispatches `gitmap probe [<repo-path>|--all] [--json]`. Phase 2.5
// will replace the sequential loop with a parallel worker pool.
func runProbe(args []string) {
	checkHelp("probe", args)

	jsonOut, positional := splitProbeArgs(args)

	db := openSfDB()
	defer db.Close()

	targets, err := resolveProbeTargets(db, positional)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if len(targets) == 0 {
		emitProbeEmpty(jsonOut)
		return
	}

	probeAndReport(db, targets, jsonOut)
}

// splitProbeArgs separates --json from positional args. Order-agnostic.
func splitProbeArgs(args []string) (bool, []string) {
	jsonOut := false
	rest := make([]string, 0, len(args))
	for _, a := range args {
		if a == constants.ProbeFlagJSON {
			jsonOut = true
			continue
		}
		rest = append(rest, a)
	}

	return jsonOut, rest
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

// probeAndReport executes RunOne for every target, persists results, and
// emits either the human summary or a JSON array depending on jsonOut.
func probeAndReport(db *store.DB, targets []model.ScanRecord, jsonOut bool) {
	if !jsonOut {
		fmt.Printf(constants.MsgProbeStartFmt, len(targets))
	}

	entries, available, unchanged, failed := runProbeLoop(db, targets, jsonOut)

	if jsonOut {
		emitProbeJSON(entries)
		return
	}
	fmt.Printf(constants.MsgProbeDoneFmt, available, unchanged, failed)
}

// runProbeLoop executes the probe per target and tallies counters. When
// jsonOut is true the per-line summaries are suppressed and entries are
// collected for a single JSON dump at the end.
func runProbeLoop(db *store.DB, targets []model.ScanRecord, jsonOut bool) ([]probeJSONEntry, int, int, int) {
	entries := make([]probeJSONEntry, 0, len(targets))
	available, unchanged, failed := 0, 0, 0

	for _, repo := range targets {
		url := pickProbeURL(repo)
		if url == "" {
			result := probe.Result{Method: constants.ProbeMethodNone, Error: fmt.Sprintf(constants.ErrProbeMissingURL, repo.Slug)}
			recordProbeResult(db, repo, result)
			entries = append(entries, makeProbeEntry(repo, result))
			if !jsonOut {
				fmt.Fprintf(os.Stderr, "%s\n", result.Error)
			}
			failed++
			continue
		}

		result := probe.RunOne(url)
		recordProbeResult(db, repo, result)
		entries = append(entries, makeProbeEntry(repo, result))
		available, unchanged, failed = tallyProbe(repo, result, available, unchanged, failed, jsonOut)
	}

	return entries, available, unchanged, failed
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
// per-repo summary line.
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
