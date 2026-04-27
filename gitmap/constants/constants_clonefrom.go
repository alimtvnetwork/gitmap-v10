package constants

// Constants for `gitmap clone-from <file>` (v3.160.0+).
//
// All user-facing strings live here per the no-magic-strings rule
// (mem://core). Format strings use printf-style verbs and document
// each substitution in a comment so future translators / log
// parsers can map %v back to call sites without re-reading the
// emit code.

// CLI surface. CmdCloneFrom and CmdCloneFromAlias are referenced
// from the dispatcher (rootcore.go) and the completion generator
// (which scans the const block for the `// gitmap:cmd top-level`
// marker — preserve that comment when reorganizing).
const (
	CmdCloneFrom      = "clone-from"
	CmdCloneFromAlias = "cf"
)

// Flag names + descriptions. Flag names are the long form only;
// short flags are deliberately avoided here because `-e` could
// collide with future global flags (the existing flag conventions
// in gitmap reserve single-letter flags for very-frequent
// operations).
const (
	FlagCloneFromExecute      = "execute"
	FlagDescCloneFromExecute  = "Actually run git clone (default: dry-run only)"
	FlagCloneFromQuiet        = "quiet"
	FlagDescCloneFromQuiet    = "Suppress per-row progress lines (summary still prints)"
	FlagCloneFromNoReport     = "no-report"
	FlagDescCloneFromNoReport = "Skip writing the .gitmap/clone-from-report-*.csv file"
	// FlagCloneFromOutput selects the dry-run / per-row format.
	// "default" preserves the legacy 4-line block (url/dest/branch/
	// depth). "terminal" emits the standardized RepoTermBlock used
	// across scan/clone-next/probe so users get one format regardless
	// of which command produced the per-repo summary. Wording mirrors
	// FlagDescCloneTermOutput so the stdout-vs-stderr split is the
	// same sentence across every clone command.
	FlagCloneFromOutput     = "output"
	FlagDescCloneFromOutput = "Per-row format: 'default' (legacy 4-line block) " +
		"or 'terminal' (standardized branch/from/to/command block on " +
		"stdout, streamed before each clone; git progress stays on stderr)"
)

// JSON report envelope. CloneFromReportSchemaVersion is embedded as
// `schemaVersion` at the top of every JSON report so downstream
// consumers (jq pipelines, dashboards, custom CI gates) can branch
// on shape changes without sniffing fields. Bump ONLY when the
// `rows[]` field set, field rename, or envelope shape changes —
// not for value-level changes (new status strings, etc.). The
// matching golden test (TestCloneFromReportJSON_SchemaVersion_Pinned)
// will fail loudly on any unintentional bump.
const CloneFromReportSchemaVersion = 1

// Status enum strings. Stable: emitted to the CSV report which
// downstream tools may parse. Renaming any of these is a breaking
// change for those consumers.
const (
	CloneFromStatusOK      = "ok"
	CloneFromStatusSkipped = "skipped"
	CloneFromStatusFailed  = "failed"
)

// CloneFromErrTrimLimit caps the per-row stderr summary length.
// 80 chars fits a typical terminal column; longer messages are
// truncated with an ellipsis. Full stderr remains in the user's
// scrollback (we use CombinedOutput so it was printed live).
const CloneFromErrTrimLimit = 80

// User-facing messages. Format strings include trailing newlines
// so the call site doesn't need to remember to add them. Header
// formats end in a blank line to visually separate from the
// per-row block.
const (
	// %s = source path, %s = format ("json"|"csv"), %d = row count.
	MsgCloneFromDryHeader = "gitmap clone-from: dry-run\n" +
		"source: %s (%s)\n" +
		"%d row(s) -- pass --execute to actually clone\n\n"
	// %d ok, %d skipped, %d failed, %d total.
	MsgCloneFromSummaryHeader = "\ngitmap clone-from: %d ok, %d skipped, %d failed (%d total)\n"
	MsgCloneFromDestExists    = "dest exists"
	MsgCloneFromMissingArg    = "clone-from: <file> argument is required (e.g. clone-from repos.csv)"
)

// Errors. All use printf-style verbs documented inline.
const (
	// %s = path, %v = err.
	ErrCloneFromAbsPath = "clone-from: resolve path %s: %v"
	ErrCloneFromOpen    = "clone-from: open %s: %v"
	// %v = err.
	ErrCloneFromJSONDecode = "clone-from: decode JSON: %v"
	// %d = 1-indexed row, %v = err.
	ErrCloneFromJSONRow = "clone-from: row %d: %v"
	// %v = err.
	ErrCloneFromCSVHeader = "clone-from: read CSV header: %v"
	ErrCloneFromCSVNoURL  = "clone-from: CSV header is missing required column 'url'"
	// %d = row number including header, %v = err.
	ErrCloneFromCSVRow = "clone-from: CSV row %d: %v"
	// %s = bad depth string.
	ErrCloneFromBadDepth = "depth %q is not a valid integer"
	ErrCloneFromEmptyURL = "url is empty after trim"
	// %s = url.
	ErrCloneFromBadURL = "url %q does not look like a git URL " +
		"(expected https://, http://, ssh://, git://, file://, or scp-style host:path)"
	// %d = bad depth.
	ErrCloneFromNegDepth = "depth %d is negative"

	// CloneFromDepthFlagFmt is the SINGLE source of truth for how
	// clone-from renders its shallow-clone flag, both in the executed
	// argv (clonefrom/execute.go buildGitArgs) and in every
	// human-facing preview (render.go cloneCommandForRow + cmd/
	// clonetermrow.go printCloneFromTermBlockRow). The joined form
	// (`--depth=N`) is intentional — it matches what the executor
	// hands to exec.Command exactly, so the printed cmd: line in the
	// terminal block is byte-faithful and the --verify-cmd-faithful
	// checker has zero false positives. Do NOT switch to the split
	// form (`--depth N`) without updating ALL three sites + the
	// golden fixture in cmd/testdata/clonetermblock_clonefrom.golden
	// + TestCloneFromDepthFlagFormat_Locked.
	// %d = depth.
	CloneFromDepthFlagFmt = "--depth=%d"
	// %s = directory, %v = err.
	ErrCloneFromReportMkdir = "clone-from: mkdir %s: %v"
	// %s = file path, %v = err.
	ErrCloneFromReportCreate = "clone-from: create report %s: %v"
)
