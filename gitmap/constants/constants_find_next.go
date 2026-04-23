package constants

// find-next: list every repo whose latest probe row reports an available
// upgrade (Phase 2.4, v3.9.0).
//
// Joins Repo against the newest VersionProbe per repo (via correlated
// subquery on MAX(ProbedAt)) and filters to IsAvailable=1. Optionally
// scoped to a single ScanFolderId so callers can query "what's new in
// E:\src" without seeing unrelated repos.

// SQL: every repo whose latest VersionProbe row has IsAvailable=1.
// Sort by NextVersionNum DESC so the freshest tags float to the top.
const SQLSelectFindNext = `
SELECT r.RepoId, r.Slug, r.RepoName, r.HttpsUrl, r.SshUrl, r.Branch,
       r.RelativePath, r.AbsolutePath, r.CloneInstruction, r.Notes,
       p.NextVersionTag, p.NextVersionNum, p.Method, p.ProbedAt
FROM Repo r
JOIN VersionProbe p ON p.RepoId = r.RepoId
WHERE p.IsAvailable = 1
  AND p.ProbedAt = (
    SELECT MAX(ProbedAt) FROM VersionProbe WHERE RepoId = r.RepoId
  )
ORDER BY p.NextVersionNum DESC, r.Slug ASC`

// SQL: same as above, scoped to a specific ScanFolderId.
const SQLSelectFindNextByScanFolder = `
SELECT r.RepoId, r.Slug, r.RepoName, r.HttpsUrl, r.SshUrl, r.Branch,
       r.RelativePath, r.AbsolutePath, r.CloneInstruction, r.Notes,
       p.NextVersionTag, p.NextVersionNum, p.Method, p.ProbedAt
FROM Repo r
JOIN VersionProbe p ON p.RepoId = r.RepoId
WHERE p.IsAvailable = 1
  AND r.ScanFolderId = ?
  AND p.ProbedAt = (
    SELECT MAX(ProbedAt) FROM VersionProbe WHERE RepoId = r.RepoId
  )
ORDER BY p.NextVersionNum DESC, r.Slug ASC`

// find-next user-facing strings.
const (
	MsgFindNextEmpty       = "No repos with available updates. Run `gitmap probe --all` first.\n"
	MsgFindNextHeaderFmt   = "Available updates (%d):\n"
	MsgFindNextRowFmt      = "  %s → %s [method=%s, probed=%s]\n      %s\n"
	MsgFindNextDoneFmt     = "Hint: run `gitmap pull` or `gitmap cn next all` to apply.\n"
	ErrFindNextQuery       = "find-next: failed to query: %v"
	ErrFindNextScanRow     = "find-next: failed to scan row: %v"
	MsgFindNextUsageHeader = "Usage: gitmap find-next [--scan-folder <id>] [--json]"
)

// find-next CLI flag tokens.
const (
	FindNextFlagScanFolder = "--scan-folder"
	FindNextFlagJSON       = "--json"
)

// gitmap:cmd top-level
// find-next CLI commands.
const (
	CmdFindNext      = "find-next"
	CmdFindNextAlias = "fn"
)
