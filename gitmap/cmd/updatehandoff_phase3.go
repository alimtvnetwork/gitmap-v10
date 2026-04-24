// Package cmd — Phase 3 of the self-update handoff chain.
//
// Phase 1 (update.go): active gitmap.exe → handoff copy (gitmap-update-<pid>.exe)
// Phase 2 (update.go): handoff copy runs build/deploy via run.ps1
// Phase 3 (this file): handoff copy spawns the freshly-deployed gitmap.exe
//
//	(a different file with no lock) detached, with a small delay, to
//	run `update-cleanup`. Only the deployed binary can safely remove
//	the still-locked handoff copy and the just-renamed *.exe.old.
//
// See spec/08-generic-update/06-cleanup.md and
// spec/03-general/02f-self-update-orchestration.md for the full sequence.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// scheduleDeployedCleanupHandoff hands off cleanup work to the freshly
// deployed gitmap binary. The deployed binary lives at a different path
// than the running handoff copy, so it can delete both:
//   - the handoff copy (gitmap-update-<pid>.exe) — locked by us
//   - the *.exe.old backup — sometimes briefly held by AV/Explorer
//
// On Windows we spawn a detached cmd.exe that pings briefly (so this
// process can exit and release its lock) and then runs the deployed
// binary's `update-cleanup`. On Unix we just exec it inline since
// no lock conflicts exist.
//
// Best-effort: any failure is silent. The user can always re-run
// `gitmap update-cleanup` manually.
func scheduleDeployedCleanupHandoff() {
	deployed := resolveDeployedBinaryPath()
	if len(deployed) == 0 {
		return
	}

	self, err := os.Executable()
	if err == nil && normalizeCleanupPath(self) == normalizeCleanupPath(deployed) {
		// We *are* the deployed binary (Unix in-place update). Just
		// run cleanup directly — no handoff needed.
		runUpdateCleanup()

		return
	}

	if runtime.GOOS != constants.OSWindows {
		spawnDeployedCleanupUnix(deployed)

		return
	}

	spawnDeployedCleanupWindows(deployed)
}

// resolveDeployedBinaryPath returns the path to the freshly-deployed
// gitmap binary on PATH, falling back to the active binary's expected
// deploy location. Returns "" if it cannot be determined.
func resolveDeployedBinaryPath() string {
	if path, err := exec.LookPath(constants.GitMapBin); err == nil {
		if resolved, evalErr := filepath.EvalSymlinks(path); evalErr == nil {
			return resolved
		}

		return path
	}

	self, err := os.Executable()
	if err != nil {
		return ""
	}

	dir := filepath.Dir(self)
	candidate := filepath.Join(dir, deployedBinaryName())
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}

	return ""
}

// deployedBinaryName returns the platform-specific deployed binary filename.
func deployedBinaryName() string {
	if runtime.GOOS == constants.OSWindows {
		return constants.GitMapBin + ".exe"
	}

	return constants.GitMapBin
}

// spawnDeployedCleanupWindows launches the deployed binary directly in a
// detached hidden process. We avoid `cmd.exe /C start ...` entirely because
// its quoting rules are brittle when combined with Go's Windows argument
// escaping and can surface GUI popups like "Windows cannot find '\\'".
func spawnDeployedCleanupWindows(deployed string) {
	fmt.Printf(constants.MsgUpdatePhase3Handoff, filepath.Base(deployed))
	fmt.Printf(constants.MsgUpdatePhase3Target, deployed)

	cmd := exec.Command(deployed, constants.CmdUpdateCleanup)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Env = append(os.Environ(), constants.EnvUpdateCleanupDelayMS+"=1500")
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrUpdatePhase3Handoff, deployed, err)
	}
}

// spawnDeployedCleanupUnix invokes the deployed binary's update-cleanup
// directly. No lock conflicts exist on Unix, so we don't need detachment.
func spawnDeployedCleanupUnix(deployed string) {
	fmt.Printf(constants.MsgUpdatePhase3Handoff, filepath.Base(deployed))

	cmd := exec.Command(deployed, constants.CmdUpdateCleanup)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
