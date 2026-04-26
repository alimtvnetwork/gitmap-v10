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

// Background probe tuning for `gitmap scan` (v3.123.0+).
//
// When scan finds a small repo set we eagerly kick off a probe pass in
// the background so the next `gitmap find-next` call already has fresh
// data. The defaults are intentionally gentle: 3 workers max so we do
// not hammer GitHub's rate limit, and the auto-trigger only fires for
// scans of <50 repos so a directory full of vendored sources doesn't
// suddenly fan out 500 ls-remote calls. Power users can override or
// disable any of this with flags below.
const (
	// ScanProbeFlagDisable disables the background probe entirely.
	ScanProbeFlagDisable = "no-probe"
	// ScanProbeFlagNoWait makes scan return immediately after kicking
	// off probes (they keep running in background until completion or
	// process exit).
	ScanProbeFlagNoWait = "no-probe-wait"
	// ScanProbeFlagConcurrency sets the worker-pool size for the
	// background probe runner.
	ScanProbeFlagConcurrency = "probe-concurrency"

	// ScanProbeDefaultConcurrency caps the background pool. Three
	// workers is the documented sweet spot: parallel enough to
	// finish 50 probes in ~ten seconds, low enough that GitHub's
	// abuse detection doesn't kick in.
	ScanProbeDefaultConcurrency = 3
	// ScanProbeAutoTriggerCeiling is the repo-count threshold under
	// which background probing fires automatically. Above it the
	// user must opt in by passing --probe-concurrency explicitly.
	ScanProbeAutoTriggerCeiling = 50

	// FlagDescScanProbeDisable, FlagDescScanProbeNoWait, and
	// FlagDescScanProbeConcurrency are the help strings shown by
	// `gitmap help scan`.
	FlagDescScanProbeDisable     = "Skip the background version probe entirely (offline / air-gapped runs)"
	FlagDescScanProbeNoWait      = "Return as soon as scan finishes; let probes keep running in the background"
	FlagDescScanProbeConcurrency = "Worker count for the background probe pool (default 3, 0 = disable)"
)

// Background probe runtime messages.
const (
	// MsgScanProbeStartFmt fires once when scan kicks off the
	// background runner. The %d is the number of repos queued.
	MsgScanProbeStartFmt = "  ↪ background probe queued for %d repo(s) (workers=%d)\n"
	// MsgScanProbeWaitingFmt is printed at scan-end while we block
	// on the runner's Wait. Includes the count remaining so users
	// can gauge how long they'll be waiting.
	MsgScanProbeWaitingFmt = "  ⏳ Waiting for background probes to finish (%d remaining)...\n"
	// MsgScanProbeDoneFmt prints the per-bucket tally once Wait
	// returns. Mirrors the foreground probe summary line.
	MsgScanProbeDoneFmt = "  ✓ Background probe done: %d available, %d unchanged, %d failed.\n"
	// MsgScanProbeDetached prints when --no-probe-wait was passed
	// and we're returning before the pool drains.
	MsgScanProbeDetached = "  ↪ Background probe detached; results will land in the DB asynchronously.\n"
	// MsgScanProbeSkippedAutoFmt prints when the auto-trigger
	// declined to start (repo count above the ceiling) so users
	// understand why no probe ran.
	MsgScanProbeSkippedAutoFmt = "  · Background probe skipped: %d repos exceeds auto-trigger ceiling (%d). Pass --probe-concurrency to force.\n"
)
