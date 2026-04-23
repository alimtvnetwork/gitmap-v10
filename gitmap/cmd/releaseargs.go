package cmd

import "strings"

// reorderFlagsBeforeArgs moves flag-like arguments (starting with "-")
// before positional arguments. Go's flag package stops parsing at the
// first non-flag argument, so "gitmap release v2.55 -y" would silently
// ignore -y. This reorders to "-y v2.55" so all flags are parsed.
//
// Flags that take a value (e.g. --bump patch, -N "note") are kept
// together with their value argument.
func reorderFlagsBeforeArgs(args []string) []string {
	var flags []string
	var positional []string

	// Known flags that consume the next argument as a value.
	valueFlags := map[string]bool{
		// release / commit-flow flags
		"--assets": true, "--commit": true, "--branch": true,
		"--bump": true, "--notes": true, "--targets": true,
		"--bundle": true, "--zip-group": true,
		"-N": true, "-Z": true,
		// self-install / self-uninstall value-taking flags
		"--dir": true, "--version": true,
		"--profile": true, "--shell-mode": true,
		// clone-next value-taking flags
		"--csv": true, "--ssh-key": true, "-K": true,
		"--target-dir": true,
		// commit-transfer value-taking flags (spec 106 §8).
		// Without these, "--drop ^WIP --no-provenance" would have
		// --drop swallow "--no-provenance" as its regex value and
		// the negation toggle would silently never fire.
		"--strip": true, "--drop": true,
		"--limit": true, "--since": true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			// If this flag takes a value, grab the next arg too.
			if valueFlags[arg] && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, arg)
		}
	}

	return append(flags, positional...)
}
