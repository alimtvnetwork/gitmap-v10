package startup

// Windows backend selection + dispatch. Lives in its own file (not
// inline in startup.go) so the cross-OS startup.go stays focused on
// Linux/.desktop logic and the Windows surface area is easy to find
// and code-review independently.
//
// Two backends are supported and a single Backend value flows
// through Add → writeManagedWindows. List enumerates BOTH backends
// unconditionally so users do not have to remember which one they
// chose at Add time. Remove also looks in BOTH backends; the
// tracking subkey under HKCU\Software\Gitmap pins which backend the
// entry actually lives in so we never delete a Run value that
// belongs to the .lnk path's tracking record (or vice versa).

import (
	"fmt"
	"runtime"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// Backend enumerates the Windows-specific add targets. Linux/macOS
// have one canonical backend each (XDG .desktop / LaunchAgents
// .plist) so they ignore this value entirely. The zero value
// (BackendUnspecified) means "let the dispatcher pick the OS default"
// which is `BackendRegistry` on Windows.
type Backend int

const (
	// BackendUnspecified is the zero value. The dispatcher
	// translates it to the per-OS default rather than failing —
	// keeps the public Add(opts) API ergonomic for non-Windows
	// callers that have no opinion on backend.
	BackendUnspecified Backend = iota
	// BackendRegistry writes the Run-key value + tracking subkey.
	BackendRegistry
	// BackendStartupFolder writes the .lnk + tracking subkey.
	BackendStartupFolder
)

// ParseBackend translates the user-facing flag string into the
// Backend enum. Unknown / empty values from non-Windows callers
// flow through as BackendUnspecified (the per-OS default); empty
// values from Windows callers also default to Registry. ONLY a
// non-empty unrecognized value is an error — that means the user
// typed something we don't understand and silently defaulting
// would hide the typo.
func ParseBackend(s string) (Backend, error) {
	switch s {
	case "":
		return BackendUnspecified, nil
	case constants.StartupBackendRegistry:
		return BackendRegistry, nil
	case constants.StartupBackendStartupFolder:
		return BackendStartupFolder, nil
	default:
		return BackendUnspecified, fmt.Errorf(constants.ErrStartupAddBadBackend, s)
	}
}

// String renders the backend as the canonical CLI flag value. Used
// by list/remove rendering to label which backend an entry lives in.
func (b Backend) String() string {
	switch b {
	case BackendRegistry:
		return constants.StartupBackendRegistry
	case BackendStartupFolder:
		return constants.StartupBackendStartupFolder
	default:
		return ""
	}
}

// resolveBackendForAdd picks the concrete backend to use when the
// caller passed BackendUnspecified. Registry is the Windows default
// (matches what most provisioning scripts use and what `regedit`
// users expect to see after running `gitmap startup-add`).
func resolveBackendForAdd(b Backend) Backend {
	if b != BackendUnspecified {
		return b
	}
	if runtime.GOOS == "windows" {
		return BackendRegistry
	}

	return BackendUnspecified
}

// addWindows is the Windows branch of Add. Routes through the
// chosen backend (Registry or Startup-folder). Lives here (not in
// add.go) so all Windows dispatch code lives together. Both
// backends return AddResult with Path set to a backend-appropriate
// locator: registry value path for the Run-key entry, filesystem
// path for the .lnk Startup-folder shortcut.
func addWindows(clean string, opts AddOptions) (AddResult, error) {
	backend := resolveBackendForAdd(opts.Backend)
	switch backend {
	case BackendRegistry:

		return addWindowsRegistry(clean, opts)
	case BackendStartupFolder:

		return addWindowsStartupFolder(clean, opts)
	default:

		return AddResult{}, fmt.Errorf(constants.ErrStartupAddBadBackend, backend.String())
	}
}
