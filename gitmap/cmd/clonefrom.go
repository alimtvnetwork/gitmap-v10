package cmd

// CLI entry point for `gitmap clone-from <file>`. Parses flags,
// loads the plan, renders dry-run by default, executes only when
// --execute is passed, then prints the summary.
//
// Exit codes (mirrors gitmap conventions):
//
//   0 — dry-run completed OR every row was ok/skipped on execute
//   1 — file open / parse error, OR any row failed on execute
//   2 — bad CLI usage (missing <file> argument)
//
// The split between exit-1 (something failed during the operation)
// and exit-2 (caller error) lets shell scripts distinguish "you
// invoked me wrong" from "I tried but git rejected one of the
// URLs" — the second is recoverable by editing the input file,
// the first needs a different command.

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/clonefrom"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// cloneFromFlags holds parsed CLI inputs. Grouped in a struct
// (rather than passed individually) so future additions don't
// churn every helper signature.
type cloneFromFlags struct {
	file     string
	execute  bool
	quiet    bool
	noReport bool
	// output selects the per-row presentation: "" / "default" keep
	// the legacy 4-line block; "terminal" renders the standardized
	// branch/from/to/command block shared with scan, clone-next, probe.
	output string
	// verifyCmdFaithful enables the dry-run argv-vs-displayed checker.
	verifyCmdFaithful bool
}

// runCloneFrom is the dispatcher entry. checkHelp handles `--help`
// per the project's help-system convention before any flag parsing
// so unparseable flags don't suppress the help text.
func runCloneFrom(args []string) {
	checkHelp("clone-from", args)
	cfg := parseCloneFromFlags(args)
	setCmdFaithfulVerify(cfg.verifyCmdFaithful)
	plan, err := clonefrom.ParseFile(cfg.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !cfg.execute {
		runCloneFromDry(plan, cfg)

		return
	}
	runCloneFromExecute(plan, cfg)
}

// parseCloneFromFlags wires flags + extracts the positional file
// argument. Exits 2 with a clear message when <file> is missing —
// failing fast here is friendlier than parsing later and reporting
// "open : no such file or directory".
func parseCloneFromFlags(args []string) cloneFromFlags {
	var cfg cloneFromFlags
	fs := flag.NewFlagSet("clone-from", flag.ExitOnError)
	fs.BoolVar(&cfg.execute, constants.FlagCloneFromExecute, false,
		constants.FlagDescCloneFromExecute)
	fs.BoolVar(&cfg.quiet, constants.FlagCloneFromQuiet, false,
		constants.FlagDescCloneFromQuiet)
	fs.BoolVar(&cfg.noReport, constants.FlagCloneFromNoReport, false,
		constants.FlagDescCloneFromNoReport)
	fs.StringVar(&cfg.output, constants.FlagCloneFromOutput, "",
		constants.FlagDescCloneFromOutput)
	fs.BoolVar(&cfg.verifyCmdFaithful, constants.FlagCloneVerifyCmdFaithful,
		false, constants.FlagDescCloneVerifyCmdFaithful)
	reordered := reorderFlagsBeforeArgs(args)
	fs.Parse(reordered)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, constants.MsgCloneFromMissingArg)
		os.Exit(2)
	}
	cfg.file = fs.Arg(0)

	return cfg
}

// runCloneFromDry renders the dry-run preview and exits with the
// dry-run conventional code (0 = "I would do these things"). No
// side effects — by design, a dry-run never touches the network
// or the filesystem outside of READING the input file.
func runCloneFromDry(plan clonefrom.Plan, cfg cloneFromFlags) {
	render := pickCloneFromRenderer(cfg.output)
	if err := render(os.Stdout, plan); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// pickCloneFromRenderer dispatches between the legacy and the
// standardized terminal renderer based on --output. Anything other
// than "terminal" (including the empty default) keeps the legacy
// output so existing scripts and doc snippets stay byte-identical.
func pickCloneFromRenderer(output string) func(io.Writer, clonefrom.Plan) error {
	if output == constants.OutputTerminal {
		return clonefrom.RenderTerminal
	}

	return clonefrom.Render
}

// runCloneFromExecute is the side-effecting branch. Picks the
// progress writer based on --quiet, executes the plan, writes the
// CSV report (unless --no-report), prints the summary, then
// translates the result tally to an exit code.
func runCloneFromExecute(plan clonefrom.Plan, cfg cloneFromFlags) {
	progress := io.Writer(os.Stderr)
	if cfg.quiet {
		progress = io.Discard
	}
	// `--output terminal`: stream one standardized RepoTermBlock per
	// row via ExecuteWithHooks's BeforeRow callback — printed
	// IMMEDIATELY before that row's `git clone` shells out. This
	// interleaves per-repo previews with clone progress instead of
	// dumping every block upfront, matching the URL-driven `clone
	// <urls...>` behavior. Dry-run still uses RenderTerminal upfront
	// (no execution to interleave with). A nil hook keeps the legacy
	// path byte-identical for callers that didn't opt in.
	var hook clonefrom.BeforeRowHook
	if cfg.output == constants.OutputTerminal {
		hook = printCloneFromTermBlockRow
	}
	results := clonefrom.ExecuteWithHooks(plan, "", progress, hook)
	reportPath := ""
	if !cfg.noReport {
		if p, err := clonefrom.WriteReport(results); err == nil {
			reportPath = p
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	if err := clonefrom.RenderSummary(os.Stdout, results, reportPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(cloneFromExitCode(results))
}

// cloneFromExitCode returns 1 if any row failed, else 0. Skipped
// rows are NOT failures — re-running an idempotent plan with all
// destinations already in place is a successful no-op.
func cloneFromExitCode(results []clonefrom.Result) int {
	for _, r := range results {
		if r.Status == constants.CloneFromStatusFailed {

			return 1
		}
	}

	return 0
}
