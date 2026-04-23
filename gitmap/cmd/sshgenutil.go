package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// validateSSHKeygen checks if ssh-keygen is available on PATH.
func validateSSHKeygen() error {
	_, err := exec.LookPath(constants.SSHKeygenBin)

	return err
}

// resolveGitEmail reads the global Git email config.
func resolveGitEmail() string {
	out, err := exec.Command("git", "config", "--global", "user.email").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

// readFingerprint reads the SHA256 fingerprint of a key file.
func readFingerprint(keyPath string) string {
	out, err := exec.Command(constants.SSHKeygenBin, "-lf", keyPath+".pub").Output()
	if err != nil {
		return "unknown"
	}

	parts := strings.Fields(string(out))
	if len(parts) >= 2 {
		return parts[1]
	}

	return "unknown"
}

// removeKeyFiles deletes private and public key files.
func removeKeyFiles(privatePath string) {
	if err := os.Remove(privatePath); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "  ⚠ Could not remove %s: %v\n", privatePath, err)
	}
	if err := os.Remove(privatePath + ".pub"); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "  ⚠ Could not remove %s: %v\n", privatePath+".pub", err)
	}
}

// defaultSSHKeyPath returns the default key path based on name.
func defaultSSHKeyPath(name string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ⚠ Could not determine home directory: %v\n", err)

		return filepath.Join(".", ".ssh", "id_rsa")
	}
	if name == constants.DefaultSSHKeyName {
		return filepath.Join(home, ".ssh", "id_rsa")
	}

	return filepath.Join(home, ".ssh", "id_rsa_"+name)
}

// expandHome expands ~ to the user's home directory.
func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ Could not determine home directory: %v\n", homeErr)

			return path
		}
		path = filepath.Join(home, path[1:])
	}

	return path
}

// ensureSSHDir creates a directory with 0700 permissions if it doesn't exist.
func ensureSSHDir(dir string) error {
	return os.MkdirAll(dir, 0o700)
}
