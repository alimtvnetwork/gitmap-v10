// Package mapper converts raw scan data into ScanRecord structs.
package mapper

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/gitutil"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/scanner"
)

// BuildRecords converts a list of RepoInfo into ScanRecords.
func BuildRecords(repos []scanner.RepoInfo, mode, defaultNote string) []model.ScanRecord {
	records := make([]model.ScanRecord, 0, len(repos))
	for _, repo := range repos {
		rec := buildOneRecord(repo, mode, defaultNote)
		records = append(records, rec)
	}

	return records
}

// buildOneRecord creates a single ScanRecord from a RepoInfo.
func buildOneRecord(repo scanner.RepoInfo, mode, note string) model.ScanRecord {
	remoteURL, _ := gitutil.RemoteURL(repo.AbsolutePath)
	branch, branchSource := gitutil.DetectBranch(repo.AbsolutePath)
	httpsURL := toHTTPS(remoteURL)
	sshURL := toSSH(remoteURL)
	cloneURL := selectCloneURL(httpsURL, sshURL, mode)
	repoName := extractRepoName(remoteURL)
	noteText := buildNote(remoteURL, note)
	instruction := buildInstruction(cloneURL, branch, repo.RelativePath)

	return model.ScanRecord{
		Slug: buildSlug(httpsURL, repoName),
		RepoName: repoName, HTTPSUrl: httpsURL, SSHUrl: sshURL,
		Branch: branch, BranchSource: branchSource,
		RelativePath: repo.RelativePath, AbsolutePath: repo.AbsolutePath,
		CloneInstruction: instruction, Notes: noteText,
	}
}

// toHTTPS converts a remote URL to HTTPS format.
func toHTTPS(raw string) string {
	if strings.HasPrefix(raw, constants.PrefixHTTPS) {
		return raw
	}
	if strings.HasPrefix(raw, constants.PrefixSSH) {
		host, path := splitSSH(raw)

		return fmt.Sprintf(constants.HTTPSFromSSHFmt, host, path)
	}

	return raw
}

// toSSH converts a remote URL to SSH format.
func toSSH(raw string) string {
	if strings.HasPrefix(raw, constants.PrefixSSH) {
		return raw
	}
	if strings.HasPrefix(raw, constants.PrefixHTTPS) {
		trimmed := strings.TrimPrefix(raw, constants.PrefixHTTPS)
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) == 2 {
			return fmt.Sprintf(constants.SSHFromHTTPSFmt, parts[0], parts[1])
		}
	}

	return raw
}

// splitSSH splits a git@host:path URL into host and path.
func splitSSH(raw string) (string, string) {
	trimmed := strings.TrimPrefix(raw, constants.PrefixSSH)
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return trimmed, ""
}

// selectCloneURL picks HTTPS or SSH URL based on mode.
func selectCloneURL(httpsURL, sshURL, mode string) string {
	if mode == constants.ModeSSH {
		return sshURL
	}

	return httpsURL
}

// extractRepoName derives the repository name from a remote URL.
func extractRepoName(raw string) string {
	if len(raw) == 0 {
		return constants.UnknownRepoName
	}
	base := filepath.Base(raw)

	return strings.TrimSuffix(base, constants.ExtGit)
}

// buildNote generates the notes field for a record.
func buildNote(remoteURL, defaultNote string) string {
	if len(remoteURL) == 0 {
		return constants.NoteNoRemote
	}

	return defaultNote
}

// buildInstruction creates the full git clone command string.
func buildInstruction(url, branch, relPath string) string {
	if len(url) == 0 {
		return ""
	}

	return fmt.Sprintf(constants.CloneInstructionFmt, branch, url, relPath)
}

// buildSlug derives a lowercase slug from the HTTPS URL.
// Falls back to repoName when the URL is empty.
func buildSlug(httpsURL, repoName string) string {
	if len(httpsURL) == 0 {
		return strings.ToLower(repoName)
	}
	base := filepath.Base(httpsURL)
	trimmed := strings.TrimSuffix(base, constants.ExtGit)

	return strings.ToLower(trimmed)
}
