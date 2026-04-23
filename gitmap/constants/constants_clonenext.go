package constants

// Clone-next command messages.
const (
	MsgCloneNextCloning      = "Cloning %s into %s...\n"
	MsgCloneNextCreating     = "Creating GitHub repo %s...\n"
	MsgCloneNextCreated      = "✓ Created GitHub repo %s\n"
	MsgCloneNextDone         = "✓ Cloned %s\n"
	MsgCloneNextDesktop      = "✓ Registered %s with GitHub Desktop\n"
	MsgCloneNextRemovePrompt = "Remove current folder %s? [y/N] "
	MsgCloneNextRemoved      = "✓ Removed %s\n"
	MsgCloneNextMovedTo      = "→ Now in %s\n"
	MsgFlattenFallback       = "→ Falling back to versioned folder %s (current folder is locked by this shell)\n"
	MsgFlattenLockedHint     = "  Tip: 'cd ..' out of %s in your shell, then re-run to flatten — or pass -f to force.\n"
	// MsgForceReleasing is printed by `cn -f` BEFORE the chdir-to-parent
	// trick that releases the Windows file lock on the cwd. Telling the
	// user "I'm leaving X" up front keeps the subsequent removal log
	// from looking like an unrelated jump.
	MsgForceReleasing = "  → Stepping out of %s to release the file lock\n"
	// Stage-banner messages emitted by `cn` to group the output into
	// three clear phases (PREPARE → CLONE → FINALIZE). Plain ASCII rule
	// glyphs so PowerShell renders them identically to bash.
	MsgCNStagePrepare  = "\n  ── 1/3  Preparing flatten (%s → %s) ──\n"
	MsgCNStageClone    = "\n  ── 2/3  Cloning %s ──\n"
	MsgCNStageFinalize = "\n  ── 3/3  Finalizing ──\n"
	MsgCNDone          = "\n  ── ✓ Done — now in %s ──\n\n"
)

// Clone-next error and warning messages.
const (
	ErrCloneNextUsage         = "Usage: gitmap clone-next <v++|vN> [flags]"
	ErrCloneNextCwd           = "Error: cannot determine current directory: %v\n"
	ErrCloneNextNoRemote      = "Error: not a git repo or no remote origin: %v\n"
	ErrCloneNextBadVersion    = "Error: %v\n"
	ErrCloneNextExists        = "Error: target directory already exists: %s\nUse 'cd' to switch to it.\n"
	ErrCloneNextFailed        = "Error: clone failed for %s\n"
	ErrCloneNextRemoteParse   = "Error: cannot parse remote URL: %v\n"
	ErrCloneNextRepoCheck     = "Error: cannot check target repo: %v\n"
	ErrCloneNextRepoCreate    = "Error: cannot create GitHub repo %s: %v\n"
	WarnCloneNextRemoveFailed = "Warning: could not remove %s: %v\n"
	// ErrCloneNextForceFailed fires when -f / --force was passed but
	// the existing flattened folder still cannot be removed (e.g. a
	// non-shell process holds a handle on it). We refuse the silent
	// versioned-folder fallback here — the user explicitly asked for
	// a flat layout, so a clear error beats a surprise rename.
	ErrCloneNextForceFailed = "Error: --force could not remove %s: %v\nClose any process holding files in %s and re-run.\n"
)

// Clone-next flag descriptions.
const (
	FlagDescCloneNextDelete       = "Auto-remove current folder after clone"
	FlagDescCloneNextKeep         = "Keep current folder without prompting"
	FlagDescCloneNextNoDesktop    = "Skip GitHub Desktop registration"
	FlagDescCloneNextCreateRemote = "Create target GitHub repo if it does not exist (requires GITHUB_TOKEN)"
	FlagDescCloneNextCSV          = "Read repo paths from CSV file (one path per row, header optional)"
	FlagDescCloneNextAll          = "Walk current folder and run cn on every git repo found one level deep"
	FlagDescCloneNextForce        = "Force flatten even when cwd is the target folder (chdir to parent, no versioned fallback)"
)

// Clone-next help strings for usage output.
const (
	HelpCloneNextFlags = "Clone-Next Flags:"
	HelpCNDelete       = "  --delete            Auto-remove current version folder after clone"
	HelpCNKeep         = "  --keep              Keep current folder without prompting for removal"
	HelpCNNoDesktop    = "  --no-desktop        Skip GitHub Desktop registration"
	HelpCNSSHKey       = "  --ssh-key, -K       SSH key name to use for clone"
	HelpCNVerbose      = "  --verbose           Show detailed clone-next output"
	HelpCNCreateRemote = "  --create-remote     Create target GitHub repo if missing (needs GITHUB_TOKEN)"
	HelpCNCSV          = "  --csv <path>        Batch mode: read repo list from CSV (one path per row)"
	HelpCNAll          = "  --all               Batch mode: cn every git repo one level under cwd"
	HelpCNForce        = "  --force, -f         Force flatten when cwd IS the target folder (no versioned fallback)"
)

// Clone-next batch mode messages and statuses (v3.42.0+).
const (
	MsgCloneNextBatchStart    = "→ Batch cn over %d repo(s)\n"
	MsgCloneNextBatchRepo     = "  • %s: %s -> %s\n"
	MsgCloneNextBatchUpToDate = "  • %s: %s (no update needed)\n"
	MsgCloneNextBatchSummary  = "✓ Batch complete: %d ok, %d failed, %d skipped\n"
	MsgCloneNextBatchReport   = "  Report: %s\n"
	WarnCloneNextBatchReport  = "Warning: could not write batch report: %v\n"
	ErrCloneNextBatchLoad     = "Error: could not load batch input: %v\n"

	BatchStatusOK      = "ok"
	BatchStatusFailed  = "failed"
	BatchStatusSkipped = "skipped"

	// BatchDetailUpToDate is the row's `detail` field when the local
	// repo's version equals the highest existing remote sibling.
	BatchDetailUpToDate = "no update needed"
)
