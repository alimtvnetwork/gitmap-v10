//go:build windows

package startup

// Registry backend for `gitmap startup-add` / `startup-remove` /
// `startup-list` on Windows. Writes to:
//
//   HKCU\Software\Microsoft\Windows\CurrentVersion\Run
//     gitmap-<name>                = "<exec>"
//     gitmap-<name>.gitmap-managed = "true"   (sibling marker)
//
//   HKCU\Software\Gitmap\StartupRegistry\<name>
//     Exec      = "<exec>"
//     CreatedAt = "<RFC3339-UTC>"
//     Source    = "registry"
//
// Both records (Run-key sibling marker AND tracking subkey) must
// agree the entry is gitmap-managed before Remove will delete the
// Run value. Belt-and-suspenders: a user manually removing the
// HKCU\Software\Gitmap subtree leaves the Run-key value behind, but
// Add can still re-claim it because the sibling marker proves
// previous gitmap ownership.
//
// Build tag: only compiled on windows. The non-windows stub in
// winregistry_other.go provides the same symbols so cross-platform
// callers in winbackend.go compile everywhere.

import (
	"fmt"
	"strings"
	"time"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
	"golang.org/x/sys/windows/registry"
)

// addWindowsRegistry implements the Registry backend's Add path.
// The two writes (Run value + tracking subkey) are NOT atomic at
// the Windows API level — there is no transactional registry write
// across keys. We tolerate this because the sibling marker value
// is written FIRST under the same key as the Run value, so a crash
// between the two writes leaves an entry that future Add/Remove
// still recognize as ours.
func addWindowsRegistry(clean string, opts AddOptions) (AddResult, error) {
	valueName := constants.StartupWinValuePrefix + clean
	exists, managed, err := classifyRunValue(valueName)
	if err != nil {

		return AddResult{}, err
	}
	if exists && !managed {

		return AddResult{Status: AddRefused, Path: runValuePath(valueName)}, nil
	}
	if exists && managed && !opts.Force {

		return AddResult{Status: AddExists, Path: runValuePath(valueName)}, nil
	}
	if err := writeRunValueAndMarker(valueName, opts.Exec); err != nil {

		return AddResult{}, err
	}
	if err := writeTrackingSubkey(constants.RegGitmapRegistrySub, clean,
		opts.Exec, constants.StartupBackendRegistry); err != nil {

		return AddResult{}, err
	}
	if exists {

		return AddResult{Status: AddOverwritten, Path: runValuePath(valueName)}, nil
	}

	return AddResult{Status: AddCreated, Path: runValuePath(valueName)}, nil
}

// classifyRunValue returns (exists, managed, error) for a Run-key
// value. "Managed" means BOTH the sibling .gitmap-managed value
// equals "true" AND the tracking subkey under HKCU\Software\Gitmap
// exists. Either alone is treated as "not ours" so a stray sibling
// marker on a third-party value cannot trick Remove into deleting it.
func classifyRunValue(valueName string) (bool, bool, error) {
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
	marker, _, err := k.GetStringValue(valueName + constants.RegMarkerSiblingSuffix)
	if err != nil || marker != "true" {

		return true, false, nil
	}
	tracking := strings.TrimPrefix(valueName, constants.StartupWinValuePrefix)
	hasTracking := trackingSubkeyExists(constants.RegGitmapRegistrySub, tracking)

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

// writeRunValueAndMarker writes the command and the sibling marker
// in a single OpenKey scope. Marker is written FIRST so a crash
// after the marker but before the command leaves a "claim" record
// future Add can recognize and overwrite (rather than refusing).
func writeRunValueAndMarker(valueName, exec string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER,
		constants.RegRunKeyPath, registry.SET_VALUE)
	if err != nil {

		return fmt.Errorf(constants.ErrStartupRegistryOpen, constants.RegRunKeyPath, err)
	}
	defer k.Close()

	markerName := valueName + constants.RegMarkerSiblingSuffix
	if err := k.SetStringValue(markerName, "true"); err != nil {

		return fmt.Errorf(constants.ErrStartupRegistryWrite, markerName, err)
	}
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
