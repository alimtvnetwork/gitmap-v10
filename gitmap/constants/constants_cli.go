package constants

// gitmap:cmd top-level
// CLI commands.
const (
	CmdScan                  = "scan"
	CmdScanAlias             = "s"
	CmdClone                 = "clone"
	CmdCloneAlias            = "c"
	CmdUpdate                = "update"
	CmdUpdateRunner          = "update-runner" // gitmap:cmd skip
	CmdUpdateCleanup         = "update-cleanup" // gitmap:cmd skip
	CmdInstalledDir          = "installed-dir" // gitmap:cmd skip
	CmdInstalledDirAlias     = "id"
	CmdVersion               = "version"
	CmdVersionAlias          = "v"
	CmdHelp                  = "help"
	CmdDesktopSync           = "desktop-sync"
	CmdDesktopSyncAlias      = "ds"
	CmdGitHubDesktop         = "github-desktop"
	CmdGitHubDesktopAlias    = "gd"
	CmdPull                  = "pull"
	CmdPullAlias             = "p"
	CmdRescan                = "rescan"
	CmdRescanAlias           = "rsc"
	CmdSetup                 = "setup"
	CmdStatus                = "status"
	CmdStatusAlias           = "st"
	CmdExec                  = "exec"
	CmdExecAlias             = "x"
	CmdRelease               = "release"
	CmdReleaseShort          = "r"
	CmdReleaseBranch         = "release-branch"
	CmdReleaseBranchAlias    = "rb"
	CmdReleasePending        = "release-pending"
	CmdReleasePendingAlias   = "rp"
	CmdChangelog             = "changelog"
	CmdChangelogAlias        = "cl"
	CmdChangelogMD           = "changelog.md" // gitmap:cmd skip
	CmdDoctor                = "doctor"
	CmdLatestBranch          = "latest-branch"
	CmdLatestBranchAlias     = "lb"
	CmdList                  = "list"
	CmdListAlias             = "ls"
	CmdGroup                 = "group"
	CmdGroupAlias            = "g"
	CmdGroupCreate           = "create" // gitmap:cmd skip
	CmdGroupAdd              = "add" // gitmap:cmd skip
	CmdGroupRemove           = "remove" // gitmap:cmd skip
	CmdGroupList             = "list" // gitmap:cmd skip
	CmdGroupShow             = "show" // gitmap:cmd skip
	CmdGroupDelete           = "delete" // gitmap:cmd skip
	CmdDBReset               = "db-reset"
	CmdReset                 = "reset"
	CmdListVersions          = "list-versions"
	CmdListVersionsAlias     = "lv"
	CmdRevert                = "revert"
	CmdRevertRunner          = "revert-runner" // gitmap:cmd skip
	CmdListReleases          = "list-releases"
	CmdListReleasesAlias     = "lr"
	CmdReleases              = "releases" // v3.20.0: alias of list-releases, intended for --all-repos batch view
	CmdCompletion            = "completion"
	CmdCompletionAlias       = "cmp"
	CmdClearReleaseJSON      = "clear-release-json"
	CmdClearReleaseJSONAlias = "crj"
	CmdDocs                  = "docs"
	CmdDocsAlias             = "d"
	CmdCloneNext             = "clone-next"
	CmdCloneNextAlias        = "cn"
	CmdReleaseSelf           = "release-self"
	CmdReleaseSelfAlias      = "rself"
	CmdReleaseSelfAlias2     = "rs"
	CmdHelpDashboard         = "help-dashboard"
	CmdHelpDashboardAlias    = "hd"
	CmdPending               = "pending" // gitmap:cmd skip
	CmdDoPending             = "do-pending" // gitmap:cmd skip
	CmdDoPendingAlias        = "dp" // gitmap:cmd skip
	CmdLLMDocs               = "llm-docs"
	CmdLLMDocsAlias          = "ld"
	CmdSetSourceRepo         = "set-source-repo" // gitmap:cmd skip
	CmdSelfInstall           = "self-install"
	CmdSelfUninstall         = "self-uninstall"
	CmdSf                    = "sf"
	CmdProbe                 = "probe"
	CmdCode                  = "code"
	CmdVSCodePMPath          = "vscode-pm-path"
	CmdVSCodePMPathAlias     = "vpath"
	CmdLFSCommon             = "lfs-common"
	CmdLFSCommonAlias        = "lfsc"
	CmdReplace               = "replace"
	CmdReplaceAlias          = "rpl"
	CmdInject                = "inject"
	CmdInjectAlias           = "inj"
)

// Usage header.
const UsageHeaderFmt = "gitmap v%s\n\n"

const (
	HelpUsage            = "Usage: gitmap <command> [flags]"
	HelpCommands         = "Commands:"
	HelpScan             = "  scan (s) [dir]      Scan directory for Git repos"
	HelpClone            = "  clone (c) <source|json|csv|text>  Re-clone from file (shorthands auto-resolve)"
	HelpUpdate           = "  update              Self-update from source repo"
	HelpUpdateCleanup    = "  update-cleanup      Remove leftover update temp files and .old backups"
	HelpInstalledDir     = "  installed-dir (id)  Show the active installed binary path"
	HelpVersion          = "  version (v)         Show version number"
	HelpDesktopSync      = "  desktop-sync (ds)   Sync repos to GitHub Desktop from output"
	HelpGitHubDesktop    = "  github-desktop (gd) Register current repo with GitHub Desktop (no scan needed)"
	HelpPull             = "  pull (p) <name>     Pull a specific repo by its name"
	HelpRescan           = "  rescan (rsc)        Re-run last scan with cached flags"
	HelpSetup            = "  setup               Configure Git diff/merge tool, aliases & core settings"
	HelpStatus           = "  status (st)         Show dirty/clean, ahead/behind, stash for all repos"
	HelpExec             = "  exec (x) <args...>  Run any git command across all repos"
	HelpRelease          = "  release (r) [ver]   Create release branch, tag, and push"
	HelpReleaseBr        = "  release-branch (rb) Complete release from existing release branch"
	HelpReleasePend      = "  release-pending (rp) Release all pending branches without tags"
	HelpChangelog        = "  changelog (cl) [ver] Show concise release notes (use --open, --source)"
	HelpDoctor           = "  doctor [--fix-path] Diagnose PATH, deploy, and version issues"
	HelpLatestBr         = "  latest-branch (lb)  Find most recently updated remote branch"
	HelpList             = "  list (ls)           Show all tracked repos with slugs"
	HelpGroup            = "  group (g) <sub>     Manage repo groups / activate group for batch ops"
	HelpMultiGroup       = "  multi-group (mg)    Select multiple groups for batch operations"
	HelpSf               = "  sf <add|list|rm>    Manage scan folders (roots that gitmap scan tracks)"
	HelpDBReset          = "  db-reset --confirm  Clear all tracked repos and groups from the database"
	HelpCompletion       = "  completion (cmp)    Generate shell tab-completion scripts"
	HelpClearReleaseJSON = "  clear-release-json (crj)  Remove a .gitmap/release/vX.Y.Z.json file"
	HelpDocs             = "  docs (d)            Open documentation website in browser"
	HelpHelpDash         = "  help-dashboard (hd) Serve the docs site locally in your browser"
	HelpCloneNext        = "  clone-next (cn)     Clone next versioned iteration of current repo"
	HelpReleaseSelf      = "  release-self (rs)   Release gitmap itself from any directory"
	HelpHelp             = "  help                Show this help message"
	HelpListVersions     = "  list-versions (lv)  Show all release tags, highest first (--limit N, --json, --source)"
	HelpListReleases     = "  list-releases (lr)  Show releases from .gitmap/release/ files or database (--limit N, --json, --source)"
	HelpRevert           = "  revert <version>    Revert to a specific release version"
	MsgHelpLFSCommon     = "  lfs-common (lfsc)   Track common binary file types with Git LFS in current repo"
)

// Help section headers and flag-line strings (HelpScanFlags, HelpCloneFlags,
// HelpReleaseFlags, ...) live in constants_helpsections.go.

// Flag descriptions.
const (
	FlagDescConfig        = "Path to config file"
	FlagDescMode          = "Clone URL style: https or ssh"
	FlagDescOutput        = "Output format: terminal, csv, json"
	FlagDescOutFile       = "Exact output file path"
	FlagDescOutputPath    = "Output directory for CSV/JSON"
	FlagDescTargetDir     = "Base directory for cloned repos"
	FlagDescSafePull      = "If repo exists, run safe git pull with retries and unlock diagnostics"
	FlagDescGHDesktop     = "Add discovered repos to GitHub Desktop"
	FlagDescOpen          = "Open output folder after scan completes"
	FlagDescQuiet         = "Suppress terminal clone help section"
	FlagDescVerbose       = "Write detailed stdout/stderr debug log to a timestamped file"
	FlagScanWorkers       = "workers"
	FlagDescScanWorkers   = "Worker pool size for scan (0 = auto, capped at 16)"
	DefaultScanWorkers    = 0
	// FlagScanRelativeRoot lets the user pin the base path used to compute
	// each repo's RelativePath in the output (CSV/JSON/text/structure/
	// clone scripts). Without it, RelativePath is derived from the scan
	// dir, which means running `gitmap scan .` from different cwds
	// produces different paths for the same repos. With --relative-root,
	// every output row is computed against the supplied (absolute or
	// relative) directory so artifacts stay byte-stable across cwds.
	FlagScanRelativeRoot     = "relative-root"
	FlagDescScanRelativeRoot = "Pin the base directory used for output RelativePath (absolute or relative; must contain every scanned repo)"
	// FlagScanMaxDepth caps how many directory levels below the scan root
	// the walker may descend. Default 0 → resolves to scanner.DefaultMaxDepth
	// (4) inside the scanner. Negative → unbounded (legacy behavior).
	// Honored regardless of whether a `.git` was found on the path —
	// repos already stop their own subtree as before.
	FlagScanMaxDepth     = "max-depth"
	FlagDescScanMaxDepth = "Max directory levels to descend below scan root (0 = default 4, negative = unlimited)"
	DefaultScanMaxDepth  = 0
	// FlagScanDefaultBranch overrides the fallback branch name written
	// to ScanRecord.Branch when none of the live detection steps in
	// gitutil.DetectBranchWithDefault returned a usable name. Without
	// the flag, the fallback is the package-default constants.DefaultBranch
	// ("main"). Useful for catalogs that target legacy infra still on
	// "master", or for forcing a project-specific convention without
	// touching the binary's compiled-in default.
	FlagScanDefaultBranch     = "default-branch"
	FlagDescScanDefaultBranch = "Fallback branch name when HEAD/remote-tracking detection finds nothing (default: main)"
	// FlagScanReportErrors enables a JSON failure report at command
	// exit for `gitmap scan` and `gitmap cn --all/--csv`. Bare boolean
	// — output path is fixed at `<binaryDir>/.gitmap/reports/errors-
	// <unixts>.json`. Honored by both `gitmap scan` (captures scanner
	// ReadDir + background ls-remote / shallow-clone failures) and
	// `gitmap cn --all/--csv` (captures per-repo clone failures).
	// Clean runs leave NO file on disk.
	//
	// NOTE: distinct from the existing `--report-errors` (with leading
	// dashes baked in) on the `gitmap update` command — that flag
	// takes a value (`json`) and writes a JSONL handoff trace. Naming
	// this one `--errors-report` keeps both UX surfaces intact.
	FlagScanReportErrors     = "errors-report"
	FlagDescScanReportErrors = "Write per-repo failures to .gitmap/reports/errors-<unixts>.json (only emitted when failures occur)"
	FlagDescSetupConfig   = "Path to git-setup.json config file"
	FlagDescDryRun        = "Preview changes without applying them"
	FlagDescAssets        = "Directory or file to attach to the release"
	FlagDescCommit        = "Create release from a specific commit"
	FlagDescRelBranch     = "Create release from latest commit of a branch"
	FlagDescBump          = "Auto-increment version: major, minor, or patch"
	FlagDescDraft         = "Create an unpublished draft release"
	FlagDescLatest        = "Show only the latest changelog entry"
	FlagDescLimit         = "Number of changelog versions to show"
	FlagDescOpenChangelog = "Open CHANGELOG.md with the default system app"
	FlagDescLBRemote      = "Remote to filter branches against (default: origin)"
	FlagDescLBAllRemotes  = "Include branches from all remotes"
	FlagDescLBContains    = "Fall back to --contains if --points-at returns empty"
	FlagDescLBTop         = "Show top N most recently updated branches"
	FlagDescLBJSON        = "Output structured JSON instead of plain text (shorthand for --format json)"
	FlagDescLBFormat      = "Output format: terminal, json, csv (default: terminal)"
	FlagDescLBNoFetch     = "Skip git fetch (use existing remote refs)"
	FlagDescLBSort        = "Sort order: date (default, descending) or name (alphabetical)"
	FlagDescLBFilter      = "Filter branches by glob or substring pattern"
	// FlagDescLBSwitch / FlagDescLBSwitchShort document `gitmap lb
	// --switch` / `gitmap lb -s`. Both register the SAME effect: after
	// printing the latest-branch report, run `git checkout` against the
	// resolved branch name. Useful as a single-keystroke "jump to whatever
	// is freshest" workflow when reviewing PR work across many branches.
	FlagDescLBSwitch      = "Checkout the resolved latest branch after printing the report"
	FlagDescLBSwitchShort = "Short alias for --switch"
	FlagDescGroup         = "Filter by group name"
	FlagDescAll           = "Run against all tracked repos from database"
	FlagDescListVerbose   = "Show full paths and URLs"
	FlagDescGroupDesc     = "Optional group description"
	FlagDescGroupColor    = "Terminal color for group display"
	FlagDescConfirm       = "Confirm destructive operation"
	FlagDescSource        = "Filter by source: release or import"
)
