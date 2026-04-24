package cmd

import (
	"errors"
	"flag"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// replaceOpts holds parsed flags for `gitmap replace`. Kept tiny so
// every handler can pass it by value.
type replaceOpts struct {
	yes    bool
	dryRun bool
	quiet  bool
	audit  bool
	exts   []string // normalized: lowercase, leading dot. Empty = no filter.
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
	ext := fs.String(constants.ReplaceFlagExt, "", constants.FlagDescReplaceExt)

	flags, positional := splitReplaceFlagsAndArgs(rest)
	if err := fs.Parse(flags); err != nil {
		return opts, nil, errors.New("replace: " + err.Error())
	}

	opts.yes = *yes || *yesShort
	opts.dryRun = *dry
	opts.quiet = *quiet || *quietShort
	opts.exts = normalizeExtList(*ext)

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
// (letters) as flags. A flag's value (`--ext .go,.md`) sticks to its
// flag token via the previous-token lookback.
func splitReplaceFlagsAndArgs(args []string) (flags, positional []string) {
	expectValue := false
	for _, a := range args {
		if expectValue {
			flags = append(flags, a)
			expectValue = false
			continue
		}
		if isReplaceFlag(a) {
			flags = append(flags, a)
			expectValue = needsValue(a)
			continue
		}
		positional = append(positional, a)
	}
	return
}

// needsValue reports whether a flag token consumes the next arg as its
// value. Boolean replace flags do not; --ext does. We don't need to
// handle `--ext=.go,.md` because flag.Parse splits that itself.
func needsValue(token string) bool {
	if strings.Contains(token, "=") {
		return false
	}
	name := strings.TrimLeft(token, "-")
	return name == constants.ReplaceFlagExt
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
	c := s[1]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// normalizeExtList parses the --ext value into a deduplicated list of
// lowercase extensions with leading dots. Empty input returns nil so
// the walker can short-circuit the "no filter" case.
func normalizeExtList(raw string) []string {
	if raw == "" {
		return nil
	}
	seen := make(map[string]struct{}, 4)
	out := make([]string, 0, 4)
	for _, piece := range strings.Split(raw, constants.ReplaceExtSep) {
		ext := normalizeOneExt(piece)
		if ext == "" {
			continue
		}
		if _, dup := seen[ext]; dup {
			continue
		}
		seen[ext] = struct{}{}
		out = append(out, ext)
	}
	return out
}

// normalizeOneExt trims spaces, lowercases, and ensures the leading dot.
func normalizeOneExt(piece string) string {
	piece = strings.ToLower(strings.TrimSpace(piece))
	if piece == "" || piece == "." {
		return ""
	}
	if piece[0] != '.' {
		piece = "." + piece
	}
	return piece
}
