package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// runUpdateCleanup handles the "update-cleanup" subcommand.
// Removes leftover temp binaries and .old backup files.
func runUpdateCleanup() {
	fmt.Println(constants.MsgUpdateCleanStart)

	ctx := loadUpdateCleanupContext()
	total := cleanupTempArtifacts(ctx)
	total += cleanupBackupArtifacts(ctx)
	total += cleanupDriveRootShim(ctx)
	total += cleanupCloneSwapDirs(ctx)
	printUpdateCleanupResult(total)
}

// printUpdateCleanupResult reports the cleanup result summary.
func printUpdateCleanupResult(total int) {
	if total > 0 {
		fmt.Printf(constants.MsgUpdateCleanDone, total)

		return
	}

	fmt.Println(constants.MsgUpdateCleanNone)
}
