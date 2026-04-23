package cmd

import (
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// runPending displays all pending tasks from the database.
func runPending() {
	checkHelp("pending", os.Args[2:])

	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.WarnPendingDBOpen, err)
		os.Exit(1)
	}
	defer db.Close()

	tasks, err := db.ListPendingTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrPendingTaskQuery, err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Print(constants.MsgPendingListEmpty)

		return
	}

	fmt.Print(constants.MsgPendingListHeader)

	for _, t := range tasks {
		fmt.Printf(constants.MsgPendingListRow, t.ID, t.TaskTypeName, t.TargetPath, t.FailureReason)
	}
}
