package cmd

// clonetermverifystate.go — request-scoped knob that turns on the
// --verify-cmd-faithful checker for the current command invocation.
//
// Why a package-level variable (vs. plumbing a bool through every
// hook signature): the existing print-row helpers
// (printCloneNowTermBlockRow / printCloneFromTermBlockRow /
// printCloneTermBlockForURL) are reached from 5+ call sites with
// fixed signatures dictated by clonenow.BeforeRowHook /
// clonefrom.BeforeRowHook. Adding a parameter to each signature
// would force a churning change in two executor packages plus every
// dry-run path. A package-level toggle set ONCE per `gitmap …`
// invocation (CLI is single-threaded at the dispatcher level) keeps
// the executor packages oblivious to verification AND keeps the
// hooks' signatures stable for tests/contracts.
//
// Concurrency note: clone batch modes use a worker pool, but each
// row-hook reads (never writes) cmdFaithfulVerify; the variable is
// set before Execute starts and never mutated again. Reads under a
// fixed-after-startup invariant don't need synchronization (Go's
// happens-before guarantees the dispatcher's set happens-before the
// goroutines' reads via the goroutine launch).

import "os"

// cmdFaithfulVerifyEnabled is the request-scoped flag. The package's
// CLI dispatchers (runClone / runCloneNow / runCloneFrom / runClonePick
// / runCloneNext) flip it to true via setCmdFaithfulVerify when
// --verify-cmd-faithful is parsed, and the print-row helpers consult
// it via cmdFaithfulVerifyEnabled() before running the checker.
//
// Default false so existing behavior is byte-identical when the flag
// is absent.
var cmdFaithfulVerify bool

// setCmdFaithfulVerify enables (or disables) the verifier for the
// remainder of the current process. Safe to call multiple times —
// last write wins, which matches the "set once at dispatcher" usage.
func setCmdFaithfulVerify(on bool) { cmdFaithfulVerify = on }

// cmdFaithfulVerifyEnabled returns true when the verifier should run.
// Predicate (vs. exposing the var) so a future move to atomic.Bool
// or a context-bound state stays a one-line refactor.
func cmdFaithfulVerifyEnabled() bool { return cmdFaithfulVerify }

// runCmdFaithfulCheck is the single integration point used by every
// per-row print helper. No-op when the flag is off so callers can
// invoke it unconditionally on the hot path.
//
// On mismatch it prints a structured report to stderr (matches the
// project stream contract: machine-pipeable per-repo blocks on
// stdout, diagnostics on stderr). It does NOT abort — the caller's
// executor decides whether a faithfulness mismatch should fail the
// clone. Today no caller aborts; the flag is informational ("here's
// where displayed/executed disagree, fix it before next release").
func runCmdFaithfulCheck(in CloneTermBlockInput, executorArgv []string) {
	if !cmdFaithfulVerifyEnabled() {
		return
	}
	report := VerifyCmdFaithful(in, executorArgv)
	if err := PrintCmdFaithfulReport(os.Stderr, report); err != nil {
		// Zero-swallow policy: surface the write failure but don't
		// abort the clone — the verifier is purely informational.
		_, _ = os.Stderr.WriteString(
			"  Warning: --verify-cmd-faithful: failed to write report: " +
				err.Error() + "\n")
	}
}
