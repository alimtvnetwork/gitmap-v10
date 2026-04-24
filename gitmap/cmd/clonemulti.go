package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/clonenext"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// flattenURLArgs splits each positional arg on commas, trims whitespace,
// drops empties, and returns the ordered, deduplicated list of URLs.
// Both space- and comma-separated forms are accepted, mixable:
//
//	gitmap clone a b c          → [a, b, c]
//	gitmap clone a,b,c          → [a, b, c]
//	gitmap clone a,b c d,e      → [a, b, c, d, e]
//
// Dedup is case-insensitive with trailing ".git" normalised; first-seen wins.
// See: spec/01-app/104-clone-multi.md and mem://features/clone-multi.
func flattenURLArgs(args []string) []string {
	out := make([]string, 0, len(args))
	seen := make(map[string]struct{}, len(args))

	for _, raw := range args {
		for _, part := range strings.Split(raw, ",") {
			cleaned := strings.TrimSpace(part)
			if cleaned == "" {
				continue
			}

			key := normaliseURLKey(cleaned)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, cleaned)
		}
	}

	return out
}

// normaliseURLKey lowercases and trims a trailing ".git" so that
// "https://x/y" and "HTTPS://X/Y.git" dedupe to the same key.
func normaliseURLKey(url string) string {
	lower := strings.ToLower(url)
	lower = strings.TrimSuffix(lower, ".git")

	return lower
}

// classifyURLs partitions a flattened arg list into valid URLs and
// invalid (non-URL, non-empty) entries. Invalid entries are reported
// to stderr but do NOT abort the batch — the caller decides exit code.
func classifyURLs(flat []string) (valid, invalid []string) {
	for _, candidate := range flat {
		if isDirectURL(candidate) {
			valid = append(valid, candidate)

			continue
		}

		invalid = append(invalid, candidate)
		fmt.Fprintf(os.Stderr, constants.MsgCloneInvalidURLFmt, candidate)
	}

	return valid, invalid
}

// executeDirectCloneOne is the non-fatal sibling of executeDirectClone:
// it clones a single URL, persists to the DB, optionally registers with
// GitHub Desktop, and returns any error instead of calling os.Exit.
// Folder name is auto-derived (versioned URLs flatten via clonenext).
func executeDirectCloneOne(url, folderName string, ghDesktopFlag, noReplace bool) error {
	repoName := repoNameFromURL(url)
	folderName = resolveCloneFolder(repoName, folderName)

	absPath, err := filepath.Abs(folderName)
	if err != nil {
		return fmt.Errorf("resolve abs path for %s: %w", folderName, err)
	}

	if noReplace {
		if _, statErr := os.Stat(absPath); statErr == nil {
			return fmt.Errorf("target exists: %s (use without --no-replace to replace)", absPath)
		}
		if cloneErr := runCloneCommand(url, absPath); cloneErr != nil {
			return fmt.Errorf("git clone: %w", cloneErr)
		}
	} else {
		if _, replaceErr := cloneReplacing(url, absPath); replaceErr != nil {
			return fmt.Errorf("clone-replace: %w", replaceErr)
		}
	}

	upsertDirectClone(url, repoName, folderName, absPath)

	if ghDesktopFlag {
		registerSingleDesktop(repoName, absPath)
		fmt.Printf(constants.MsgCloneRegisteredInline, repoName)
	}

	return nil
}

// resolveCloneFolder derives the destination folder name when none is given,
// auto-flattening versioned URLs (e.g., wp-onboarding-v13 → wp-onboarding/).
func resolveCloneFolder(repoName, folderName string) string {
	if len(folderName) > 0 {
		return folderName
	}

	parsed := clonenext.ParseRepoName(repoName)
	if parsed.HasVersion {
		return parsed.BaseName
	}

	return repoName
}
