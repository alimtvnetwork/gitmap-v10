package cmd

// Two-pass orchestration for `gitmap regoldens`. Split from
// regoldens.go to respect the 200-line file cap and to keep each
// orchestration step under the 15-line function cap. The diff
// summary (when --diff is set) runs between pass 1 and pass 2 and
// fires regardless of pass-1 success so contributors can see what
// pass 1 wrote even when it failed mid-way.

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// executeRegoldens is the top-level orchestrator. Each step is a
// dedicated helper so the control flow reads top-to-bottom and the
// per-function line budget stays comfortably under 15.
func executeRegoldens(cfg regoldensFlags) {
	pass1Code := runRegoldensPassCapture(cfg, true, constants.MsgRegoldensPass1Header)
	maybeEmitDiffSummary(cfg)
	exitOnPass1Failure(pass1Code)
	if handleSkipVerify(cfg) {
		return
	}
	runPass2AndAnnounce(cfg)
}

// runRegoldensPassCapture prints the header and runs one pass,
// returning the exit code instead of exiting. Used so downstream
// work (diff summary) runs whether the pass succeeded or failed.
func runRegoldensPassCapture(cfg regoldensFlags, withGate bool, header string) int {
	fmt.Fprint(os.Stderr, header)
	return runGoTestPass(cfg, withGate)
}

// maybeEmitDiffSummary fires the post-pass-1 diff report when the
// user passed --diff. Kept as a tiny shim to keep executeRegoldens
// linear and readable.
func maybeEmitDiffSummary(cfg regoldensFlags) {
	if cfg.showDiff {
		emitGoldenDiffSummary()
	}
}

// emitGoldenDiffSummary runs git status to find changed golden files
// and prints a concise summary to stderr.
func emitGoldenDiffSummary() {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return
	}
	lines := strings.Split(string(out), "\n")
	var changed []string
	for _, line := range lines {
		if strings.Contains(line, ".golden") {
			changed = append(changed, strings.TrimSpace(line[2:]))
		}
	}
	if len(changed) > 0 {
		fmt.Fprintln(os.Stderr, "\nChanged golden files:")
		for _, f := range changed {
			fmt.Fprintf(os.Stderr, "  %s\n", f)
		}
		fmt.Fprintln(os.Stderr)
	}
}

// exitOnPass1Failure logs the pass-1 error and exits 1 when pass 1
// returned non-zero. Called AFTER the diff summary so users see the
// partial fixture state before the process dies.
func exitOnPass1Failure(code int) {
	if code == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, constants.ErrRegoldensPass1Failed, code)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

// handleSkipVerify emits the skip-verify success path and returns
// true when --skip-verify is set so the caller can short-circuit
// the verify pass.
func handleSkipVerify(cfg regoldensFlags) bool {
	if !cfg.skipVerify {
		return false
	}
	fmt.Fprint(os.Stderr, constants.MsgRegoldensSkipVerify)
	fmt.Fprintf(os.Stdout, constants.MsgRegoldensSuccessNoVeri,
		cfg.pattern, cfg.pkg)
	return true
}

// runPass2AndAnnounce executes the determinism verification pass
// and prints the final success line. Pass 2 exits 1 internally on
// failure (via runRegoldensPass), so reaching the final Fprintf
// implies both passes succeeded.
func runPass2AndAnnounce(cfg regoldensFlags) {
	runRegoldensPass(cfg, false,
		constants.MsgRegoldensPass2Header,
		constants.ErrRegoldensPass2Failed)
	fmt.Fprintf(os.Stdout, constants.MsgRegoldensSuccess,
		cfg.pattern, cfg.pkg)
}
