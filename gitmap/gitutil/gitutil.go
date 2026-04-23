// Package gitutil extracts Git metadata by running git commands.
package gitutil

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// RepoStatus holds the live state of a Git repository.
type RepoStatus struct {
	Branch       string
	Dirty        bool
	Untracked    int
	Modified     int
	Staged       int
	Ahead        int
	Behind       int
	StashCount   int
	Unreachable  bool
}

// RemoteURL returns the origin remote URL for a repo at the given path.
func RemoteURL(repoPath string) (string, error) {
	out, err := runGit(repoPath,
		constants.GitConfigCmd, constants.GitGetFlag, constants.GitRemoteOrigin)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// Branch source labels describe how a repo's branch was determined.
const (
	BranchSourceHEAD            = "HEAD"
	BranchSourceDetached        = "detached"
	BranchSourceRemoteTracking  = "remote-tracking"
	BranchSourceDefault         = "default"
	BranchSourceUnknown         = "unknown"
)

// CurrentBranch returns the current branch name for a repo.
func CurrentBranch(repoPath string) (string, error) {
	out, err := runGit(repoPath,
		constants.GitRevParse, constants.GitAbbrevRef, constants.GitHEAD)
	if err != nil {
		return constants.DefaultBranch, err
	}

	return strings.TrimSpace(out), nil
}

// DetectBranch returns the branch name and a label describing how it was
// detected. Resolution order:
//  1. HEAD via `git rev-parse --abbrev-ref HEAD` — labeled "HEAD" when on a
//     named branch, or "detached" when HEAD points directly at a commit.
//  2. Remote-tracking branch via `git symbolic-ref refs/remotes/origin/HEAD`
//     — labeled "remote-tracking".
//  3. Built-in default branch — labeled "default".
// If no resolution succeeds the returned source is "unknown".
func DetectBranch(repoPath string) (branch, source string) {
	out, err := runGit(repoPath,
		constants.GitRevParse, constants.GitAbbrevRef, constants.GitHEAD)
	if err == nil {
		name := strings.TrimSpace(out)
		if name == constants.GitHEAD {
			return name, BranchSourceDetached
		}
		if len(name) > 0 {
			return name, BranchSourceHEAD
		}
	}

	out, err = runGit(repoPath,
		"symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		ref := strings.TrimSpace(out)
		const prefix = "refs/remotes/origin/"
		if strings.HasPrefix(ref, prefix) {
			return strings.TrimPrefix(ref, prefix), BranchSourceRemoteTracking
		}
	}

	if len(constants.DefaultBranch) > 0 {
		return constants.DefaultBranch, BranchSourceDefault
	}

	return "", BranchSourceUnknown
}

// Status returns the full live status of a repository.
// If the path does not exist or is not a git repo, Unreachable is set.
func Status(repoPath string) RepoStatus {
	rs := RepoStatus{}

	if _, err := os.Stat(repoPath); err != nil {
		rs.Unreachable = true
		return rs
	}

	branch, err := CurrentBranch(repoPath)
	if err != nil {
		rs.Unreachable = true
		return rs
	}

	rs.Branch = branch
	rs.Dirty, rs.Untracked, rs.Modified, rs.Staged = parsePortcelainStatus(repoPath)
	rs.Ahead, rs.Behind = parseAheadBehind(repoPath)
	rs.StashCount = countStashes(repoPath)

	return rs
}

// parsePortcelainStatus runs git status --porcelain and counts file states.
func parsePortcelainStatus(repoPath string) (dirty bool, untracked, modified, staged int) {
	out, err := runGit(repoPath, "status", "--porcelain")
	if err != nil {
		return false, 0, 0, 0
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		x := line[0]
		y := line[1]
		if x == '?' && y == '?' {
			untracked++
		} else if x != ' ' && x != '?' {
			staged++
		}
		if y != ' ' && y != '?' {
			modified++
		}
	}
	dirty = (untracked + modified + staged) > 0

	return dirty, untracked, modified, staged
}

// parseAheadBehind extracts ahead/behind counts from rev-list.
func parseAheadBehind(repoPath string) (ahead, behind int) {
	out, err := runGit(repoPath, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) == 2 {
		ahead, _ = strconv.Atoi(parts[0])
		behind, _ = strconv.Atoi(parts[1])
	}

	return ahead, behind
}

// countStashes returns the number of stash entries.
func countStashes(repoPath string) int {
	out, err := runGit(repoPath, "stash", "list")
	if err != nil || len(strings.TrimSpace(out)) == 0 {
		return 0
	}

	return len(strings.Split(strings.TrimSpace(out), "\n"))
}

// runGit executes a git command in the given directory and returns stdout.
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command(constants.GitBin, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
