package cmd

import (
	"flag"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// CloneNextFlags bundles every parsed flag from the clone-next command so
// the dispatcher in runCloneNext can branch on batch vs single mode without
// a 9-arg return list.
type CloneNextFlags struct {
	VersionArg   string
	Delete       bool
	Keep         bool
	NoDesktop    bool
	CreateRemote bool
	SSHKeyName   string
	Verbose      bool
	CSVPath      string
	All          bool
	// Force forces a flat clone even when the user's cwd IS the target
	// folder. Triggers a chdir-to-parent before the existence check (to
	// release Windows file locks) and DISABLES the versioned-folder
	// fallback so the user gets either a flat layout or a clear error.
	// See spec/01-app/87-clone-next-flatten.md.
	Force bool
}

// parseCloneNextFlags parses flags for the clone-next command.
func parseCloneNextFlags(args []string) CloneNextFlags {
	fs := flag.NewFlagSet(constants.CmdCloneNext, flag.ExitOnError)
	delFlag := fs.Bool("delete", false, constants.FlagDescCloneNextDelete)
	kpFlag := fs.Bool("keep", false, constants.FlagDescCloneNextKeep)
	noDesk := fs.Bool("no-desktop", false, constants.FlagDescCloneNextNoDesktop)
	createRem := fs.Bool("create-remote", false, constants.FlagDescCloneNextCreateRemote)
	sshKey := fs.String("ssh-key", "", "SSH key name for clone")
	fs.StringVar(sshKey, "K", "", "SSH key name (short)")
	verboseFlag := fs.Bool("verbose", false, constants.FlagDescVerbose)
	csvPath := fs.String("csv", "", constants.FlagDescCloneNextCSV)
	allFlag := fs.Bool("all", false, constants.FlagDescCloneNextAll)
	// Force-flatten: long --force and short -f both bind to the same
	// bool var so either form is canonical (mirrors the --ssh-key/-K
	// pairing convention used elsewhere in this flagset).
	forceFlag := fs.Bool("force", false, constants.FlagDescCloneNextForce)
	fs.BoolVar(forceFlag, "f", false, constants.FlagDescCloneNextForce)
	// Reorder so flags placed AFTER the positional version (e.g.
	// `gitmap cn v+1 -f`) are still recognized. Go's stdlib flag
	// parser stops at the first non-flag arg, so without this the
	// `-f` in the screenshot above was silently dropped.
	fs.Parse(reorderFlagsBeforeArgs(args))

	out := CloneNextFlags{
		Delete:       *delFlag,
		Keep:         *kpFlag,
		NoDesktop:    *noDesk,
		CreateRemote: *createRem,
		SSHKeyName:   *sshKey,
		Verbose:      *verboseFlag,
		CSVPath:      *csvPath,
		All:          *allFlag,
		Force:        *forceFlag,
	}
	if fs.NArg() > 0 {
		out.VersionArg = fs.Arg(0)
	}

	return out
}
