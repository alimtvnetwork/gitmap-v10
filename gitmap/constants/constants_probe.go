package constants

// VersionProbe operations (v3.8.0+, Phase 2.3).
//
// The probe inspects a repo's remote to discover the next available
// version tag. Hybrid strategy: try `git ls-remote` against the HEAD
// first (cheap, network-only), and only fall back to a `--depth 1
// --filter=blob:none` clone when ls-remote returns nothing usable.
//
// Results land in the VersionProbe table. The "Method" column records
// which strategy succeeded ("ls-remote" or "shallow-clone"); "Error"
// captures the failure reason when IsAvailable = 0 so operators can
// debug a probe without re-running it.

// Probe method tokens (stored in VersionProbe.Method).
const (
	ProbeMethodLsRemote     = "ls-remote"
	ProbeMethodShallowClone = "shallow-clone"
	ProbeMethodNone         = "none"
)

// SQL: insert a new probe row.
const SQLInsertVersionProbe = `INSERT INTO VersionProbe
	(RepoId, NextVersionTag, NextVersionNum, Method, IsAvailable, Error)
	VALUES (?, ?, ?, ?, ?, ?)`

// SQL: latest probe per repo.
const SQLSelectLatestVersionProbe = `SELECT VersionProbeId, RepoId, ProbedAt,
		NextVersionTag, NextVersionNum, Method, IsAvailable, Error
	FROM VersionProbe WHERE RepoId = ?
	ORDER BY ProbedAt DESC, VersionProbeId DESC LIMIT 1`

// SQL: bulk-tag every repo whose AbsolutePath was just scanned with the
// active ScanFolderId. Path list is interpolated as `?,?,?,...` because
// SQLite has no array binding.
const SQLTagReposByScanFolderTpl = `UPDATE Repo SET ScanFolderId = ? WHERE AbsolutePath IN (%s)`

// VersionProbe error/message strings.
const (
	ErrProbeOpenDB       = "version probe: failed to open database: %v"
	ErrProbeMissingURL   = "version probe: repo %q has no clone URL"
	ErrProbeLsRemoteFail = "ls-remote failed: %v"
	ErrProbeCloneFail    = "shallow clone failed: %v"
	ErrProbeRecord       = "version probe: failed to record result for repo %d: %v"
	ErrProbeNoRepo       = "version probe: no repo found at %q"
	ErrProbeTagFail      = "scan: failed to tag repos with scan folder %d: %v"
)

// VersionProbe user-facing CLI strings.
const (
	MsgProbeStartFmt    = "→ Probing %d repo(s)...\n"
	MsgProbeOkFmt       = "  ✓ %s → %s (method=%s)\n"
	MsgProbeNoneFmt     = "  · %s → no new version (method=%s)\n"
	MsgProbeFailFmt     = "  ✗ %s → %s\n"
	MsgProbeDoneFmt     = "✓ Probe complete: %d available, %d unchanged, %d failed.\n"
	MsgProbeUsageHeader = "Usage: gitmap probe [<repo-path>|--all]"
	MsgProbeNoTargets   = "No repos to probe. Pass a path or --all.\n"
)

// VersionProbe CLI tokens.
const (
	ProbeFlagAll  = "--all"
	ProbeFlagJSON = "--json"
)
