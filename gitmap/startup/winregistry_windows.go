//go:build windows

package startup

// Registry backend for `gitmap startup-add` / `startup-remove` /
// `startup-list` on Windows. Writes ONE direct value per entry to:
//
//   HKCU\Software\Microsoft\Windows\CurrentVersion\Run
//     gitmap-<name> = "<exec>"
//
// Ownership is tracked OUT-OF-BAND under a separate scope:
//
//   HKCU\Software\Gitmap\StartupRegistry\<name>
//     Exec      = "<exec>"
//     CreatedAt = "<RFC3339-UTC>"
//     Source    = "registry"
//
// Why no sibling marker in the Run key: Windows treats EVERY value
// under Run as an autostart command and feeds it to the shell at
// login. A `gitmap-<name>.gitmap-managed = "true"` sibling value
// shows up in Task Manager's Startup tab and is dispatched as a
// command (the literal string "true" — silently fails, but
// pollutes the user's startup surface). Keeping the ownership
// marker in a SEPARATE scope (HKCU\Software\Gitmap) means the Run
// key contains only real autostart commands, exactly like a
// hand-edited entry would. The trade-off: a user who manually
// deletes HKCU\Software\Gitmap loses the ability to refuse-overwrite
// a same-named third-party Run value — Add will treat any non-
// tracked Run value as third-party (refuse) regardless of who
// originally wrote it. This is the safer default.
//
// Build tag: only compiled on windows. The non-windows stub in
// winregistry_other.go provides the same symbols so cross-platform
// callers in winbackend.go compile everywhere.

import (
	"fmt"
	"time"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
	"golang.org/x/sys/windows/registry"
)

// addWindowsRegistry implements the Registry backend's Add path.
// The two writes (Run value + tracking subkey) are NOT atomic at
// the Windows API level — there is no transactional registry write
// across keys. We tolerate this: the tracking subkey is written
// FIRST, so a crash between the two writes leaves an "owned but
// inactive" record that future Add re-runs can safely overwrite
// (because classifyRunValue will see managed=true).
func addWindowsRegistry(clean string, opts AddOptions) (AddResult, error) {
	valueName := constants.StartupWinValuePrefix + clean
	exists, managed, err := classifyRunValue(valueName, clean)
	if err != nil {

		return AddResult{}, err
	}
	if exists && !managed {

		return AddResult{Status: AddRefused, Path: runValuePath(valueName)}, nil
	}
	if exists && managed && !opts.Force {

		return AddResult{Status: AddExists, Path: runValuePath(valueName)}, nil
	}
	if err := writeTrackingSubkey(constants.RegGitmapRegistrySub, clean,
		opts.Exec, constants.StartupBackendRegistry); err != nil {

		return AddResult{}, err
	}
	if err := writeRunValue(valueName, opts.Exec); err != nil {

		return AddResult{}, err
	}
	if exists {

		return AddResult{Status: AddOverwritten, Path: runValuePath(valueName)}, nil
	}

	return AddResult{Status: AddCreated, Path: runValuePath(valueName)}, nil
}

// classifyRunValue returns (exists, managed, error) for a Run-key
// value. "Managed" means the tracking subkey under HKCU\Software\
// Gitmap\StartupRegistry\<clean> exists. The Run key itself carries
// NO marker — keeping the autostart surface clean is the whole
// point of the direct-value model.
func classifyRunValue(valueName, clean string) (bool, bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, constants.RegRunKeyPath, registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {

			return false, false, nil
		}

		return false, false, fmt.Errorf(constants.ErrStartupRegistryOpen, constants.RegRunKeyPath, err)
	}
	defer k.Close()

	if _, _, err := k.GetStringValue(valueName); err != nil {
		if err == registry.ErrNotExist {

			return false, false, nil
		}

		return false, false, fmt.Errorf(constants.ErrStartupRegistryRead, valueName, err)
	}
	hasTracking := trackingSubkeyExists(constants.RegGitmapRegistrySub, clean)

	return true, hasTracking, nil
}

// trackingSubkeyExists checks for HKCU\<parent>\<name>. Returns
// false on any open error — same conservative posture as
// classifyTarget on Linux: an unreadable key is "not ours" so we
// refuse to delete it.
func trackingSubkeyExists(parent, name string) bool {
	full := parent + `\` + name
	k, err := registry.OpenKey(registry.CURRENT_USER, full, registry.QUERY_VALUE)
	if err != nil {

		return false
	}
	k.Close()

	return true
}

// writeRunValue writes the autostart command to the Run key. ONE
// value per entry — no sibling marker. Ownership tracking lives
// entirely under HKCU\Software\Gitmap\StartupRegistry.
func writeRunValue(valueName, exec string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER,
		constants.RegRunKeyPath, registry.SET_VALUE)
	if err != nil {

		return fmt.Errorf(constants.ErrStartupRegistryOpen, constants.RegRunKeyPath, err)
	}
	defer k.Close()

	if err := k.SetStringValue(valueName, exec); err != nil {

		return fmt.Errorf(constants.ErrStartupRegistryWrite, valueName, err)
	}

	return nil
}

// writeTrackingSubkey creates HKCU\<parent>\<name> with the three
// metadata values (Exec / CreatedAt / Source). CreatedAt is RFC3339
// UTC for stable cross-locale parsing by future tooling. Used by
// BOTH the registry backend (Source="registry") and the .lnk
// backend (Source="startup-folder").
func writeTrackingSubkey(parent, name, exec, source string) error {
	full := parent + `\` + name
	k, _, err := registry.CreateKey(registry.CURRENT_USER, full, registry.SET_VALUE)
	if err != nil {

		return fmt.Errorf(constants.ErrStartupRegistryOpen, full, err)
	}
	defer k.Close()

	if err := k.SetStringValue(constants.RegTrackKeyExec, exec); err != nil {

		return fmt.Errorf(constants.ErrStartupRegistryWrite, constants.RegTrackKeyExec, err)
	}
	if err := k.SetStringValue(constants.RegTrackKeyCreatedAt,
		time.Now().UTC().Format(time.RFC3339)); err != nil {

		return fmt.Errorf(constants.ErrStartupRegistryWrite, constants.RegTrackKeyCreatedAt, err)
	}
	if err := k.SetStringValue(constants.RegTrackKeySource, source); err != nil {

		return fmt.Errorf(constants.ErrStartupRegistryWrite, constants.RegTrackKeySource, err)
	}

	return nil
}

// runValuePath formats a stable user-facing locator for a Run-key
// value: `HKCU\<RunPath>\<valueName>`. Used for AddResult.Path so
// `gitmap startup-list` can show the user the full registry path
// they would see in regedit.
func runValuePath(valueName string) string {
	return `HKCU\` + constants.RegRunKeyPath + `\` + valueName
}
