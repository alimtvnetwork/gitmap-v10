package cmd

import (
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// commitTransferSpec describes one of the three commit-transfer
// commands. It mirrors the mergeSpec pattern in dispatchmovemerge.go so
// the eventual implementation can share dispatch + arg-validation code
// with the merge-* family.
type commitTransferSpec struct {
	Name      string // e.g. constants.CmdCommitLeft
	LogPrefix string // e.g. constants.LogPrefixCommitLeft
}

// runCommitTransfer is the single entry point for commit-left,
// commit-right, and commit-both.
//
// Status (v3.74.0): scaffold only. Validates arg count, prints the
// would-be plan header, then exits 2 with a clear pointer to the spec.
// The full replay engine ships in a follow-up session per spec §18
// (Phase 1 — commit-right; Phase 2 — commit-left; Phase 3 — commit-both).
func runCommitTransfer(spec commitTransferSpec, args []string) {
	checkHelp(spec.Name, args)
	positional := filterNonFlagArgs(args)
	if len(positional) != 2 {
		fmt.Fprintf(os.Stderr, constants.ErrCTArgCountFmt, spec.Name, len(positional))
		fmt.Fprintf(os.Stderr, constants.MsgCTUsageFmt, spec.Name, spec.Name)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, constants.ErrCTNotImplementedFmt, spec.Name)
	os.Exit(2)
}

// filterNonFlagArgs returns positional args (no --flag tokens or their
// values). This is intentionally simple — the real implementation will
// use a proper flag.FlagSet once the engine lands.
func filterNonFlagArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if isFlagToken(a) {
			continue
		}
		out = append(out, a)
	}

	return out
}

// commitTransferSpecFor maps a command name or alias to its spec.
// Returns ok=false when the command does not belong to this family.
func commitTransferSpecFor(command string) (commitTransferSpec, bool) {
	switch command {
	case constants.CmdCommitLeft, constants.CmdCommitLeftA:
		return commitTransferSpec{
			Name: constants.CmdCommitLeft, LogPrefix: constants.LogPrefixCommitLeft,
		}, true
	case constants.CmdCommitRight, constants.CmdCommitRightA:
		return commitTransferSpec{
			Name: constants.CmdCommitRight, LogPrefix: constants.LogPrefixCommitRight,
		}, true
	case constants.CmdCommitBoth, constants.CmdCommitBothA:
		return commitTransferSpec{
			Name: constants.CmdCommitBoth, LogPrefix: constants.LogPrefixCommitBoth,
		}, true
	}

	return commitTransferSpec{}, false
}
