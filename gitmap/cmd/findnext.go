package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
)

// runFindNext dispatches `gitmap find-next [--scan-folder <id>] [--json]`.
func runFindNext(args []string) {
	checkHelp("find-next", args)

	scanFolderID, jsonOut := parseFindNextFlags(args)

	db := openSfDB()
	defer db.Close()

	rows, err := db.FindNext(scanFolderID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	emitFindNext(rows, jsonOut)
}

// parseFindNextFlags extracts --scan-folder and --json from args. Unknown
// tokens are silently ignored — find-next has no positional arguments.
func parseFindNextFlags(args []string) (int64, bool) {
	var scanFolderID int64
	jsonOut := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case constants.FindNextFlagJSON:
			jsonOut = true
		case constants.FindNextFlagScanFolder:
			if i+1 < len(args) {
				if v, err := strconv.ParseInt(args[i+1], 10, 64); err == nil {
					scanFolderID = v
				}
				i++
			}
		}
	}

	return scanFolderID, jsonOut
}

// emitFindNext writes either JSON or the human-readable summary.
func emitFindNext(rows []model.FindNextRow, jsonOut bool) {
	if jsonOut {
		emitFindNextJSON(rows)
		return
	}
	emitFindNextText(rows)
}

// emitFindNextJSON dumps the result array as indented JSON to stdout.
func emitFindNextJSON(rows []model.FindNextRow) {
	if rows == nil {
		rows = []model.FindNextRow{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rows); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// emitFindNextText prints the human summary (header + per-repo rows + hint).
func emitFindNextText(rows []model.FindNextRow) {
	if len(rows) == 0 {
		fmt.Print(constants.MsgFindNextEmpty)
		return
	}

	fmt.Printf(constants.MsgFindNextHeaderFmt, len(rows))
	for _, r := range rows {
		fmt.Printf(constants.MsgFindNextRowFmt,
			r.Repo.Slug, r.NextVersionTag, r.Method, r.ProbedAt, r.Repo.AbsolutePath)
	}
	fmt.Print(constants.MsgFindNextDoneFmt)
}
