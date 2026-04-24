package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// runUpdateCleanup handles the "update-cleanup" subcommand.
// Removes leftover temp binaries and .old backup files.
func runUpdateCleanup() {
	fmt.Println(constants.MsgUpdateCleanStart)
	delayUpdateCleanupIfNeeded()

	ctx := loadUpdateCleanupContext()
	total := cleanupTempArtifacts(ctx)
	total += cleanupBackupArtifacts(ctx)
	total += cleanupDriveRootShim(ctx)
	total += cleanupCloneSwapDirs(ctx)
	printUpdateCleanupResult(total)
}

// delayUpdateCleanupIfNeeded gives the just-exited handoff/update process time
// to release Windows file handles before deletion begins.
func delayUpdateCleanupIfNeeded() {
	raw := os.Getenv(constants.EnvUpdateCleanupDelayMS)
	if len(raw) == 0 {
		return
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms <= 0 {
		return
	}
	fmt.Printf(constants.MsgUpdateCleanDelay, ms)
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// printUpdateCleanupResult reports the cleanup result summary.
func printUpdateCleanupResult(total int) {
	if total > 0 {
		fmt.Printf(constants.MsgUpdateCleanDone, total)

		return
	}

	fmt.Println(constants.MsgUpdateCleanNone)
}
