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

	// FlagCloneVerifyCmdFaithful is the shared boolean flag name used
	// by every clone-related command to enable the dry-run verifier
	// that compares the rendered `cmd:` line against the executor's
	// real argv. Single constant so a future rename happens in one
	// place and `gitmap <cmd> --help` stays consistent across surfaces.
	FlagCloneVerifyCmdFaithful = "verify-cmd-faithful"

	// FlagDescCloneVerifyCmdFaithful explains the flag in --help.
	// Wording calls out that it's a SAFETY check (no execution
	// difference, no output unless something is wrong) so users feel
	// safe leaving it on in CI.
	FlagDescCloneVerifyCmdFaithful = "Verify the rendered cmd: line " +
		"matches the executor's real git argv. Prints a structured " +
		"mismatch report to stderr on divergence; silent on match. " +
		"Pure check — does not change clone behavior."

	// FlagClonePrintArgv is the shared boolean flag name used by every
	// clone-related command to dump the exact argv tokens that would
	// be passed to exec.Command. Companion to --verify-cmd-faithful:
	// the verifier proves displayed==executed, this flag SHOWS the
	// executed form so users can paste it elsewhere or grep one token
	// at a time without re-quoting the cmd: string.
	FlagClonePrintArgv = "print-clone-argv"

	// FlagDescClonePrintArgv explains the flag in --help. Wording
	// calls out the format (one token per line, prefixed `argv[i]=`)
	// and the stream (stderr, not stdout) so users redirecting the
	// terminal block to a file aren't surprised by extra noise.
	FlagDescClonePrintArgv = "Print the exact git-clone argv tokens " +
		"to stderr (one per line, `argv[i]=<token>`) right after each " +
		"terminal block. Audit-only — does not change clone behavior."
)


