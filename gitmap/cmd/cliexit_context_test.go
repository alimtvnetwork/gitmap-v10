package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runGitmap executes the gitmap binary with the given args.
// It returns the exit code, stdout, and stderr.
func runGitmap(t *testing.T, args []string, stdin string) (int, string, string) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "gitmap")
	
	// Build the binary for the current test run
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build gitmap: %v\n%s", err, out)
	}

	cmd = exec.Command(bin, args...)
	cmd.Stdin = strings.NewReader(stdin)
	
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run gitmap: %v", err)
		}
	}
	
	return exitCode, stdout.String(), stderr.String()
}

// containsCI checks if haystack contains needle (case-insensitive).
func containsCI(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// Integration tests asserting that user-facing failure stderr from
// scan and clone-family commands carries the standardized context
// fields produced by gitmap/cliexit:
//
//   - the command attempted (e.g. "gitmap scan", "gitmap clone-from")
//   - the subject path the command was operating on (the failing
//     repo / manifest path)
//   - some recognizable trace of the underlying error (a phrase
//     specific to the failure mode, not a generic "error")
//
// These run against the real built binary so we exercise the actual
// stderr that wrapper scripts and CI annotations consume — not an
// in-process formatter mock.

// TestCLI_FailureContext_Scan asserts a missing-directory scan
// surfaces the command name + the missing path on stderr. The
// "underlying error" check looks for any of the documented
// not-exist phrasings so OS-specific wording (Linux vs Windows)
// doesn't make the test brittle.
func TestCLI_FailureContext_Scan(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "definitely-not-here")
	code, stdout, stderr := runGitmap(t, []string{"scan", "--quiet", missing}, "")
	if code == 0 {
		t.Fatalf("scan of missing dir unexpectedly succeeded\nstdout=%s\nstderr=%s",
			stdout, stderr)
	}
	assertStderrContext(t, "scan", stderr, missing, []string{
		"no such file", "cannot find", "does not exist", "not exist",
	})
}

// TestCLI_FailureContext_CloneFromMissingManifest drives clone-from
// against a manifest path that doesn't exist. Asserts the command
// label, the manifest path, and an open/read failure phrase.
func TestCLI_FailureContext_CloneFromMissingManifest(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "no-such-manifest.json")
	code, stdout, stderr := runGitmap(t,
		[]string{"clone-from", "--execute", missing}, "")
	if code == 0 {
		t.Fatalf("clone-from with missing manifest unexpectedly succeeded\nstdout=%s\nstderr=%s",
			stdout, stderr)
	}
	assertStderrContext(t, "clone-from", stderr, missing, []string{
		"no such file", "cannot find", "does not exist", "not exist", "open",
	})
}

// TestCLI_FailureContext_CloneNowMissingManifest is the clone-now
// counterpart. Distinct from clone-from: clone-now uses a different
// parser path and we want both surfaces validated.
func TestCLI_FailureContext_CloneNowMissingManifest(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "no-such-manifest.json")
	code, stdout, stderr := runGitmap(t,
		[]string{"clone-now", "--execute", "--quiet", missing}, "")
	if code == 0 {
		t.Fatalf("clone-now with missing manifest unexpectedly succeeded\nstdout=%s\nstderr=%s",
			stdout, stderr)
	}
	// clone-now's parser and dispatcher both refer to the
	// command via its alias chain ("clone-now" / "reclone"); we
	// only require *some* form of the command label to be
	// present, plus the path and an open/read failure phrase.
	assertStderrContextAny(t, []string{"clone-now", "reclone"}, stderr, missing, []string{
		"no such file", "cannot find", "does not exist", "not exist", "open",
	})
}

// assertStderrContext is the per-command wrapper around the more
// general assertStderrContextAny. Single-command label form keeps
// the common case readable.
func assertStderrContext(t *testing.T, command, stderr, subject string, errPhrases []string) {
	t.Helper()
	assertStderrContextAny(t, []string{command}, stderr, subject, errPhrases)
}

// assertStderrContextAny verifies stderr carries (a) at least one of
// the acceptable command labels, (b) the subject string verbatim,
// and (c) at least one of the underlying-error phrases. Each missing
// piece is reported separately so a failing run pinpoints exactly
// which contract field regressed.
func assertStderrContextAny(t *testing.T, commands []string, stderr, subject string, errPhrases []string) {
	t.Helper()
	if !containsAnyCI(stderr, commands) {
		t.Errorf("stderr missing command label (any of %v)\nstderr=%s",
			commands, stderr)
	}
	if !strings.Contains(stderr, subject) {
		t.Errorf("stderr missing subject path %q\nstderr=%s", subject, stderr)
	}
	if !containsAnyCI(stderr, errPhrases) {
		t.Errorf("stderr missing underlying-error phrase (any of %v)\nstderr=%s",
			errPhrases, stderr)
	}
}

// containsAnyCI returns true when haystack contains any of the
// needles (case-insensitive). Pulled out so the assertion helpers
// stay one-liners.
func containsAnyCI(haystack string, needles []string) bool {
	for _, n := range needles {
		if containsCI(haystack, n) {
			return true
		}
	}

	return false
}
