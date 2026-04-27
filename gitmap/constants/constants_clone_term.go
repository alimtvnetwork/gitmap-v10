package constants

import "time"

// Shared constants for the `--output terminal` adapter used by every
// clone-related command (clone, clone-next, clone-now, clone-pick,
// clone-from). Centralized here so a contract test can assert that
// every clone command surfaces the same flag name + description.

const (
	// CloneTermDetectTimeout caps how long `git ls-remote --symref
	// <url> HEAD` is allowed to run when previewing a clone in
	// `--output terminal` mode. The clone itself has no timeout —
	// this only protects the per-repo PREVIEW from a hung remote.
	CloneTermDetectTimeout = 4 * time.Second

	// BranchSourceRemoteHEAD is the canonical RepoTermBlock
	// BranchSource label used when the branch was discovered via
	// `git ls-remote --symref <url> HEAD` (i.e. the remote default
	// branch). Mirrors "HEAD" / "manifest" / "default HEAD" used by
	// the other clone commands so the rendered output is grep-able.
	BranchSourceRemoteHEAD = "remote HEAD"

	// FlagCloneTermOutput is the shared --output flag name used by
	// every clone-related command. clone-next and clone-from already
	// use the same string ("output") but kept as a constant so a
	// future rename only happens in one place.
	FlagCloneTermOutput = "output"

	// FlagDescCloneTermOutput is the shared description for clone,
	// clone-now, and clone-pick. clone-next and clone-from keep
	// their existing descriptions verbatim to avoid breaking help
	// snapshots; new commands use this one. Wording calls out the
	// stream split explicitly so users know where to redirect:
	// blocks go to stdout (machine-pipeable), git progress + the
	// human summary go to stderr.
	FlagDescCloneTermOutput = "Per-repo summary format: '' (legacy) or " +
		"'terminal' (standardized branch/from/to/command block on " +
		"stdout, streamed before each clone; git progress stays on stderr)"
)
