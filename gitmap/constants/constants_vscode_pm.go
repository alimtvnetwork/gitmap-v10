package constants

// VS Code Project Manager (alefragnani.project-manager) sync constants.
//
// Path resolution: discover the VS Code USER-DATA root per OS first, then
// append the relative tail. Never hardcode the full path.
//
//	Windows : %APPDATA%\Code           (fallback %USERPROFILE%\AppData\Roaming\Code)
//	macOS   : $HOME/Library/Application Support/Code
//	Linux   : $XDG_CONFIG_HOME/Code   (fallback $HOME/.config/Code)
//
// Final path = <userDataRoot>/User/globalStorage/alefragnani.project-manager/projects.json
//
// See: spec/01-vscode-project-manager-sync/README.md

// User-data root segments per OS.
const (
	VSCodeUserDataRootDirName     = "Code"
	VSCodeUserDataMacRel          = "Library/Application Support/Code"
	VSCodeUserDataLinuxFallback   = ".config/Code"
	VSCodeEnvAppData              = "APPDATA"
	VSCodeEnvUserProfile          = "USERPROFILE"
	VSCodeEnvHome                 = "HOME"
	VSCodeEnvXDGConfigHome        = "XDG_CONFIG_HOME"
	VSCodeUserProfileAppDataRel   = "AppData/Roaming/Code"
)

// Relative tail under the user-data root (constant across all OSes).
const (
	VSCodePMUserDir              = "User"
	VSCodePMGlobalStorageDir     = "globalStorage"
	VSCodePMExtensionDir         = "alefragnani.project-manager"
	VSCodePMProjectsFile         = "projects.json"
	VSCodePMProjectsTempSuffix   = ".tmp"
	VSCodePMJSONIndent           = "\t"
)

// Default field values gitmap writes when inserting a NEW projects.json entry.
// Existing entries' values are preserved across re-syncs.
const (
	VSCodePMDefaultEnabled = true
	VSCodePMDefaultProfile = ""
)

// CLI flag for opting out of the automatic sync during scan.
const (
	FlagNoVSCodeSync     = "no-vscode-sync"
	FlagDescNoVSCodeSync = "skip syncing scanned repos into VS Code Project Manager projects.json"
)

// Error messages (Code Red zero-swallow policy).
const (
	ErrVSCodePMUserDataNotFound  = "vscode: user data directory not found at %q (is VS Code installed?)\n"
	ErrVSCodePMExtDirMissing     = "vscode: project-manager extension dir not found at %q (open VS Code, install the alefragnani.project-manager extension, then retry)\n"
	ErrVSCodePMReadFailed        = "vscode: failed to read %s: %v\n"
	ErrVSCodePMParseFailed       = "vscode: %s is not valid JSON: %v (left untouched)\n"
	ErrVSCodePMWriteTempFailed   = "vscode: failed to write temp %s: %v\n"
	ErrVSCodePMRenameFailed      = "vscode: failed to commit %s: %v\n"
	ErrVSCodePMNoUserDataEnv     = "vscode: cannot determine user-data directory (no APPDATA / USERPROFILE / HOME env)\n"
)

// User-facing messages.
const (
	MsgVSCodePMSectionHeader = "  → VS Code Project Manager: %s\n"
	MsgVSCodePMSyncSummary   = "  ✓ projects.json synced: %d added, %d updated, %d unchanged (%d total)\n"
	MsgVSCodePMSyncSkipped   = "  • VS Code Project Manager sync skipped (--no-vscode-sync)\n"
	MsgVSCodePMRenamed       = "  ✓ projects.json: renamed %q -> %q\n"
	MsgVSCodePMRenameNoMatch = "  • projects.json: no entry matched %q (skipped rename)\n"

	// Diagnostic messages used by `gitmap vscode-pm-path` (v3.41.0+).
	MsgVSCodePMPathRootMissing = "vscode: user-data directory not found (is VS Code installed? checked APPDATA / HOME / XDG_CONFIG_HOME)"
	MsgVSCodePMPathExtMissing  = "vscode: project-manager extension storage dir not found near %s (open VS Code, install the alefragnani.project-manager extension, then retry)\n"
)
