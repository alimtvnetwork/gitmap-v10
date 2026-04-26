package startup

// Remove implements `gitmap startup-remove <name>` with a strict
// "managed-only, never escalate" contract:
//
//  1. Reject inputs that contain a path separator OR are empty —
//     prevents `..`-based traversal and accidental deletion of files
//     outside the autostart dir.
//  2. Resolve the full target path inside AutostartDir(). If the
//     file does not exist → return RemoveNoOp (clean exit 0, "did
//     nothing"). Missing files are NOT errors — list/remove must be
//     idempotent so users can safely script them.
//  3. Open and re-check the X-Gitmap-Managed marker. If the file
//     exists but is not ours → return RemoveRefused (also clean exit
//     0 with a clear message; never delete a third-party file).
//  4. Only on a verified gitmap-managed file: os.Remove and return
//     RemoveDeleted with the absolute path so the CLI can echo it.

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// RemoveResult tags every Remove outcome so callers can render the
// right user-facing message without inspecting error text. The path
// is empty for NotOurs / NoOp / BadName so callers don't accidentally
// print a path the user can't act on. DryRun reports whether the
// outcome was simulated (no filesystem mutation) — set when the
// caller passed RemoveOptions.DryRun=true and the status would
// otherwise have been RemoveDeleted.
type RemoveResult struct {
	Status RemoveStatus
	Path   string
	DryRun bool
}

// RemoveStatus enumerates the four mutually-exclusive outcomes.
type RemoveStatus int

const (
	// RemoveDeleted = file existed, was gitmap-managed, was unlinked
	// (or, under DryRun, would have been unlinked).
	RemoveDeleted RemoveStatus = iota
	// RemoveNoOp = no file by that name in the autostart dir.
	RemoveNoOp
	// RemoveRefused = file exists but is not gitmap-managed.
	RemoveRefused
	// RemoveBadName = input failed validation (empty / separator).
	RemoveBadName
)

// RemoveOptions carries optional knobs for Remove. Kept as a struct
// (not extra positional args) so future flags like Trash bool or
// BackupTo string can be added without breaking callers.
type RemoveOptions struct {
	// DryRun runs the full classification (existence + marker check)
	// but skips the actual os.Remove call. The returned RemoveResult
	// has DryRun=true and the Status the live call would have
	// produced — letting CLI renderers print "would delete X" with
	// the same accuracy as a real run.
	DryRun bool
}

// Remove deletes the named gitmap-managed autostart entry. `name`
// is the basename WITHOUT the platform extension (the same form
// `List` returns); a trailing platform extension is tolerated so
// users who copy/paste from `ls` get the same behavior. The
// extension is `.desktop` on Linux/Unix and `.plist` on macOS.
//
// This is the legacy entry point — equivalent to
// RemoveWithOptions(name, RemoveOptions{}). New callers that need
// dry-run semantics should call RemoveWithOptions directly.
func Remove(name string) (RemoveResult, error) {
	return RemoveWithOptions(name, RemoveOptions{})
}

// RemoveWithOptions is the full-featured entry point. The dry-run
// branch reuses every classification check the live branch runs so
// `--dry-run` cannot disagree with the real command — only the
// final os.Remove is suppressed.
func RemoveWithOptions(name string, opts RemoveOptions) (RemoveResult, error) {
	clean := normalizeName(name)
	if !isValidName(clean) {

		return RemoveResult{Status: RemoveBadName, DryRun: opts.DryRun}, nil
	}
	dir, err := AutostartDir()
	if err != nil {

		return RemoveResult{}, err
	}
	full := joinPath(dir, clean+platformExt())

	return removeIfManaged(full, opts)
}

// normalizeName strips an optional platform extension and surrounding
// whitespace so `Remove("foo")`, `Remove("foo.desktop")`/`("foo.plist")`,
// and `Remove(" foo ")` all resolve to the same target file.
func normalizeName(name string) string {
	clean := strings.TrimSpace(name)

	return strings.TrimSuffix(clean, platformExt())
}

// isValidName rejects empty strings and any input containing a path
// separator. Both are security-critical — without this check a user
// (or a script bug) could pass `../../.bashrc` and we would happily
// stat / open it. We also reject names containing NUL since some
// Linux filesystems treat embedded NULs as path terminators.
func isValidName(name string) bool {
	if len(name) == 0 {

		return false
	}
	if strings.ContainsAny(name, "/\\\x00") {

		return false
	}

	return true
}

// removeIfManaged runs the existence + managed-marker checks and
// performs the unlink. Splits cleanly into four mutually-exclusive
// branches so the caller's switch on RemoveStatus is exhaustive.
func removeIfManaged(full string) (RemoveResult, error) {
	if _, err := os.Stat(full); os.IsNotExist(err) {

		return RemoveResult{Status: RemoveNoOp}, nil
	}
	if !isManagedFile(full) {

		return RemoveResult{Status: RemoveRefused, Path: full}, nil
	}
	if err := os.Remove(full); err != nil {

		return RemoveResult{}, fmt.Errorf("delete %s: %w", full, err)
	}

	return RemoveResult{Status: RemoveDeleted, Path: full}, nil
}

// isManagedFile is the second-look marker check. We re-open the file
// at remove time (rather than trusting a stale List() snapshot) so a
// race between `startup-list` printing and the user typing
// `startup-remove` cannot trick us into deleting a file that has
// since been replaced by a third-party autostart entry with the same
// name. Dispatches to the per-OS parser since the marker grammar
// differs (`X-Gitmap-Managed=true` line vs `<key>XGitmapManaged</key>
// <true/>` plist element pair).
func isManagedFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {

		return false
	}
	defer f.Close()

	if runtime.GOOS == "darwin" {
		managed, _ := parsePlistFields(f)

		return managed
	}
	managed, _ := parseDesktopFields(newScanner(f))

	return managed
}

// platformExt returns the autostart filename extension for the
// running OS. Centralized so Remove and normalizeName agree without
// either importing the runtime constant directly.
func platformExt() string {
	if runtime.GOOS == "darwin" {

		return constants.StartupPlistExt
	}

	return constants.StartupDesktopExt
}
