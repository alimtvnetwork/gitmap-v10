package cmd

import (
	"errors"
	"flag"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// replaceOpts holds parsed flags for `gitmap replace`. Kept tiny so
// every handler can pass it by value.
type replaceOpts struct {
	yes    bool
	dryRun bool
	quiet  bool
	audit  bool
}

// parseReplaceFlags consumes flag tokens (in any position) and returns
// the remaining positional arguments. We hand-roll the audit detection
// because `--audit` is documented as a subcommand, not a typed flag,
// and we want the same token to work whether the user writes it before
// or after positional args.
func parseReplaceFlags(args []string) (replaceOpts, []string, error) {
	opts, rest := stripAuditToken(args)

	fs := flag.NewFlagSet(constants.CmdReplace, flag.ContinueOnError)
	yes := fs.Bool(constants.ReplaceFlagYes, false, "Skip the confirmation prompt")
	yesShort := fs.Bool(constants.ReplaceFlagYesS, false, "Alias for --yes")
	dry := fs.Bool(constants.ReplaceFlagDryRun, false, constants.FlagDescDryRun)
	quiet := fs.Bool(constants.ReplaceFlagQuiet, false, "Suppress per-file diff lines")
	quietShort := fs.Bool(constants.ReplaceFlagQuietS, false, "Alias for --quiet")

	flags, positional := splitReplaceFlagsAndArgs(rest)
	if err := fs.Parse(flags); err != nil {
		return opts, nil, errors.New("replace: " + err.Error())
	}

	opts.yes = *yes || *yesShort
	opts.dryRun = *dry
	opts.quiet = *quiet || *quietShort

	return opts, positional, nil
}

// stripAuditToken pulls --audit out of args before flag.Parse sees it.
func stripAuditToken(args []string) (replaceOpts, []string) {
	out := make([]string, 0, len(args))
	var opts replaceOpts
	for _, a := range args {
		if a == constants.ReplaceSubcmdAudit {
			opts.audit = true
			continue
		}
		out = append(out, a)
	}
	return opts, out
}

// splitReplaceFlagsAndArgs partitions args so flag.Parse only sees
// flag-like tokens. Anything else (including `-N` and `all`) is a
// positional. We treat tokens beginning with `--` or `-X`/`-XX`
// (letters) as flags.
func splitReplaceFlagsAndArgs(args []string) (flags, positional []string) {
	for _, a := range args {
		if isReplaceFlag(a) {
			flags = append(flags, a)
		} else {
			positional = append(positional, a)
		}
	}
	return
}

// isReplaceFlag returns true for `--xxx`, `-y`, `-q`. It deliberately
// returns false for `-N` digit forms so they remain positional.
func isReplaceFlag(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	if s[1] == '-' {
		return true
	}
	// single-dash: a flag iff the next char is a letter.
	c := s[1]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
