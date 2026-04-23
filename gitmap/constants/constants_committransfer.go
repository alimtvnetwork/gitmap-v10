// Package constants centralizes every user-facing string in the CLI.
// This file owns the `commit-left` / `commit-right` / `commit-both`
// command surface (commit-transfer family).
//
// Spec: spec/01-app/106-commit-left-right-both.md
//
// **Status:** SCAFFOLD ONLY (v3.74.0). Constants, dispatch wiring, and
// helptext are in place; the actual replay engine in
// `gitmap/committransfer/` is deferred to its own session per the spec's
// phasing plan (§18). Running these commands today prints a clear
// "not yet implemented — see spec 106" message and exits 2.
//
// **Alias note:** Spec §13 reserves `cl`, `cr`, `cb`. Two of those are
// already taken at the top-level alias namespace
// (`cl` → changelog, `cr` → cpp-repos), so the commit-transfer family
// uses the disambiguated three-letter aliases `cml`, `cmr`, `cmb`. The
// long-form names (`commit-left` etc.) match the spec verbatim.
package constants

// gitmap:cmd top-level
// Commit-transfer commands.
const (
	CmdCommitLeft   = "commit-left"
	CmdCommitLeftA  = "cml"
	CmdCommitRight  = "commit-right"
	CmdCommitRightA = "cmr"
	CmdCommitBoth   = "commit-both"
	CmdCommitBothA  = "cmb"
)

// Commit-transfer log prefixes (matches merge-* style).
const (
	LogPrefixCommitLeft  = "[commit-left]"
	LogPrefixCommitRight = "[commit-right]"
	LogPrefixCommitBoth  = "[commit-both]"
)

// Commit-transfer flag names. See spec §8 for semantics.
const (
	FlagCTMirror          = "mirror"
	FlagCTIncludeMerges   = "include-merges"
	FlagCTLimit           = "limit"
	FlagCTSince           = "since"
	FlagCTStrip           = "strip"
	FlagCTNoStrip         = "no-strip"
	FlagCTDrop            = "drop"
	FlagCTNoDrop          = "no-drop"
	FlagCTConventional    = "conventional"
	FlagCTNoConventional  = "no-conventional"
	FlagCTProvenance      = "provenance"
	FlagCTNoProvenance    = "no-provenance"
	FlagCTPreferSource    = "prefer-source"
	FlagCTPreferTarget    = "prefer-target"
	FlagCTForceReplay     = "force-replay"
	FlagCTDryRun          = "dry-run"
	FlagCTYes             = "yes"
	FlagCTNoPush          = "no-push"
	FlagCTNoCommit        = "no-commit"
)

// Commit-transfer flag descriptions.
const (
	FlagDescCTMirror         = "Delete target files not present in source commit (true mirror)"
	FlagDescCTIncludeMerges  = "Include merge commits in the replay set"
	FlagDescCTLimit          = "Replay at most N source commits (oldest first); 0 = no limit"
	FlagDescCTSince          = "Override the divergence base (sha or ISO date)"
	FlagDescCTStrip          = "Add a strip pattern (regex, repeatable)"
	FlagDescCTNoStrip        = "Disable all strip patterns"
	FlagDescCTDrop           = "Add a drop pattern (regex, repeatable)"
	FlagDescCTNoDrop         = "Replay every commit (disable drop filter)"
	FlagDescCTConventional   = "Force conventional-commit normalization on"
	FlagDescCTNoConventional = "Disable conventional-commit normalization"
	FlagDescCTProvenance     = "Append provenance footer (default true)"
	FlagDescCTNoProvenance   = "Skip provenance footer"
	FlagDescCTPreferSource   = "Source side wins file conflicts"
	FlagDescCTPreferTarget   = "Target side wins file conflicts"
	FlagDescCTForceReplay    = "Replay even commits that already carry a gitmap-replay footer"
	FlagDescCTDryRun         = "Print the full plan + cleaned messages; perform no writes"
	FlagDescCTYes            = "Skip the confirmation prompt"
	FlagDescCTNoPush         = "Stop after the local commit (skip git push)"
	FlagDescCTNoCommit       = "Copy files but skip both commit and push"
)

// Commit-transfer messages and errors.
const (
	MsgCTUsageFmt              = "Usage: gitmap %s LEFT RIGHT [flags]\n\nSee 'gitmap help %s' for the full flag table.\n"
	ErrCTNotImplementedFmt     = "Error: '%s' is scaffolded but the replay engine is not yet implemented.\n  → Spec: spec/01-app/106-commit-left-right-both.md\n  → Track progress: gitmap reinstall once v3.75 lands.\n"
	ErrCTArgCountFmt           = "Error: %s requires exactly two endpoints (LEFT and RIGHT). Got %d.\n"
	ErrCTSourceCheckoutFmt     = "Error: failed to checkout %s in source: %v (operation: git checkout)\n"
	ErrCTReplayFailedFmt       = "Error: replay failed at source commit %s: %v (operation: %s)\n"
	MsgCTReplayPlanFmt         = "%s replaying %d commits from %s onto %s:\n"
	MsgCTReplayStepFmt         = "%s [%d/%d] %s → %s  %s\n"
	MsgCTReplaySkipFmt         = "%s [%d/%d] %s → -        skipped: %s\n"
	MsgCTSummaryFmt            = "%s done: replayed %d, skipped %d (%s)\n"
	MsgCTPushedFmt             = "%s pushed %d commits to %s\n"
	MsgCTConfirmProceedFmt     = "%s proceed? [y/N] "
)
