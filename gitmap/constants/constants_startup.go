// Package constants — Linux/Unix startup-management commands and the
// marker convention used to scope list/remove operations to entries
// gitmap itself created (never third-party autostart files).
package constants

// gitmap:cmd top-level
// Startup CLI commands. The pair `startup-list` / `startup-remove`
// is intentionally limited in scope on Linux/Unix: it ONLY enumerates
// and deletes entries gitmap created itself, identified by the
// `X-Gitmap-Managed=true` key inside the .desktop file. This matches
// the user requirement that no third-party autostart entry is ever
// touched, even if its filename happens to start with "gitmap".
//
// Aliases mirror the install/uninstall short-form pattern so users
// who already know `in`/`un` can guess `sl`/`sr` without lookup.
const (
	CmdStartupAdd         = "startup-add"
	CmdStartupAddAlias    = "sa"
	CmdStartupList        = "startup-list"
	CmdStartupListAlias   = "sl"
	CmdStartupRemove      = "startup-remove"
	CmdStartupRemoveAlias = "sr"
)

// Startup help text. Format mirrors HelpInstall / HelpUninstall so
// the rendered `gitmap help` table stays visually consistent.
const (
	HelpStartupAdd    = "  startup-add (sa)          Create a Linux/Unix or macOS autostart entry pointing at gitmap"
	HelpStartupList   = "  startup-list (sl)         List Linux/Unix autostart entries created by gitmap"
	HelpStartupRemove = "  startup-remove (sr) <name> Remove a gitmap-managed autostart entry by name"
)

// Startup .desktop file convention.
//
// StartupDesktopExt is the file extension XDG autostart entries MUST
// use; the directory scanner only considers files ending in this
// suffix to avoid matching swap / backup files.
//
// StartupMarkerKey is the .desktop key whose presence with value
// "true" marks an entry as gitmap-managed. We deliberately use a
// custom `X-` namespaced key (per the freedesktop.org spec — keys
// prefixed with `X-` are reserved for application-specific use) so a
// future tool with the same key cannot collide with us.
//
// StartupFilePrefix is the filename prefix every gitmap-created
// .desktop file gets. It's a SECONDARY safeguard — list/remove still
// require StartupMarkerKey to be present, so renaming a third-party
// file to start with "gitmap-" does NOT make us touch it.
const (
	StartupDesktopExt = ".desktop"
	StartupMarkerKey  = "X-Gitmap-Managed"
	StartupMarkerVal  = "true"
	StartupFilePrefix = "gitmap-"
)

// macOS LaunchAgent convention. The marker uses a hyphen-free key
// (`XGitmapManaged`) because plist keys are XML element text and
// keeping them simple avoids any escaping ambiguity in tests / hand
// edits. The on-disk filename prefix is `gitmap.` (reverse-DNS-ish)
// so .plist files match the LaunchAgent convention while still
// being cheap to pre-filter — same two-gate safety as Linux: the
// filename prefix is a hint, the in-file marker is the proof.
const (
	StartupPlistExt    = ".plist"
	StartupPlistMarker = "XGitmapManaged"
	StartupPlistPrefix = "gitmap."
)

// Startup user-visible messages. Plain ASCII arrows / glyphs to stay
// safe across terminals that don't render Unicode (matches the
// PowerShell-encoding constraint documented in mem://constraints/
// powershell-encoding even though this is a Linux command — keeps
// the message style uniform across the whole CLI).
const (
	MsgStartupListHeader     = "Linux/Unix autostart entries managed by gitmap (%s):\n"
	MsgStartupListEmpty      = "  (none — no gitmap-managed autostart entries found)\n"
	MsgStartupListRow        = "  • %s  →  %s\n"
	MsgStartupListFooter     = "\nTotal: %d entry(ies). Remove one with: gitmap startup-remove <name>\n"
	MsgStartupRemoveOK       = "✓ Removed gitmap-managed autostart entry: %s\n"
	MsgStartupRemoveNoOp     = "  (no-op) no gitmap-managed entry named %q found\n"
	MsgStartupRemoveNotOurs  = "  (refused) %q exists but was NOT created by gitmap — skipping\n"
	MsgStartupRemoveBadName  = "  (refused) name %q is empty or contains a path separator\n"
	// startup-remove --dry-run mirror messages. Same four outcomes,
	// each prefixed with `(dry-run)` so log-scrapers can tell a
	// preview from a real action without parsing flags.
	MsgStartupRemoveDryOK      = "  (dry-run) would remove gitmap-managed autostart entry: %s\n"
	MsgStartupRemoveDryNoOp    = "  (dry-run) no gitmap-managed entry named %q found — nothing to remove\n"
	MsgStartupRemoveDryNotOurs = "  (dry-run) %q exists but was NOT created by gitmap — would refuse\n"
	MsgStartupRemoveDryBadName = "  (dry-run) name %q is empty or contains a path separator — would refuse\n"
	// startup-add result messages. One line per outcome so log
	// scrapers can grep on the leading symbol/prefix.
	MsgStartupAddCreated     = "✓ Created gitmap-managed autostart entry: %s\n"
	MsgStartupAddOverwritten = "✓ Overwrote gitmap-managed autostart entry: %s\n"
	MsgStartupAddExists      = "  (exists) gitmap-managed entry already at %s — pass --force to overwrite\n"
	MsgStartupAddRefused     = "  (refused) %q exists but was NOT created by gitmap — refusing to overwrite\n"
	MsgStartupAddBadName     = "  (refused) name %q is empty or contains a path separator\n"
	ErrStartupResolveDir     = "could not resolve autostart directory: %v"
	ErrStartupReadDir        = "could not read autostart directory %s: %v"
	ErrStartupRemoveUsage    = "usage: gitmap startup-remove <name>"
	ErrStartupAddMissingExec = "startup-add: --exec is required " +
		"(or run from an installed gitmap binary so we can auto-detect it)"
	ErrStartupUnsupportedOS = "startup commands are not supported on this OS " +
		"(Linux/Unix XDG autostart and macOS LaunchAgents are supported; " +
		"on Windows use the Windows startup commands)"
	ErrStartupAddDarwinTODO = "startup-add is not yet implemented for macOS " +
		"(use list/remove for existing LaunchAgents; add support is tracked " +
		"in the OS-agnostic startup roadmap)"
)

// startup-add CLI flag descriptions. Kept here (not in
// constants_cli.go) so all startup-related strings live together and
// the flag parser file imports just one constants block.
const (
	FlagStartupAddName        = "name"
	FlagStartupAddExec        = "exec"
	FlagStartupAddDisplay     = "display-name"
	FlagStartupAddComment     = "comment"
	FlagStartupAddNoDisplay   = "no-display"
	FlagStartupAddForce       = "force"
	FlagDescStartupAddName    = "Logical name for the entry (filename becomes gitmap-<name>.desktop). Required."
	FlagDescStartupAddExec    = "Command to run at login (default: path to the running gitmap binary)"
	FlagDescStartupAddDisplay = "Override the Name= field shown in desktop session managers"
	FlagDescStartupAddComment = "Optional Comment= field text"
	FlagDescStartupAddNoDisplay = "Set NoDisplay=true so the entry stays out of app menus"
	FlagDescStartupAddForce   = "Overwrite an existing gitmap-managed entry (never overwrites third-party files)"
)

// startup-list CLI flag. Reuses the project-wide OutputTerminal/CSV/
// JSON constants for values; "table" is accepted as an alias for the
// default human-readable rendering ("terminal" works too).
const (
	FlagStartupListFormat     = "format"
	FlagDescStartupListFormat = "Output format: table (default), json, jsonl, or csv"
	StartupListFormatTable    = "table"
	// StartupListFormatJSONL is the jq/fluentd/BigQuery-friendly
	// line-oriented variant of --format=json: one compact object
	// per line, no array wrapper, empty list emits zero bytes.
	// Stable key order within each object is guaranteed via
	// stablejson.WriteJSONLines (same Field-slice contract as JSON).
	StartupListFormatJSONL    = "jsonl"
	ErrStartupListBadFormat   = "startup-list: unknown --format %q (expected: table, json, jsonl, csv)"
	// CSV header row for `--format csv`. Kept here so the column
	// order is centralized and tests can reference it.
	StartupListCSVHeader = "name,path,exec"
)

// startup-remove CLI flag. Single boolean for now; kept in its own
// const block so future flags (e.g. --trash, --backup-to) can land
// alongside without disturbing the list-format block above.
const (
	FlagStartupRemoveDryRun     = "dry-run"
	FlagDescStartupRemoveDryRun = "Show what would be deleted (or refused/no-op) without touching the filesystem"
)
