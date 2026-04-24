package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// repoRoot returns the absolute path of the current git repo. Exits 1
// when not inside a repo, per the zero-swallow error policy.
func repoRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrReplaceNotInRepo, err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}

// walkRepoFiles returns every text file under root, honoring the
// directory and prefix exclusions from constants.
func walkRepoFiles(root string) ([]string, error) {
	out := make([]string, 0, 1024)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		return visitReplaceEntry(root, path, d, &out)
	})
	return out, err
}

// visitReplaceEntry implements one step of walkRepoFiles. Split out so
// the closure stays under the 15-line ceiling.
func visitReplaceEntry(root, path string, d fs.DirEntry, out *[]string) error {
	if d.IsDir() {
		if isExcludedDir(d.Name()) || isExcludedPrefix(root, path) {
			return filepath.SkipDir
		}
		return nil
	}
	if isBinaryFile(path) {
		return nil
	}
	*out = append(*out, path)
	return nil
}

// isExcludedDir checks the directory's base name against the deny list.
func isExcludedDir(name string) bool {
	for _, ex := range constants.ReplaceExcludedDirs {
		if name == ex {
			return true
		}
	}
	return false
}

// isExcludedPrefix matches a path against the prefix deny list, using
// forward slashes for portability.
func isExcludedPrefix(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	for _, p := range constants.ReplaceExcludedPrefixes {
		if rel == p || strings.HasPrefix(rel, p+"/") {
			return true
		}
	}
	return false
}

// isBinaryFile sniffs the first ReplaceBinarySniffBytes bytes for a
// null byte — git's classic heuristic.
func isBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()

	buf := make([]byte, constants.ReplaceBinarySniffBytes)
	n, _ := f.Read(buf)
	return bytes.IndexByte(buf[:n], 0) >= 0
}
