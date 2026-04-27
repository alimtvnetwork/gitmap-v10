package cmd

// clonetermrow.go — single-row helpers that build the standardized
// RepoTermBlock input for one clone-now / clone-from row at a time.
// Used by the streaming BeforeRow hooks so each block prints
// IMMEDIATELY before its row's git clone shells out (vs. the
// previous behavior of dumping every block up-front before
// Execute started).
//
// Design rationale: clonetermplan.go's batch helpers iterate the
// whole Plan and print N blocks back-to-back. That made the user
// wait through every ls-remote round-trip before any clone progress
// appeared. Streaming flips that — one ls-remote → one block → one
// clone → repeat — which matches what `gitmap clone <url1> <url2>`
// already does and gives the user a live, scannable transcript.

import (
	"github.com/alimtvnetwork/gitmap-v7/gitmap/clonefrom"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/clonenow"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// printCloneNowTermBlockRow emits one RepoTermBlock for one
// clone-now row. Mirrors the per-row branch-source logic from
// printCloneNowTermBlocks (the batch variant) so streaming and
// upfront output stay byte-for-byte identical per row — only the
// timing changes.
func printCloneNowTermBlockRow(index, total int, row clonenow.Row,
	url, dest string) {
	_ = total // total is reserved for a future "[i/N]" prefix; unused today
	branch := row.Branch
	source := "manifest"
	if len(branch) == 0 {
		branch = detectRemoteHEAD(url)
		source = remoteBranchSource(branch)
	}
	maybePrintCloneTermBlock(constants.OutputTerminal, CloneTermBlockInput{
		Index:        index,
		Name:         pickCloneNowName(row, dest),
		Branch:       branch,
		BranchSource: source,
		OriginalURL:  url,
		TargetURL:    url,
		Dest:         dest,
	})
}

// printCloneFromTermBlockRow emits one RepoTermBlock for one
// clone-from row. clone-from never rewrites URLs, so OriginalURL
// and TargetURL are both the row's URL. Branch source mirrors
// clonefrom.branchSourceForRow ("manifest" if pinned, else the
// ls-remote-discovered default — falling through to "(unknown)"
// inside the renderer if detection fails).
func printCloneFromTermBlockRow(index, total int, row clonefrom.Row,
	dest string) {
	_ = total // reserved for future "[i/N]" prefix
	branch := row.Branch
	source := "manifest"
	if len(branch) == 0 {
		branch = detectRemoteHEAD(row.URL)
		source = remoteBranchSource(branch)
	}
	maybePrintCloneTermBlock(constants.OutputTerminal, CloneTermBlockInput{
		Index:        index,
		Name:         dest,
		Branch:       branch,
		BranchSource: source,
		OriginalURL:  row.URL,
		TargetURL:    row.URL,
		Dest:         dest,
	})
}
