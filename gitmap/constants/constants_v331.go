// Package constants — constants_v331.go centralizes the new strings introduced
// in v3.31.0 for: cross-dir release (`r <repo> <ver>`), cross-dir clone-next
// (`cn <repo> <ver>`), `has-change` (`hc`), and the SSH existing-key-on-disk
// fix. Kept in a single file (rather than scattered across domain files) to
// keep the v3.31.0 surface area easy to audit and unwind.
package constants

// gitmap:cmd top-level
// New CLI commands.
const (
	CmdHasChange      = "has-change"
	CmdHasChangeAlias = "hc"
)

// has-change flag names + descriptions.
const (
	FlagHCMode      = "mode"
	FlagHCAll       = "all"
	FlagHCFetch     = "fetch"
	FlagDescHCMode  = "Dimension to check: dirty (default), ahead, or behind"
	FlagDescHCAll   = "Print all three dimensions as structured output"
	FlagDescHCFetch = "Run 'git fetch' before checking ahead/behind (default true)"
)

// has-change mode values + literals.
const (
	HCModeDirty  = "dirty"
	HCModeAhead  = "ahead"
	HCModeBehind = "behind"
	HCTrue       = "true"
	HCFalse      = "false"
)

// has-change messages + errors.
const (
	MsgHCAllFmt        = "dirty=%s ahead=%s behind=%s\n"
	MsgHCAllNoUpstream = "dirty=%s ahead=n/a behind=n/a (no upstream)\n"
	ErrHCUsage         = "Usage: gitmap has-change <repo> [--mode dirty|ahead|behind] [--all]"
	ErrHCBadMode       = "  ✗ Unknown mode %q. Use one of: dirty, ahead, behind.\n"
	WarnHCFetchFailed  = "  ⚠ git fetch failed in %s: %v (ahead/behind may be stale)\n"
)

// Cross-dir release (`r <repo> <ver>`) messages + errors.
const (
	MsgRRStartingFmt    = "  → Releasing %s at %s (version %s)...\n"
	MsgRRFetchingFmt    = "  📡 Fetching remote refs in %s...\n"
	MsgRRRebasingFmt    = "  🔁 Pull --rebase in %s...\n"
	MsgRRReturnedFmt    = "  ↩ Returned to %s\n"
	ErrRRFetchFailedFmt = "  ✗ git fetch failed in %s: %v"
	ErrRRRebaseFailedFmt = "  ✗ git pull --rebase failed in %s: %v\n  Resolve the conflict, then re-run the release."
)

// Cross-dir clone-next (`cn <repo> <ver>`) messages.
const (
	MsgCNXStartingFmt = "  → clone-next for %s at %s (version %s)...\n"
	MsgCNXReturnedFmt = "  ↩ Returned to %s\n"
)

// SSH existing-key-on-disk fix messages.
const (
	MsgSSHExistsOnDisk = "\n  ℹ SSH key already exists on disk: %s\n  Reusing existing key (no regeneration needed).\n\n"
	MsgSSHForceHint    = "\n  💡 Pass --force to back up and regenerate this key.\n"
	MsgSSHBackedUp     = "  💾 Backed up existing key: %s.bak.<timestamp>\n"
	ErrSSHBackup       = "  ✗ Could not back up existing SSH key: %v\n"
)
