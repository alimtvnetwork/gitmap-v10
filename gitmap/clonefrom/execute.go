package clonefrom

// Executor: walk Plan.Rows sequentially, shell out to `git clone`
// for each, and accumulate per-row Results. Sequential by design
// — parallel fan-out is a follow-up (see _TODO in
// .lovable/question-and-ambiguity/03-clone-from-scope.md). Adding
// it later is a one-function change because Result has no shared
// state between rows.
//
// Skip rule: if the resolved dest already exists AND is a non-
// empty directory → mark as `skipped` without invoking git. Makes
// re-running the same plan idempotent (common pattern: user fixes
// a typo in row 4, re-runs, doesn't want rows 1-3 re-cloned).
//
// We deliberately do NOT try to detect "dest exists AND points at
// the same URL" — that requires `git remote get-url` which (a)
// adds latency to the skip check and (b) would make the rule
// behave differently on partially-cloned dests. Conservative
// "non-empty dir = skip" is easier to reason about.

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// Result is one row's outcome. Status is one of "ok" | "skipped"
// | "failed". Detail is human-readable context: for "ok" the empty
// string; for "skipped" the reason ("dest exists"); for "failed"
// the trimmed git stderr (capped at GitErrorTrimLimit chars to
// keep the summary table readable).
type Result struct {
	Row      Row
	Dest     string // resolved (after DeriveDest fallback)
	Status   string
	Detail   string
	Duration time.Duration
}

// Execute runs every row sequentially and writes per-row progress
// lines to progress (typically os.Stderr; pass io.Discard to
// silence). Returns the full result slice — callers render the
// summary AFTER Execute returns so a write failure on progress
// doesn't truncate the result set.
//
// Working directory: each clone runs in `cwd`. Empty cwd → use
// the current process cwd at call time.
func Execute(plan Plan, cwd string, progress io.Writer) []Result {
	if len(cwd) == 0 {
		if wd, err := os.Getwd(); err == nil {
			cwd = wd
		}
	}
	out := make([]Result, 0, len(plan.Rows))
	for i, r := range plan.Rows {
		res := executeRow(r, cwd)
		out = append(out, res)
		writeProgress(progress, i+1, len(plan.Rows), res)
	}

	return out
}

// executeRow handles one row's lifecycle: resolve dest, check
// skip rule, build git args, run, time, return.
func executeRow(r Row, cwd string) Result {
	start := time.Now()
	dest := r.Dest
	if len(dest) == 0 {
		dest = DeriveDest(r.URL)
	}
	absDest := dest
	if !filepath.IsAbs(absDest) {
		absDest = filepath.Join(cwd, dest)
	}
	if shouldSkip(absDest) {
		return Result{Row: r, Dest: dest, Status: constants.CloneFromStatusSkipped,
			Detail: constants.MsgCloneFromDestExists, Duration: time.Since(start)}
	}
	detail, ok := runGitClone(r, dest, cwd)
	status := constants.CloneFromStatusOK
	if !ok {
		status = constants.CloneFromStatusFailed
	}

	return Result{Row: r, Dest: dest, Status: status, Detail: detail,
		Duration: time.Since(start)}
}

// shouldSkip returns true when the dest is a non-empty directory.
// Errors reading the dir (permission denied) → false (let git try
// and fail with a clearer message than we could craft).
func shouldSkip(absDest string) bool {
	info, err := os.Stat(absDest)
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(absDest)
	if err != nil {
		return false
	}

	return len(entries) > 0
}

// runGitClone shells out to `git clone` with the row's options.
// Returns (detail, ok). On success detail is empty. On failure
// detail is a single-line summary of the trimmed stderr.
func runGitClone(r Row, dest, cwd string) (string, bool) {
	args := buildGitArgs(r, dest)
	cmd := exec.Command(constants.GitBin, args...)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	if err != nil {
		return trimGitError(string(out), err), false
	}

	return "", true
}

// buildGitArgs translates a Row + resolved dest into the git
// clone argument vector. Order matters for git's flag parser
// (`--branch` and `--depth` must precede positionals).
func buildGitArgs(r Row, dest string) []string {
	args := []string{constants.GitClone}
	if len(r.Branch) > 0 {
		args = append(args, constants.GitBranchFlag, r.Branch)
	}
	if r.Depth > 0 {
		args = append(args, fmt.Sprintf(constants.CloneFromDepthFlagFmt, r.Depth))
	}
	args = append(args, r.URL, dest)

	return args
}

// trimGitError collapses multi-line git stderr to a single line
// with a length cap so the dry-run-then-execute summary table
// stays scannable. Full stderr is in the user's terminal scrollback
// already if they want it.
func trimGitError(stderr string, err error) string {
	last := stderr
	if i := strings.LastIndex(strings.TrimSpace(stderr), "\n"); i >= 0 {
		last = strings.TrimSpace(stderr)[i+1:]
	}
	last = strings.TrimSpace(last)
	if len(last) == 0 {
		last = err.Error()
	}
	if len(last) > constants.CloneFromErrTrimLimit {
		last = last[:constants.CloneFromErrTrimLimit] + "..."
	}

	return last
}

// writeProgress emits one line per finished row. Ignored on
// io.Writer error (caller may have passed io.Discard or a closed
// stderr — neither should abort the batch).
func writeProgress(w io.Writer, n, total int, res Result) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, "  [%d/%d] %-7s %s\n", n, total, res.Status, res.Row.URL)
}
