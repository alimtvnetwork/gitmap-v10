package clonefrom

// Validation + dedup helpers shared by both parsers. Kept in a
// dedicated file so format-specific parsing (parse.go) stays
// focused and the validation rules have one obvious home for
// future tightening (e.g., adding a hostname allowlist).

import (
	"fmt"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// validateRow enforces per-row invariants:
//
//   - URL non-empty after trim
//   - URL has a recognizable scheme (https://, http://, git://,
//     ssh://) OR matches the scp-style `user@host:path` shape that
//     git accepts but net/url rejects
//   - Depth non-negative (negative depths are rejected by git but
//     the user-facing error from us is clearer than git's)
//
// Dest is NOT validated for path safety here — we let git fail
// loudly at execute time with its own (well-tested) checks. Adding
// a duplicate validator here would just create a second source of
// truth that could drift from git's actual behavior.
func validateRow(r Row) error {
	if len(r.URL) == 0 {

		return fmt.Errorf(constants.ErrCloneFromEmptyURL)
	}
	if !looksLikeGitURL(r.URL) {

		return fmt.Errorf(constants.ErrCloneFromBadURL, r.URL)
	}
	if r.Depth < 0 {

		return fmt.Errorf(constants.ErrCloneFromNegDepth, r.Depth)
	}

	return nil
}

// looksLikeGitURL is a permissive shape check. We deliberately do
// NOT use net/url.Parse because:
//
//   - scp-style `git@github.com:owner/repo.git` is valid for
//     `git clone` but parses as a relative path in net/url.
//   - we don't want to reject exotic but legal forms (file://,
//     git protocol over SSH tunnels, etc.) just because our
//     validator has a smaller world view than git itself.
//
// The check exists only to catch obvious typos like a bare
// "owner/repo" or a copy-pasted markdown link.
func looksLikeGitURL(s string) bool {
	if hasGitScheme(s) {

		return true
	}

	return looksLikeSCP(s)
}

// hasGitScheme returns true for any URL beginning with one of the
// schemes git supports natively.
func hasGitScheme(s string) bool {
	prefixes := []string{"https://", "http://", "ssh://", "git://", "file://"}
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {

			return true
		}
	}

	return false
}

// looksLikeSCP matches the `[user@]host:path` form. Cheap test:
// has a colon, no slashes before the colon, has at least one slash
// after the colon. Lets `git@github.com:owner/repo.git` through
// while rejecting `owner/repo.git` (no colon) and `://path` (slash
// before colon).
func looksLikeSCP(s string) bool {
	colon := strings.Index(s, ":")
	if colon <= 0 {

		return false
	}
	host := s[:colon]
	path := s[colon+1:]
	if strings.ContainsAny(host, "/\\") {

		return false
	}

	return strings.Contains(path, "/")
}

// dedupRows collapses rows with identical URL+Dest. Later rows
// overwrite earlier ones for branch/depth so users can re-list a
// URL further down the file to override a default — common
// spreadsheet workflow ("global defaults at top, exceptions at
// bottom"). Order of first occurrence is preserved so the dry-run
// preview lists rows in the user's original sequence.
func dedupRows(rows []Row) []Row {
	seen := make(map[string]int, len(rows))
	out := make([]Row, 0, len(rows))
	for _, r := range rows {
		key := r.URL + "\x00" + r.Dest
		if i, ok := seen[key]; ok {
			out[i] = mergeRows(out[i], r)
			continue
		}
		seen[key] = len(out)
		out = append(out, r)
	}

	return out
}

// mergeRows overlays the later row's optional fields onto the
// earlier row. Branch/depth take the later value if the later
// value is non-empty/non-zero.
func mergeRows(first, later Row) Row {
	out := first
	if len(later.Branch) > 0 {
		out.Branch = later.Branch
	}
	if later.Depth > 0 {
		out.Depth = later.Depth
	}

	return out
}
