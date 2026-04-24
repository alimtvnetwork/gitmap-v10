// Package cmd — extended `--debug-windows` output that prints the
// exact commands and filesystem operations the update-cleanup handoff
// will perform, so the user can audit and reproduce them.
//
// Two functions are exported within the package:
//
//   dumpDebugWindowsCommandPlan — renders the Phase 3 spawn command
//   line as a copy-pastable shell invocation (proper quoting), and
//   prints an explicit note that no `git` subprocess is launched.
//
//   dumpDebugWindowsCleanupPlan — enumerates the filepath.Glob
//   patterns the deployed binary will scan and the actual file
//   matches that will be passed to os.Remove / os.RemoveAll. Called
//   from runUpdateCleanup BEFORE any deletion happens so the output
//   reflects intent, not outcome.
//
// These complement the structured-event log written by
// updatehandofflog.go (which is post-hoc) with a pre-flight view.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// dumpDebugWindowsCommandPlan renders the exact spawn invocation as
// a single shell-friendly line: `"<deployed>" "<arg1>" "<arg2>" ...`.
// Always quotes for cross-platform copy-paste safety. Followed by an
// explicit note that the cleanup itself is pure Go syscalls — no
// `git` or other subprocess is launched.
func dumpDebugWindowsCommandPlan(deployed string, childArgs []string) {
	if !isDebugWindowsRequested() {
		return
	}
	full := append([]string{deployed}, childArgs...)
	fmt.Fprintf(os.Stderr, constants.MsgDebugWinCmdLine,
		renderShellCommand(full))
	fmt.Fprint(os.Stderr, constants.MsgDebugWinCmdNote)
}

// renderShellCommand quotes each token with double quotes (escaping
// any embedded `"`) and joins with spaces. Cross-platform safe for
// copy-paste — works in PowerShell, cmd.exe, bash, and zsh.
func renderShellCommand(tokens []string) string {
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		parts = append(parts, quoteShellToken(t))
	}

	return strings.Join(parts, " ")
}

// quoteShellToken wraps a token in double quotes, escaping inner `"`
// as `\"`. Tokens with no whitespace or special chars are still
// quoted for visual consistency in the dump.
func quoteShellToken(t string) string {
	escaped := strings.ReplaceAll(t, `"`, `\"`)

	return `"` + escaped + `"`
}

// dumpDebugWindowsCleanupPlan enumerates every filesystem operation
// the deployed binary will attempt during update-cleanup. Called
// from runUpdateCleanup AFTER loadUpdateCleanupContext but BEFORE
// the actual cleanupTempArtifacts/cleanupBackupArtifacts/etc. calls
// so the user sees the *plan*, not the outcome.
//
// Operations enumerated:
//   1. Temp-artifact globs (gitmap-update-* in TEMP and deploy dirs)
//   2. Backup-artifact globs (*.old next to deployed binary)
//   3. Clone-swap dir globs (*.gitmap-tmp-*)
//   4. Drive-root shim candidate (Windows only)
func dumpDebugWindowsCleanupPlan(ctx updateCleanupContext) {
	if !isDebugWindowsRequested() {
		return
	}
	fmt.Fprint(os.Stderr, constants.MsgDebugWinCleanHdr)
	dumpPlannedRemovals(ctx.tempPatterns, ctx.selfPath,
		constants.MsgDebugWinCleanMatch)
	dumpPlannedRemovals(ctx.backupPatterns, ctx.selfPath,
		constants.MsgDebugWinCleanMatch)
	dumpPlannedSwapDirs(ctx)
	dumpPlannedDriveRootShim(ctx)
	fmt.Fprint(os.Stderr, constants.MsgDebugWinCleanFooter)
}

// dumpPlannedRemovals prints each glob pattern and the matches that
// would be passed to os.Remove. Skips matches equal to selfPath
// because removeCleanupMatch will skip them too.
func dumpPlannedRemovals(patterns []string, selfPath, matchFmt string) {
	for _, pattern := range patterns {
		fmt.Fprintf(os.Stderr, constants.MsgDebugWinCleanGlob, pattern)
		matches, err := filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			fmt.Fprint(os.Stderr, constants.MsgDebugWinCleanEmpty)

			continue
		}
		printed := 0
		for _, m := range matches {
			if isActiveCleanupPath(normalizeCleanupPath(m), selfPath) {
				continue
			}
			fmt.Fprintf(os.Stderr, matchFmt, m)
			printed++
		}
		if printed == 0 {
			fmt.Fprint(os.Stderr, constants.MsgDebugWinCleanEmpty)
		}
	}
}

// dumpPlannedSwapDirs prints every *.gitmap-tmp-* dir the cleanup
// pass will pass to os.RemoveAll. Mirrors removeCloneSwapDirsIn so
// the dump cannot drift from the live behaviour.
func dumpPlannedSwapDirs(ctx updateCleanupContext) {
	dirs := uniqueParentDirs(ctx.tempPatterns, ctx.backupPatterns)
	for _, dir := range dirs {
		pattern := filepath.Join(dir, cloneSwapDirGlob)
		fmt.Fprintf(os.Stderr, constants.MsgDebugWinCleanGlob, pattern)
		matches, err := filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			fmt.Fprint(os.Stderr, constants.MsgDebugWinCleanEmpty)

			continue
		}
		for _, m := range matches {
			info, statErr := os.Stat(m)
			if statErr != nil || !info.IsDir() {
				continue
			}
			fmt.Fprintf(os.Stderr, constants.MsgDebugWinCleanSwap, m)
		}
	}
}

// dumpPlannedDriveRootShim reports the Windows-only drive-root shim
// candidate and whether cleanup would actually delete it. Mirrors
// the gating logic in cleanupDriveRootShim / isRemovableDriveRootShim.
func dumpPlannedDriveRootShim(ctx updateCleanupContext) {
	shim := resolveDriveRootShimPath(ctx.selfPath)
	if len(shim) == 0 {
		return
	}
	verdict := constants.MsgDebugWinCleanShimSkip
	if isRemovableDriveRootShim(shim, ctx.selfPath) {
		verdict = constants.MsgDebugWinCleanShimDel
	}
	fmt.Fprintf(os.Stderr, constants.MsgDebugWinCleanShim, shim, verdict)
}
