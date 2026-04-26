package cmd

// CLI runners for `gitmap startup-list` and `gitmap startup-remove`.
// Both commands work on Linux/Unix (XDG `.desktop` files) and macOS
// (LaunchAgent `.plist` files); on Windows they exit with a clear
// "unsupported OS" message rather than silently doing nothing
// (silent no-ops on a different OS would be a UX trap: the user
// would assume their startup entries are managed when they're not).
//
// Output contract (per user requirement):
//   - Clear: every outcome prints exactly one line summary.
//   - Safe no-op: missing files, third-party files, and empty
//     listings all exit 0 with a message — never an error.
//   - Scoped: only entries carrying the gitmap marker are touched.
//     The startup package enforces this; this layer just renders the
//     result. `--dry-run` runs the same classification but skips
//     the actual unlink.

import (
	"flag"
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/startup"
)

// runStartupList enumerates gitmap-managed XDG autostart entries and
// renders them in the chosen format. The default `table` format
// matches the legacy human-readable output; `json` and `csv` exist
// for piping into other tools (jq, spreadsheet imports, etc).
func runStartupList(args []string) {
	format, err := parseStartupListFlags(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
	entries, err := startup.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	dir, _ := startup.AutostartDir()
	if err := renderStartupList(format, dir, entries); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// parseStartupListFlags extracts the --format value and validates it
// against the accepted set. Unknown values fail fast with exit 2 so
// scripts can detect a typo immediately rather than getting silent
// fall-through to a default rendering.
func parseStartupListFlags(args []string) (string, error) {
	fs := flag.NewFlagSet("startup-list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	format := fs.String(
		constants.FlagStartupListFormat,
		constants.StartupListFormatTable,
		constants.FlagDescStartupListFormat,
	)
	if err := fs.Parse(args); err != nil {

		return "", err
	}
	switch *format {
	case constants.StartupListFormatTable, constants.OutputTerminal,
		constants.OutputJSON, constants.StartupListFormatJSONL,
		constants.OutputCSV:

		return *format, nil
	default:

		return "", fmt.Errorf(constants.ErrStartupListBadFormat, *format)
	}
}

// runStartupRemove deletes a single managed entry. After --dry-run
// is parsed off the args, exactly one positional name must remain;
// missing or extra positionals trigger the usage error. All four
// RemoveStatus outcomes map to a distinct user-visible message
// (with a `(dry-run)` mirror set when --dry-run is active) so the
// CLI is unambiguous about what happened — or what would happen.
func runStartupRemove(args []string) {
	name, dryRun, err := parseStartupRemoveFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, constants.ErrStartupRemoveUsage)
		os.Exit(2)
	}
	res, err := startup.RemoveWithOptions(name, startup.RemoveOptions{DryRun: dryRun})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	printRemoveResult(name, res)
}

// parseStartupRemoveFlags pulls --dry-run off the args and returns
// the remaining single positional name. Returns an error when the
// positional count is wrong so the caller exits 2 with the usage
// message — matching the pre-flag behavior for malformed invocations.
func parseStartupRemoveFlags(args []string) (string, bool, error) {
	fs := flag.NewFlagSet(constants.CmdStartupRemove, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dryRun := fs.Bool(
		constants.FlagStartupRemoveDryRun, false,
		constants.FlagDescStartupRemoveDryRun,
	)
	if err := fs.Parse(args); err != nil {

		return "", false, err
	}
	rest := fs.Args()
	if len(rest) != 1 {

		return "", false, fmt.Errorf("expected 1 positional name, got %d", len(rest))
	}

	return rest[0], *dryRun, nil
}

// printRemoveResult routes one of four messages depending on the
// status. Dry-run results pick the `(dry-run)` mirror message so a
// preview is visually distinct from a real action — important when
// users pipe both into the same log. Each branch is single-line so
// log-scrapers can classify outcomes without parsing multi-line
// output.
func printRemoveResult(name string, res startup.RemoveResult) {
	if res.DryRun {
		printRemoveResultDryRun(name, res)

		return
	}
	switch res.Status {
	case startup.RemoveDeleted:
		fmt.Printf(constants.MsgStartupRemoveOK, res.Path)
	case startup.RemoveNoOp:
		fmt.Printf(constants.MsgStartupRemoveNoOp, name)
	case startup.RemoveRefused:
		fmt.Printf(constants.MsgStartupRemoveNotOurs, res.Path)
	case startup.RemoveBadName:
		fmt.Printf(constants.MsgStartupRemoveBadName, name)
	}
}

// printRemoveResultDryRun is the `--dry-run` parallel of
// printRemoveResult. Kept as its own function (rather than an
// if/else inside the parent switch) so the live and preview message
// tables are easy to diff line-for-line during code review.
func printRemoveResultDryRun(name string, res startup.RemoveResult) {
	switch res.Status {
	case startup.RemoveDeleted:
		fmt.Printf(constants.MsgStartupRemoveDryOK, res.Path)
	case startup.RemoveNoOp:
		fmt.Printf(constants.MsgStartupRemoveDryNoOp, name)
	case startup.RemoveRefused:
		fmt.Printf(constants.MsgStartupRemoveDryNotOurs, res.Path)
	case startup.RemoveBadName:
		fmt.Printf(constants.MsgStartupRemoveDryBadName, name)
	}
}
