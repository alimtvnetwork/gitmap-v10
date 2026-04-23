package constants

// gitmap:cmd top-level
// SSH command names.
const (
	CmdSSH           = "ssh"
	SubCmdSSHCat     = "cat"
	SubCmdSSHList    = "list"
	SubCmdSSHListS   = "ls"
	SubCmdSSHDelete  = "delete"
	SubCmdSSHDeleteS = "rm"
	SubCmdSSHConfig  = "config"
)

// SshKey table (v15: singular + SshKeyId PK; abbreviation per v15: Ssh, not SSH).
const TableSshKey = "SshKey"

// Legacy plural retained for migration detection.
const LegacyTableSSHKeys = "SSHKeys"

// SQL: create SshKey table (v15).
const SQLCreateSshKey = `CREATE TABLE IF NOT EXISTS SshKey (
	SshKeyId    INTEGER PRIMARY KEY AUTOINCREMENT,
	Name        TEXT NOT NULL UNIQUE,
	PrivatePath TEXT NOT NULL,
	PublicKey   TEXT NOT NULL,
	Fingerprint TEXT NOT NULL,
	Email       TEXT DEFAULT '',
	CreatedAt   TEXT DEFAULT CURRENT_TIMESTAMP
)`

// SQL: SshKey operations (v15). Constant names retain SSH for callsite stability.
const (
	SQLInsertSSHKey = `INSERT INTO SshKey (Name, PrivatePath, PublicKey, Fingerprint, Email)
		VALUES (?, ?, ?, ?, ?)`

	SQLUpdateSSHKey = `UPDATE SshKey SET PrivatePath = ?, PublicKey = ?, Fingerprint = ?, Email = ?
		WHERE Name = ?`

	SQLSelectAllSSHKeys = `SELECT SshKeyId, Name, PrivatePath, PublicKey, Fingerprint, Email, CreatedAt
		FROM SshKey ORDER BY Name`

	SQLSelectSSHKeyByName = `SELECT SshKeyId, Name, PrivatePath, PublicKey, Fingerprint, Email, CreatedAt
		FROM SshKey WHERE Name = ?`

	SQLDeleteSSHKeyByName = `DELETE FROM SshKey WHERE Name = ?`
)

// SQL: drop SshKey table (and legacy plural).
const (
	SQLDropSshKey  = "DROP TABLE IF EXISTS SshKey"
	SQLDropSSHKeys = "DROP TABLE IF EXISTS SSHKeys" // legacy
)

// SSH key generation defaults.
const (
	SSHKeyType        = "rsa"
	SSHKeyBits        = "4096"
	DefaultSSHKeyName = "default"
	SSHKeygenBin      = "ssh-keygen"
)

// SSH key generation flags.
const (
	FlagSSHName    = "--name"
	FlagSSHNameS   = "-n"
	FlagSSHPath    = "--path"
	FlagSSHPathS   = "-p"
	FlagSSHEmail   = "--email"
	FlagSSHEmailS  = "-e"
	FlagSSHForce   = "--force"
	FlagSSHForceS  = "-f"
	FlagSSHFiles   = "--files"
	FlagSSHKey     = "--ssh-key"
	FlagSSHKeyS    = "-K"
	FlagSSHHost    = "--host"
	FlagSSHHostS   = "-H"
	FlagSSHJSON    = "--json"
	FlagSSHConfirm = "--confirm"
)

// SSH defaults.
const (
	DefaultSSHHost = "github.com"
)

// SSH config markers.
const (
	SSHConfigMarkerStart = "# --- gitmap managed (do not edit) ---"
	SSHConfigMarkerEnd   = "# --- end gitmap managed ---"
)

// SSH config host template.
const SSHConfigHostEntry = `Host %s
    HostName %s
    User git
    IdentityFile %s
    IdentitiesOnly yes
`

// SSH messages.
const (
	MsgSSHGenerated     = "  \u2713 SSH key %q generated\n"
	MsgSSHPath          = "    Path:        %s\n"
	MsgSSHFingerprint   = "    Fingerprint: %s\n"
	MsgSSHPubLabel      = "    Public key:\n\n"
	MsgSSHCopyHint      = "\n  \u2139  Copy the public key above and add it to your Git provider.\n"
	MsgSSHExists        = "  Key %q already exists at %s\n"
	MsgSSHExistsFP      = "    Fingerprint: %s\n"
	MsgSSHPromptAction  = "  [R]egenerate / [N]ew path / [C]ancel: "
	MsgSSHRegenerated   = "  \u2713 SSH key %q regenerated\n"
	MsgSSHDeleted       = "  \u2713 SSH key %q deleted\n"
	MsgSSHDeletedFiles  = "  \u2713 Key files removed from disk\n"
	MsgSSHDeleteConfirm = "  Delete SSH key %q? (y/N): "
	MsgSSHListHeader    = "\n  SSH Keys (%d):\n\n"
	MsgSSHListRow       = "  %-15s %-30s %-25s %s\n"
	MsgSSHListColumns   = "  %-15s %-30s %-25s %s\n"
	MsgSSHConfigDone    = "  \u2713 SSH config updated\n"
	MsgSSHConfigShow    = "\n  Managed SSH config:\n\n"
	MsgSSHNewPathPrompt = "  Enter new key path: "
	MsgSSHCloneUsing    = "  \u2192 Cloning with SSH key %q (%s)\n"
	MsgSSHMultiKeyHint  = `
  Multiple SSH keys detected. Use host aliases in your remote URLs:
    git remote set-url origin git@github.com-%s:%s/%s.git
`
	MsgSSHConfirmPrompt = "  Generate SSH key %q at %s? (y/N): "
	MsgSSHCanceled      = "  Canceled.\n"
	MsgSSHHostUsed      = "    Host:        %s\n"
)

// SSH error messages — Code Red: all file errors include exact path and reason.
const (
	ErrSSHKeygen        = "Error: SSH key generation failed at %s: %v (operation: write)\n"
	ErrSSHReadPub       = "Error: failed to read public key at %s: %v (operation: read, reason: file does not exist)\n"
	ErrSSHNotFound      = "Error: SSH key not found: %s\n"
	ErrSSHAvailable     = "  Available keys: %s\n"
	ErrSSHNameEmpty     = "SSH key name cannot be empty"
	ErrSSHCreate        = "failed to create SSH key record: %v"
	ErrSSHQuery         = "failed to query SSH keys: %v"
	ErrSSHDelete        = "failed to delete SSH key: %v"
	ErrSSHConfig        = "Error: failed to update SSH config at %s: %v (operation: write)\n"
	ErrSSHKeygenMissing = "Error: ssh-keygen not found on PATH (operation: resolve, reason: file does not exist)\n"
	ErrSSHEmailResolve  = "could not resolve email; use --email flag\n"
	ErrSSHFingerprint   = "Error: failed to read key fingerprint at %s: %v (operation: read)\n"
)

// SSH completion flag.
const CompListSSHKeys = "--list-ssh-keys"
