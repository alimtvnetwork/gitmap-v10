// Package cmd — latest-branch output formatters.
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/gitutil"
)

// latestBranchJSON is the JSON output structure.
type latestBranchJSON struct {
	Branch     []string              `json:"branch"`
	Remote     string                `json:"remote"`
	Sha        string                `json:"sha"`
	CommitDate string                `json:"commitDate"`
	Subject    string                `json:"subject"`
	Ref        string                `json:"ref"`
	Top        []latestBranchTopItem `json:"top,omitempty"`
}

// latestBranchTopItem is a single entry in the top-N list.
type latestBranchTopItem struct {
	Branch     string `json:"branch"`
	Sha        string `json:"sha"`
	CommitDate string `json:"commitDate"`
	Subject    string `json:"subject"`
}

// dispatchLatestOutput routes to the correct output formatter.
func dispatchLatestOutput(result latestBranchResult, items []gitutil.RemoteBranchInfo, cfg latestBranchConfig) {
	if cfg.format == constants.OutputJSON {
		printLatestJSON(result, items, cfg.top)

		return
	}
	if cfg.format == constants.OutputCSV {
		printLatestCSV(items, result.selectedRemote, cfg.top)

		return
	}
	printLatestTerminal(result, items, cfg.top)
}

// printLatestJSON outputs the latest branch result as JSON.
func printLatestJSON(result latestBranchResult, items []gitutil.RemoteBranchInfo, top int) {
	out := buildLatestJSON(result)
	if top > 0 {
		out.Top = buildTopItems(items, top)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", constants.JSONIndent)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ Failed to encode latest branch JSON: %v\n", err)
	}
}

// buildLatestJSON constructs the base JSON output struct.
func buildLatestJSON(result latestBranchResult) latestBranchJSON {

	return latestBranchJSON{
		Branch:     result.branchNames,
		Remote:     result.selectedRemote,
		Sha:        result.shortSha,
		CommitDate: result.commitDate,
		Subject:    result.latest.Subject,
		Ref:        result.latest.RemoteRef,
	}
}

// buildTopItems constructs the top-N list for JSON output.
func buildTopItems(items []gitutil.RemoteBranchInfo, top int) []latestBranchTopItem {
	count := top
	if count > len(items) {
		count = len(items)
	}
	topItems := make([]latestBranchTopItem, 0, count)
	for _, item := range items[:count] {
		topItems = append(topItems, latestBranchTopItem{
			Branch:     gitutil.StripRemotePrefix(item.RemoteRef),
			Sha:        gitutil.TruncSha(item.Sha),
			CommitDate: gitutil.FormatDisplayDate(item.CommitDate),
			Subject:    item.Subject,
		})
	}

	return topItems
}

// printLatestCSV outputs the latest branch result as CSV.
func printLatestCSV(items []gitutil.RemoteBranchInfo, remote string, top int) {
	count := resolveTopCount(top, len(items))
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(constants.LatestBranchCSVHeaders); err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ Failed to write CSV header: %v\n", err)

		return
	}
	for _, item := range items[:count] {
		writeCSVRow(w, item, remote)
	}
	w.Flush()
}

// resolveTopCount determines how many items to display.
func resolveTopCount(top, total int) int {
	count := 1
	if top > 0 {
		count = top
	}
	if count > total {
		count = total
	}

	return count
}

// writeCSVRow writes a single CSV row for a branch item.
func writeCSVRow(w *csv.Writer, item gitutil.RemoteBranchInfo, remote string) {
	if err := w.Write([]string{
		gitutil.StripRemotePrefix(item.RemoteRef),
		remote,
		gitutil.TruncSha(item.Sha),
		gitutil.FormatDisplayDate(item.CommitDate),
		item.Subject,
		item.RemoteRef,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ Failed to write CSV row: %v\n", err)
	}
}

// printLatestTerminal outputs the latest branch result as text.
func printLatestTerminal(result latestBranchResult, items []gitutil.RemoteBranchInfo, top int) {
	fmt.Println()
	printTerminalHeader(result)
	if top > 0 {
		printTerminalTopTable(items, result.selectedRemote, top)
	}
	fmt.Println()
}

// printTerminalHeader prints the main latest-branch info block.
func printTerminalHeader(result latestBranchResult) {
	fmt.Printf(constants.LBTermLatestFmt, strings.Join(result.branchNames, ", "))
	fmt.Printf(constants.LBTermRemoteFmt, result.selectedRemote)
	fmt.Printf(constants.LBTermSHAFmt, result.shortSha)
	fmt.Printf(constants.LBTermDateFmt, result.commitDate)
	fmt.Printf(constants.LBTermSubjectFmt, result.latest.Subject)
	fmt.Printf(constants.LBTermRefFmt, result.latest.RemoteRef)
}

// printTerminalTopTable prints the top-N branches table.
func printTerminalTopTable(items []gitutil.RemoteBranchInfo, remote string, top int) {
	count := resolveTopCount(top, len(items))
	fmt.Println()
	fmt.Printf(constants.LBTermTopHdrFmt, count, remote)
	printTerminalTopHeader()
	for _, item := range items[:count] {
		printTerminalTopRow(item)
	}
}

// printTerminalTopHeader prints the table column headers.
func printTerminalTopHeader() {
	fmt.Printf(constants.LBTermRowFmt,
		constants.LatestBranchTableColumns[0], constants.LatestBranchTableColumns[1],
		constants.LatestBranchTableColumns[2], constants.LatestBranchTableColumns[3])
}

// printTerminalTopRow prints a single branch row.
func printTerminalTopRow(item gitutil.RemoteBranchInfo) {
	fmt.Printf(constants.LBTermRowFmt,
		gitutil.FormatDisplayDate(item.CommitDate),
		gitutil.StripRemotePrefix(item.RemoteRef),
		gitutil.TruncSha(item.Sha),
		item.Subject)
}
