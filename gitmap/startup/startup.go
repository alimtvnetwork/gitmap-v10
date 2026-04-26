// Package startup manages Linux/Unix XDG autostart entries created by
// gitmap. Scope is intentionally narrow:
//
//   - List enumerates ONLY .desktop files in the user's autostart
//     directory that contain the X-Gitmap-Managed=true marker key.
//     Third-party autostart entries are silently ignored.
//   - Remove deletes a single named entry, but ONLY after re-confirming
//     it carries the marker. A request to remove a third-party file
//     becomes a refused no-op, never a deletion.
//
// macOS LaunchAgents (~/Library/LaunchAgents/*.plist) are NOT covered
// here — they use a different format and lifecycle (launchctl
// load/unload) that warrants its own implementation. The directory
// resolver returns an error on darwin so the CLI prints the
// "Linux/Unix-only" message instead of silently doing nothing.
package startup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// Entry is one gitmap-managed autostart record. Path is the absolute
// .desktop file path; Name is the basename WITHOUT the .desktop
// extension (the form users pass to `startup-remove`); Exec is the
// `Exec=` line value, surfaced so `startup-list` shows what would
// actually run at login.
type Entry struct {
	Name string
	Path string
	Exec string
}

// AutostartDir returns the absolute path to the user's XDG autostart
// directory. Honors $XDG_CONFIG_HOME when set, otherwise falls back
// to $HOME/.config/autostart per the freedesktop.org base-dir spec.
//
// Returns an error on Windows / macOS so callers can print the
// platform-specific "unsupported OS" message rather than touching a
// non-existent directory.
func AutostartDir() (string, error) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {

		return "", fmt.Errorf(constants.ErrStartupUnsupportedOS)
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); len(xdg) > 0 {

		return filepath.Join(xdg, "autostart"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {

		return "", err
	}

	return filepath.Join(home, ".config", "autostart"), nil
}

// List returns every gitmap-managed entry in the autostart dir. A
// MISSING directory is treated as "zero entries", NOT as an error —
// fresh user accounts that have never had any autostart file
// shouldn't see a scary error from `gitmap startup-list`.
func List() ([]Entry, error) {
	dir, err := AutostartDir()
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

	return collectManaged(dir, files), nil
}
