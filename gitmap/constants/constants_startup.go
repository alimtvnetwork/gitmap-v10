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
	CmdStartupList        = "startup-list"
	CmdStartupListAlias   = "sl"
	CmdStartupRemove      = "startup-remove"
	CmdStartupRemoveAlias = "sr"
)

// Startup help text. Format mirrors HelpInstall / HelpUninstall so
// the rendered `gitmap help` table stays visually consistent.
const (
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
	ErrStartupResolveDir     = "could not resolve autostart directory: %v"
	ErrStartupReadDir        = "could not read autostart directory %s: %v"
	ErrStartupRemoveUsage    = "usage: gitmap startup-remove <name>"
	ErrStartupUnsupportedOS  = "startup-list / startup-remove are Linux/Unix-only " +
		"(use the Windows startup commands on Windows)"
)
