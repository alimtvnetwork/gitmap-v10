package cmd

// CLI entry point for `gitmap clone-now <file>`. Reads scan output
// (JSON / CSV / text) and re-runs `git clone` for each entry using
// the recorded folder structure and the user-selected URL mode.
//
// Exit codes (mirrors clone-from for consistency):
//
//   0 -- dry-run completed OR every row was ok/skipped on execute
//   1 -- file open / parse error, OR any row failed on execute
//   2 -- bad CLI usage (missing <file> argument or invalid flag value)
//
// The split between exit-1 and exit-2 lets shell scripts distinguish
// "you invoked me wrong" from "I tried but git rejected one of the
// URLs" -- the first is a coding error, the second is recoverable
// by editing the input file or fixing network/auth.

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/clonenow"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// cloneNowFlags holds parsed CLI inputs. Grouped in a struct so
// future additions don't churn every helper signature.
type cloneNowFlags struct {
	file     string
	execute  bool
	quiet    bool
	mode     string
	format   string
	cwd      string
	onExists string
	// output: "" (legacy) or "terminal" (standardized RepoTermBlock
	// streamed before each row's clone). Mirrors clone-from /
	// clone-next so all clone commands share one flag shape.
	output string
}

// runCloneNow is the dispatcher entry. checkHelp handles `--help`
// per the project help-system convention before any flag parsing
// so unparseable flags don't suppress the help text.
func runCloneNow(args []string) {
	checkHelp("clone-now", args)
	cfg := parseCloneNowFlags(args)
	plan, err := clonenow.ParseFile(cfg.file, cfg.format, cfg.mode, cfg.onExists)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !cfg.execute {
		runCloneNowDry(plan, cfg)

		return
	}
	runCloneNowExecute(plan, cfg)
}

// parseCloneNowFlags wires flags + extracts the positional file
// argument. Validates --mode and --format up-front so a typo exits
// 2 with a clear message instead of cascading into a confusing
// parse-time error later.
func parseCloneNowFlags(args []string) cloneNowFlags {
	var cfg cloneNowFlags
	fs := flag.NewFlagSet("clone-now", flag.ExitOnError)
	fs.BoolVar(&cfg.execute, constants.FlagCloneNowExecute, false,
		constants.FlagDescCloneNowExecute)
	fs.BoolVar(&cfg.quiet, constants.FlagCloneNowQuiet, false,
		constants.FlagDescCloneNowQuiet)
	fs.StringVar(&cfg.mode, constants.FlagCloneNowMode,
		constants.CloneNowModeHTTPS, constants.FlagDescCloneNowMode)
	fs.StringVar(&cfg.format, constants.FlagCloneNowFormat, "",
		constants.FlagDescCloneNowFormat)
	fs.StringVar(&cfg.cwd, constants.FlagCloneNowCwd, "",
		constants.FlagDescCloneNowCwd)
	fs.StringVar(&cfg.onExists, constants.FlagCloneNowOnExists,
		constants.CloneNowOnExistsSkip, constants.FlagDescCloneNowOnExists)
	fs.StringVar(&cfg.output, constants.FlagCloneTermOutput, "",
		constants.FlagDescCloneTermOutput)
	reordered := reorderFlagsBeforeArgs(args)
	fs.Parse(reordered)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, constants.MsgCloneNowMissingArg)
		os.Exit(2)
	}
	cfg.file = fs.Arg(0)
	validateCloneNowFlags(cfg)

	return cfg
}

// validateCloneNowFlags hard-fails (exit 2) on invalid --mode or
// --format values. Done after flag.Parse so the user sees one error
// at a time instead of a wall of stacked usage text.
func validateCloneNowFlags(cfg cloneNowFlags) {
	if cfg.mode != constants.CloneNowModeHTTPS && cfg.mode != constants.CloneNowModeSSH {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNowBadMode+"\n", cfg.mode)
		os.Exit(2)
	}
	switch cfg.format {
	case "", constants.CloneNowFormatJSON, constants.CloneNowFormatCSV, constants.CloneNowFormatText:
	default:
		fmt.Fprintf(os.Stderr, constants.ErrCloneNowBadFormat+"\n", cfg.format)
		os.Exit(2)
	}
	switch cfg.onExists {
	case constants.CloneNowOnExistsSkip,
		constants.CloneNowOnExistsUpdate,
		constants.CloneNowOnExistsForce:
		return
	}
	fmt.Fprintf(os.Stderr, constants.ErrCloneNowBadOnExists+"\n", cfg.onExists)
	os.Exit(2)
}

// runCloneNowDry renders the dry-run preview. No side effects --
// dry-run never touches the network or filesystem outside reading
// the input file.
func runCloneNowDry(plan clonenow.Plan, cfg cloneNowFlags) {
	if cfg.output == constants.OutputTerminal {
		printCloneNowTermBlocks(plan)

		return
	}
	if err := clonenow.Render(os.Stdout, plan); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runCloneNowExecute is the side-effecting branch. Picks the
// progress writer based on --quiet, executes the plan, prints the
// summary, then translates the result tally to an exit code.
func runCloneNowExecute(plan clonenow.Plan, cfg cloneNowFlags) {
	progress := io.Writer(os.Stderr)
	if cfg.quiet {
		progress = io.Discard
	}
	// `--output terminal`: stream one standardized block per row,
	// printed by ExecuteWithHooks's BeforeRow callback IMMEDIATELY
	// before that row's `git clone` starts. This interleaves the
	// per-repo preview with live clone progress instead of dumping
	// every block upfront — matches the URL-driven `clone <urls...>`
	// behavior. A nil hook keeps the legacy code path identical for
	// callers that didn't opt in.
	var hook clonenow.BeforeRowHook
	if cfg.output == constants.OutputTerminal {
		hook = printCloneNowTermBlockRow
	}
	results := clonenow.ExecuteWithHooks(plan, cfg.cwd, progress, hook)
	if err := clonenow.RenderSummary(os.Stdout, results); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(cloneNowExitCode(results))
}

// cloneNowExitCode returns 1 if any row failed, else 0. Skipped
// rows are NOT failures -- re-running an idempotent plan with all
// destinations already in place is a successful no-op.
func cloneNowExitCode(results []clonenow.Result) int {
	for _, r := range results {
		if r.Status == constants.CloneNowStatusFailed {

			return 1
		}
	}

	return 0
}
