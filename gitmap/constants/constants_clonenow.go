package constants

// Constants for `gitmap clone-now <file>` (v3.161.0+).
//
// `clone-now` re-runs `git clone` against scan output: it consumes the
// JSON / CSV / text artifacts produced by `gitmap scan` (the same
// files written under `.gitmap/output/`) and re-creates each repo at
// its recorded relative path using the user-selected URL mode.
//
// Why a separate command from clone-from?
//
//   - clone-from is plan-driven (user-authored row schema: url, dest,
//     branch, depth) — the file describes intent.
//   - clone-now is round-trip-driven: input is gitmap's own scan
//     output, and we honor the recorded RelativePath verbatim so the
//     destination tree is byte-identical to the original layout.
//
// All user-facing strings live here per the no-magic-strings rule.

// CLI surface — canonical name + backward-compat aliases.
//
// CmdCloneReclone ("reclone") is the CANONICAL verb as of v3.x.
// It was renamed from clone-now so the CLI vocabulary makes the
// split between the two clone families unambiguous:
//
//   - `gitmap clone <url>`   — fresh clone from a URL.
//   - `gitmap reclone <file>` — RE-clone from a scan artifact
//     (round-trip the recorded RelativePath layout).
//
// All four legacy spellings are kept as aliases so existing scripts
// keep working forever:
//   - clone-now / cnow  — original name (v3.161.0).
//   - relclone / rc     — earlier explicit "re-clone" alias.
//
// The dispatcher in rootcore.go binds every spelling to runCloneNow.
// The completion generator picks them up via the
// `// gitmap:cmd top-level` marker on this const block.
const (
	CmdCloneReclone      = "reclone"
	CmdCloneRecloneAlias = "rec"
	CmdCloneNow          = "clone-now"
	CmdCloneNowAlias     = "cnow"
	CmdCloneRel          = "relclone"
	CmdCloneRelAlias     = "rc"
)

// Flag names + descriptions. Long-form only; short flags are
// reserved for very-frequent operations per the project convention.
const (
	FlagCloneNowExecute     = "execute"
	FlagDescCloneNowExecute = "Actually run git clone (default: dry-run only)"
	FlagCloneNowQuiet       = "quiet"
	FlagDescCloneNowQuiet   = "Suppress per-row progress lines (summary still prints)"
	// FlagCloneNowMode picks which URL column to use when the input
	// supplies both. Values: "https" (default) | "ssh". When the
	// requested mode is missing on a given row we fall back to the
	// other one rather than skipping the row -- the user's intent is
	// "clone these repos now", not "clone only the ones that have
	// the preferred URL shape".
	FlagCloneNowMode     = "mode"
	FlagDescCloneNowMode = "URL mode to clone with: 'https' (default) or 'ssh'. " +
		"Falls back to the other mode if the preferred URL is missing on a row."
	// FlagCloneNowFormat lets the caller force the input format when
	// the file extension is missing or wrong (e.g., `repos.out`).
	// Values: "" (auto from extension) | "json" | "csv" | "text".
	FlagCloneNowFormat     = "format"
	FlagDescCloneNowFormat = "Force input format: '' (auto from extension), 'json', 'csv', or 'text'."
	// FlagCloneNowCwd sets the working directory clones run in.
	// Empty (default) = the current process cwd. Honored as-is so
	// that scripts can re-create a tree under a fresh root with
	// `gitmap clone-now scan.json --cwd ./mirror --execute`.
	FlagCloneNowCwd     = "cwd"
	FlagDescCloneNowCwd = "Working directory for git clone (default: current dir)."
	// FlagCloneNowOnExists controls re-clone behavior when the
	// destination already contains a git repo. Default "skip" keeps
	// the historical no-op behavior. "update" runs fetch + checkout
	// without destroying local work. "force" removes the directory
	// and re-clones from scratch -- destructive, opt-in only.
	FlagCloneNowOnExists     = "on-exists"
	FlagDescCloneNowOnExists = "Behavior when target already exists: " +
		"'skip' (default, no-op when repo+branch match), " +
		"'update' (fetch + checkout to align with the planned URL/branch), " +
		"'force' (remove target and re-clone -- destructive)."
	// FlagCloneNowManifest is the unified, explicit-path alias for the
	// positional <file> argument. Accepts a JSON or CSV produced by
	// `gitmap scan`; format is auto-detected from the extension (use
	// --format to override). Precedence:
	//   1. --manifest <path>     (explicit, highest priority)
	//   2. positional <file>     (legacy positional form)
	//   3. auto-pickup           (./.gitmap/output/gitmap.{json,csv})
	// Passing BOTH --manifest and a positional file is a usage error
	// (exit 2) so the run is unambiguous.
	FlagCloneNowManifest     = "manifest"
	FlagDescCloneNowManifest = "Path to a scan artifact (JSON or CSV) to consume. " +
		"Format auto-detected from the extension. Equivalent to the positional " +
		"<file> argument; when omitted, auto-pickup under ./.gitmap/output/ is used."
)

// On-exists policy enum strings. Stable: surfaced in --on-exists,
// the dry-run header, and the per-row Result.Detail field. Renaming
// is a breaking change for shell scripts that grep these values.
const (
	CloneNowOnExistsSkip   = "skip"
	CloneNowOnExistsUpdate = "update"
	CloneNowOnExistsForce  = "force"
)

// Mode enum strings. Stable: surfaced in the dry-run header and the
// per-row progress lines.
const (
	CloneNowModeHTTPS = "https"
	CloneNowModeSSH   = "ssh"
)

// Format enum strings. The "auto" empty value is intentionally not
// exported because callers detect it via len(format)==0.
const (
	CloneNowFormatJSON = "json"
	CloneNowFormatCSV  = "csv"
	CloneNowFormatText = "text"
)

// Status enum strings. Mirrors clone-from for cross-tool grep-ability:
// downstream pipelines that already filter on "ok"/"skipped"/"failed"
// keep working without a status-name translation table.
const (
	CloneNowStatusOK      = "ok"
	CloneNowStatusSkipped = "skipped"
	CloneNowStatusFailed  = "failed"
)

// CloneNowErrTrimLimit caps the per-row stderr summary length so the
// summary table stays scannable in an 80-column terminal. Full stderr
// remains in the user's scrollback because we use CombinedOutput.
const CloneNowErrTrimLimit = 80

// User-facing messages. Trailing newlines are baked in so call sites
// don't need to remember them.
const (
	// %s = source path, %s = format, %s = mode, %d = row count.
	MsgCloneNowDryHeader = "gitmap clone-now: dry-run\n" +
		"source: %s (%s, mode=%s)\n" +
		"%d row(s) -- pass --execute to actually clone\n\n"
	// %d ok, %d skipped, %d failed, %d total.
	MsgCloneNowSummaryHeader = "\ngitmap clone-now: %d ok, %d skipped, %d failed (%d total)\n"
	MsgCloneNowDestExists    = "dest exists"
	MsgCloneNowMissingArg = "reclone: <file> argument is required and " +
		"no scan artifact was found under ./.gitmap/output/ " +
		"(looked for gitmap.json then gitmap.csv). " +
		"Run `gitmap scan` first, or pass an explicit path " +
		"(e.g. reclone .gitmap/output/gitmap.json)."
	// %s = auto-discovered manifest path. Printed to stderr when
	// reclone is invoked with no <file> arg and a scan artifact is
	// found in the conventional location. Lets users see exactly
	// which file fed the run instead of guessing.
	MsgCloneNowAutoPickup = "reclone: using scan artifact %s (auto-discovered; pass an explicit path to override)\n"
	// %s = positional file, %s = --manifest value. Printed when the
	// caller supplies BOTH forms — refusing is safer than silently
	// preferring one and having the run consume the wrong artifact.
	MsgCloneNowManifestConflict = "reclone: cannot combine positional <file> %q with --manifest %q; pass only one\n"
	MsgCloneNowNoURL = "no url for selected mode"
	// Idempotency / re-clone messages. Each lands in Result.Detail
	// so the per-row summary tells the user exactly which branch
	// of the on-exists policy fired. The mismatch + fail messages
	// take printf args -- documented per-line.
	MsgCloneNowAlreadyMatches = "already matches (url + branch)"
	MsgCloneNowNotARepo       = "dest exists but is not a git repo"
	// %s = current remote, %s = expected remote.
	MsgCloneNowURLMismatch = "skipped: remote url differs (have=%s, want=%s)"
	// %s = current branch, %s = expected branch.
	MsgCloneNowBranchMismatch = "skipped: branch differs (have=%s, want=%s)"
	MsgCloneNowUpdated        = "updated (fetch + checkout)"
	MsgCloneNowForceRecloned  = "force-recloned (previous dir removed)"
	// %s = path, %v = err.
	MsgCloneNowForceRemoveFail = "force: remove %s: %v"
	// %s = trimmed git stderr.
	MsgCloneNowFetchFail = "update: git fetch failed: %s"
	// %s = branch, %s = trimmed git stderr.
	MsgCloneNowCheckoutFail = "update: git checkout %s failed: %s"
)

// Errors. All use printf-style verbs documented inline.
const (
	// %s = path, %v = err.
	ErrCloneNowAbsPath = "clone-now: resolve path %s: %v"
	ErrCloneNowOpen    = "clone-now: open %s: %v"
	// %v = err.
	ErrCloneNowJSONDecode = "clone-now: decode JSON: %v"
	ErrCloneNowCSVRead    = "clone-now: read CSV: %v"
	ErrCloneNowTextRead   = "clone-now: read text: %v"
	// %s = bad value.
	ErrCloneNowBadMode     = "clone-now: --mode must be 'https' or 'ssh', got %q"
	ErrCloneNowBadFormat   = "clone-now: --format must be 'json', 'csv', or 'text', got %q"
	ErrCloneNowBadOnExists = "clone-now: --on-exists must be 'skip', 'update', or 'force', got %q"
	// %s = file extension (with leading dot, or "" when missing),
	// %s = path. Emitted when auto-detect cannot map the file
	// extension to a supported format. Use --format to override.
	ErrCloneNowUnsupportedExt = "clone-now: unsupported file extension %q for %s; " +
		"supported extensions are .json, .csv, .txt (or pass --format json|csv|text)"
	// %s = path.
	ErrCloneNowEmpty = "clone-now: %s contains zero clonable rows"
	// MsgCloneNowMkdirParentFailFmt is the per-row Detail set when
	// pre-creating the destination's parent directory fails. Mirrors
	// the clonefrom equivalent so summary tables read consistently.
	// %v = err.
	MsgCloneNowMkdirParentFailFmt = "mkdir parent: %v"
	// ErrCloneNowMkdirParent is the standardized stderr log emitted
	// alongside MsgCloneNowMkdirParentFailFmt (Code Red zero-swallow).
	// %s = parent path, %v = err.
	ErrCloneNowMkdirParent = "Error: clone-now: failed to create dest parent at %s: %v " +
		"(check permissions / disk space)\n"
)
