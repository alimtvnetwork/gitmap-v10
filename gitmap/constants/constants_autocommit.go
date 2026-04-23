package constants

// Auto-commit messages for post-release commit behavior.
const (
	MsgAutoCommitScanning    = "\n  ■ Checking for uncommitted changes...\n"
	MsgAutoCommitReleaseOnly = "  ✓ Release metadata committed: %s\n"
	MsgAutoCommitPushed      = "  ✓ Pushed to %s\n"
	MsgAutoCommitNone        = "  ✓ Working tree clean — nothing else to commit\n"
	MsgAutoCommitPrompt      = "  → Uncommitted changes outside .gitmap/release/:\n"
	MsgAutoCommitFile        = "      • %s\n"
	MsgAutoCommitAsk         = "  → Auto-commit these alongside the release? [y/N]: "
	MsgAutoCommitAll         = "  ✓ All changes committed: %s\n"
	MsgAutoCommitPartial     = "  ✓ Committed .gitmap/release/ changes only: %s\n"
	MsgAutoCommitSkipped     = "  → Skipped auto-commit (--no-commit)\n"
	MsgAutoCommitDryRun      = "  [dry-run] Would auto-commit release changes\n"
	MsgAutoCommitSyncRetry   = "  → Remote %s moved; rebasing and retrying push...\n"
	ErrAutoCommitFailed      = "  ✗ Auto-commit failed: %v\n"
	ErrAutoCommitPush        = "  ✗ Push failed: %v\n"
	AutoCommitMsgFmt         = "Release %s"
	FlagDescNoCommit         = "Skip post-release auto-commit and push"
	FlagDescYes              = "Auto-confirm all prompts (e.g. commit)"
	MsgAutoCommitAutoYes     = "  → Auto-confirmed via -y flag\n"

	// Git diff arguments for detecting changes.
	GitDiff         = "diff"
	GitDiffNameOnly = "--name-only"
	GitDiffCached   = "--cached"
)
