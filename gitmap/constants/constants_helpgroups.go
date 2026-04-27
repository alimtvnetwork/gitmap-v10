package constants

// Help group headers.
const (
	HelpGroupScanning    = "  Scanning & Discovery:"
	HelpGroupCloning     = "  Cloning & Sync:"
	HelpGroupGitOps      = "  Git Operations:"
	HelpGroupNavigation  = "  Navigation & Organization:"
	HelpGroupRelease     = "  Release & Versioning:"
	HelpGroupReleaseInfo = "  Release History & Info:"
	HelpGroupData        = "  Data, Profiles & Bookmarks:"
	HelpGroupHistory     = "  History & Stats:"
	HelpGroupAmendGroup  = "  Author Amendment:"
	HelpGroupProject     = "  Project Detection:"
	HelpGroupSSH         = "  SSH Key Management:"
	HelpGroupZip         = "  Zip Groups (Release Archives):"
	HelpGroupEnvTools    = "  Environment & Tools:"
	HelpGroupTasks       = "  File-Sync Tasks:"
	HelpGroupUtilities   = "  Utilities:"
	HelpGroupVisualize   = "  Visualization:"
	HelpGroupCommitXfer  = "  Commit Transfer (replay between repos):"

	HelpGroupHint    = "  Run any command with --help or -h for detailed usage and examples."
	HelpGroupExample = "  Quick start:"
	HelpExampleScan  = "    $ gitmap scan ~/projects"
	HelpExampleList  = "    $ gitmap ls"
	HelpExamplePull  = "    $ gitmap pull my-api"
	HelpExampleCD    = "    $ gitmap cd my-api"
	HelpCompactHint  = "  Use --compact for a minimal command list without descriptions."

	HelpAlias    = "  alias (a) <sub>     Assign short names to repos (set, remove, list, show, suggest)"
	HelpSSH      = "  ssh <sub>           Generate, list, and manage SSH keys for Git authentication"
	HelpZipGroup = "  zip-group (z) <sub> Manage named file collections for release ZIP archives"

	// Compact-mode lines: command (alias) only.
	CompactScanning   = "  scan (s), rescan (rsc), rescan-subtree (rss), list (ls)"
	CompactCloning    = "  clone (c), clone-next (cn), desktop-sync (ds), github-desktop (gd)"
	CompactGitOps     = "  pull (p), exec (x), status (st), watch (w), has-any-updates, latest-branch (lb)"
	CompactNavigation = "  cd (go), group (g), multi-group (mg), alias (a), diff-profiles (dp)"
	CompactRelease    = "  release (r), release-self (rs), release-branch (rb), temp-release"
	CompactRelInfo    = "  changelog (cl), changelog-generate, list-versions (lv), list-releases (lr), release-pending (rp), revert, clear-release-json (crj), prune"
	CompactData       = "  export (ex), import (im), profile (pf), bookmark (bk), db-reset"
	CompactHistory    = "  history (hi), history-reset (hr), stats (ss)"
	CompactAmend      = "  amend (am), amend-list (al)"
	CompactProject    = "  go-repos (gr), node-repos (nr), react-repos (rr), cpp-repos (cr), csharp-repos (csr)"
	CompactSSH        = "  ssh"
	CompactZip        = "  zip-group (z)"
	CompactEnvTools   = "  env, install (in), uninstall (un)"
	CompactTasks      = "  task"
	CompactVisualize  = "  dashboard (db)"
	CompactCommitXfer = "  commit-right (cmr) — LIVE,  commit-left (cml), commit-both (cmb) — scaffolds"
	CompactUtilities  = "  setup, doctor, update, update-cleanup, version (v), completion (cmp), interactive (i), docs (d), help-dashboard (hd), gomod (gm), seo-write (sw), help"

	CompactNoMatchFmt = "  No group matching '%s'. Showing all groups:\n"
)

// HelpGroupKeys returns short keywords for tab-completion of group filtering.
var HelpGroupKeys = []string{
	"scanning",
	"cloning",
	"gitops",
	"navigation",
	"release",
	"release-info",
	"data",
	"history",
	"amend",
	"project",
	"ssh",
	"zip",
	"environment",
	"tasks",
	"visualization",
	"commit-transfer",
	"utilities",
}
