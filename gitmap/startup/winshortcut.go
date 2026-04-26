package startup

// .lnk Startup folder backend. Unlike the registry backend, this
// file compiles on every OS — it shells out to powershell.exe which
// only exists on Windows, but the Go code itself uses no Windows-
// only APIs. A runtime guard in addWindowsStartupFolder rejects
// non-Windows callers with the unsupported-OS error before any
// powershell invocation is attempted.
//
// PowerShell shellout rationale (vs hand-rolling [MS-SHLLINK]):
// the binary Shell Link format is ~80 pages of spec with multiple
// optional sub-records (LinkInfo, IDList, ExtraData) that tools
// like Explorer and `start` are picky about. Shipping a hand-rolled
// writer untested by the developer is a regression risk; shelling
// to `WScript.Shell.CreateShortcut` produces a .lnk that Windows
// itself authored, guaranteeing format correctness. Trade-off:
// powershell.exe must be on PATH (covered by ErrStartupPowerShellMissing).
//
// Marker contract: the .lnk filename uses the `gitmap-` prefix,
// AND a tracking subkey under HKCU\Software\Gitmap\StartupFolder\
// <name> records the entry as ours. Both must agree before Remove
// will delete the .lnk. Same belt-and-suspenders rule as the
// Registry backend.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// addWindowsStartupFolder writes a .lnk via PowerShell, then
// records ownership in the tracking subkey. Same five-status
// outcome model as every other backend.
func addWindowsStartupFolder(clean string, opts AddOptions) (AddResult, error) {
	if runtime.GOOS != "windows" {

		return AddResult{}, fmt.Errorf(constants.ErrStartupUnsupportedOS)
	}
	dir, err := startupFolderDir()
	if err != nil {

		return AddResult{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {

		return AddResult{}, fmt.Errorf("create startup folder %s: %w", dir, err)
	}
	full := filepath.Join(dir, constants.StartupWinValuePrefix+clean+constants.StartupLnkExt)

	return writeStartupShortcut(full, clean, opts)
}

// writeStartupShortcut performs the existence + ownership checks
// then either writes the .lnk (and tracking subkey) or refuses.
// Mirrors the structure of writeManaged on Linux so the five-status
// flow is visually parallel.
func writeStartupShortcut(full, clean string, opts AddOptions) (AddResult, error) {
	exists := fileExists(full)
	managed := exists && trackingSubkeyExists(constants.RegGitmapStartupFolder, clean)
	if exists && !managed {

		return AddResult{Status: AddRefused, Path: full}, nil
	}
	if exists && managed && !opts.Force {

		return AddResult{Status: AddExists, Path: full}, nil
	}
	if err := createShortcutViaPowerShell(full, opts.Exec); err != nil {

		return AddResult{}, fmt.Errorf(constants.ErrStartupShortcutCreate, full, err)
	}
	if err := writeTrackingSubkey(constants.RegGitmapStartupFolder, clean,
		opts.Exec, constants.StartupBackendStartupFolder); err != nil {

		return AddResult{}, err
	}
	if exists {

		return AddResult{Status: AddOverwritten, Path: full}, nil
	}

	return AddResult{Status: AddCreated, Path: full}, nil
}

// removeWindowsStartupFolder deletes a managed .lnk and its
// tracking subkey. Returns NoOp / Refused / Deleted with the same
// semantics as the Linux Remove path.
func removeWindowsStartupFolder(clean string, opts RemoveOptions) (RemoveResult, error) {
	if runtime.GOOS != "windows" {

		return RemoveResult{}, fmt.Errorf(constants.ErrStartupUnsupportedOS)
	}
	dir, err := startupFolderDir()
	if err != nil {

		return RemoveResult{}, err
	}
	full := filepath.Join(dir, constants.StartupWinValuePrefix+clean+constants.StartupLnkExt)
	if !fileExists(full) {

		return RemoveResult{Status: RemoveNoOp, DryRun: opts.DryRun}, nil
	}
	if !trackingSubkeyExists(constants.RegGitmapStartupFolder, clean) {

		return RemoveResult{Status: RemoveRefused, Path: full, DryRun: opts.DryRun}, nil
	}
	if opts.DryRun {

		return RemoveResult{Status: RemoveDeleted, Path: full, DryRun: true}, nil
	}
	if err := os.Remove(full); err != nil {

		return RemoveResult{}, fmt.Errorf("delete %s: %w", full, err)
	}
	if err := deleteTrackingSubkey(constants.RegGitmapStartupFolder, clean); err != nil {

		return RemoveResult{}, err
	}

	return RemoveResult{Status: RemoveDeleted, Path: full}, nil
}

// listWindowsStartupFolder enumerates managed .lnk files. Two-gate
// filter: filename prefix scan, then tracking-subkey re-check.
// Missing folder is "zero entries", not an error.
func listWindowsStartupFolder() ([]Entry, error) {
	if runtime.GOOS != "windows" {

		return nil, nil
	}
	dir, err := startupFolderDir()
	if err != nil {

		return nil, err
	}
	files, err := os.ReadDir(dir)
	if os.IsNotExist(err) {

		return nil, nil
	}
	if err != nil {

		return nil, fmt.Errorf(constants.ErrStartupReadDir, dir, err)
	}

	return collectStartupFolderManaged(dir, files), nil
}

// collectStartupFolderManaged is the per-file filter. Same shape
// as collectManagedDesktop / collectManagedPlist; differs only in
// the ownership re-check (subkey-based instead of file-marker).
func collectStartupFolderManaged(dir string, files []os.DirEntry) []Entry {
	var out []Entry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !looksLikeOursLnk(name) {
			continue
		}
		clean := strings.TrimSuffix(strings.TrimPrefix(name,
			constants.StartupWinValuePrefix), constants.StartupLnkExt)
		if !trackingSubkeyExists(constants.RegGitmapStartupFolder, clean) {
			continue
		}
		out = append(out, Entry{
			Name: strings.TrimSuffix(name, constants.StartupLnkExt),
			Path: filepath.Join(dir, name),
			Exec: "", // .lnk Exec lives inside the binary; not parsed here
		})
	}

	return out
}

// looksLikeOursLnk is the cheap pre-filter: filename ends in .lnk
// AND starts with the gitmap- prefix.
func looksLikeOursLnk(filename string) bool {
	if !strings.HasSuffix(filename, constants.StartupLnkExt) {
		return false
	}

	return strings.HasPrefix(filename, constants.StartupWinValuePrefix)
}

// startupFolderDir returns %APPDATA%\Microsoft\Windows\Start Menu\
// Programs\Startup. Honors $APPDATA so test fixtures can redirect
// to a temp dir.
func startupFolderDir() (string, error) {
	roaming := os.Getenv("APPDATA")
	if len(roaming) == 0 {

		return "", fmt.Errorf("APPDATA env var is empty")
	}

	return filepath.Join(roaming, constants.StartupFolderRelative), nil
}

// fileExists is a tiny os.Stat wrapper. Treats permission errors
// as "exists" (conservative — better to refuse a write than to
// silently overwrite a file we can't read).
func fileExists(p string) bool {
	if _, err := os.Stat(p); err == nil || !os.IsNotExist(err) {
		return true
	}

	return false
}

// createShortcutViaPowerShell shells out to WScript.Shell to author
// a syntactically perfect .lnk. The script is a single command
// passed via -Command; we deliberately avoid -EncodedCommand so the
// PowerShell history shows what gitmap actually ran (helps users
// audit what was created in their Startup folder).
func createShortcutViaPowerShell(lnkPath, target string) error {
	ps, err := exec.LookPath("powershell.exe")
	if err != nil {

		return fmt.Errorf(constants.ErrStartupPowerShellMissing)
	}
	script := buildShortcutScript(lnkPath, target)
	cmd := exec.Command(ps, "-NoProfile", "-NonInteractive", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {

		return fmt.Errorf("powershell: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}

	return nil
}

// buildShortcutScript composes the PowerShell snippet. Single-quoted
// path strings so backslashes survive without escape gymnastics;
// any embedded single-quote in the user-supplied target is doubled
// (PowerShell's standard single-quote escape).
func buildShortcutScript(lnkPath, target string) string {
	esc := func(s string) string { return strings.ReplaceAll(s, "'", "''") }

	return fmt.Sprintf("$ws = New-Object -ComObject WScript.Shell; "+
		"$s = $ws.CreateShortcut('%s'); "+
		"$s.TargetPath = '%s'; "+
		"$s.Save()", esc(lnkPath), esc(target))
}
