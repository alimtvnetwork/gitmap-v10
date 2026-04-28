package cmd

// Shared subprocess test harness for the `cliexit_*_test.go` suite.
//
// These tests assert the documented exit codes of `gitmap` subcommands
// (success / user-canceled / failure). Asserting *real* exit codes
// requires running the compiled binary out-of-process because the
// production code calls os.Exit directly — there's no in-process
// "return an int" seam to stub.
//
// Wiring overview:
//
//   1. TestMain builds the gitmap binary once into t.TempDir-style
//      shared cache (under os.TempDir) and reuses it for every test.
//   2. runGitmap(t, args, stdin) execs the binary with a hermetic
//      working dir + minimal env, pipes optional stdin, and returns
//      (exitCode, stdout, stderr).
//   3. Per-command files (cliexit_scan_test.go, cliexit_clone_test.go)
//      drive the harness with table-driven cases.
//
// The whole suite skips when `go` is not on PATH (stripped CI images)
// or the build itself fails for sandbox reasons (e.g. cgo disabled
// without a C toolchain). That keeps the larger test matrix green
// while still failing loudly on real exit-code regressions in any
// environment that *can* build the binary.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// gitmapBinary holds the path to the once-built binary. Populated
// lazily by ensureGitmapBinary so a test run that touches none of
// the cliexit files pays no build cost.
var (
	gitmapBinary    string
	gitmapBuildErr  error
	gitmapBuildOnce sync.Once
)

// ensureGitmapBinary builds the gitmap binary the first time it is
// called and caches the result. Returns the absolute path or skips
// the calling test when the build couldn't run (no `go`, sandbox
// blocks compilation, etc.).
func ensureGitmapBinary(t *testing.T) string {
	t.Helper()
	gitmapBuildOnce.Do(buildGitmapBinaryOnce)
	if gitmapBuildErr != nil {
		t.Skipf("gitmap binary unavailable for cliexit tests: %v", gitmapBuildErr)
	}

	return gitmapBinary
}

// buildGitmapBinaryOnce is invoked under sync.Once so concurrent
// t.Parallel tests share a single artifact.
func buildGitmapBinaryOnce() {
	if _, err := exec.LookPath("go"); err != nil {
		gitmapBuildErr = err

		return
	}
	out := filepath.Join(os.TempDir(), gitmapBinaryName())
	// Build from the gitmap module root. The cwd of `go test` is
	// already the package dir (gitmap/cmd) so "../" is the module.
	cmd := exec.Command("go", "build", "-o", out, "./")
	cmd.Dir = ".." // gitmap/ module root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		gitmapBuildErr = wrapBuildErr(err, &stderr)

		return
	}
	gitmapBinary = out
}

// gitmapBinaryName returns the right artifact name per platform.
func gitmapBinaryName() string {
	name := "gitmap_cliexit_test"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	return name
}

// wrapBuildErr produces a single-line error that carries the tail of
// `go build` stderr so a failing CI log shows the real reason.
func wrapBuildErr(err error, stderr *bytes.Buffer) error {
	tail := strings.TrimSpace(stderr.String())
	if tail == "" {
		return err
	}

	return &buildError{cause: err, stderr: tail}
}

// buildError formats the underlying error + stderr tail.
type buildError struct {
	cause  error
	stderr string
}

// Error renders both the cause and the captured stderr so the test
// log contains everything needed to diagnose a build failure.
func (e *buildError) Error() string {
	return e.cause.Error() + ": " + e.stderr
}

// runGitmap executes the prebuilt binary with args + optional stdin
// and returns (exit code, stdout, stderr). Wraps the awkward parts
// of os/exec so call sites stay declarative.
func runGitmap(t *testing.T, args []string, stdin string) (int, string, string) {
	t.Helper()
	bin := ensureGitmapBinary(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = t.TempDir()
	cmd.Env = hermeticEnv()
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return extractExitCode(err), stdout.String(), stderr.String()
}

// hermeticEnv strips variables that could change behavior between
// developer machines and CI (NO_COLOR honored so terminal renderers
// produce stable output if the test ever asserts on stdout).
func hermeticEnv() []string {
	keep := []string{"PATH", "HOME", "USERPROFILE", "SystemRoot", "TEMP", "TMP", "TMPDIR"}
	out := make([]string, 0, len(keep)+2)
	for _, k := range keep {
		if v := os.Getenv(k); v != "" {
			out = append(out, k+"="+v)
		}
	}
	out = append(out, "NO_COLOR=1")
	out = append(out, "GITMAP_DISABLE_TASKS=1")

	return out
}

// extractExitCode normalizes os/exec's error -> int conversion. A
// non-ExitError (couldn't start, signal, etc.) returns -1 so the
// caller's table assertion fails with a clear "got -1" rather than
// silently passing on an exit-0 default.
func extractExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if asExitError(err, &exitErr) {
		return exitErr.ExitCode()
	}

	return -1
}

// asExitError isolates the type-assert so extractExitCode stays one
// line of logic per branch (project style cap on cyclomatic noise).
func asExitError(err error, target **exec.ExitError) bool {
	e, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}
	*target = e

	return true
}
