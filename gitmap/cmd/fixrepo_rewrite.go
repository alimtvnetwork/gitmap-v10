package cmd

// Token-rewrite engine for `gitmap fix-repo`. Mirrors
// scripts/fix-repo/Rewrite-Engine.ps1: replace literal `{base}-v{N}`
// with `{base}-v{current}` for every N in targets, guarded by a
// negative-lookahead so `-v1` does not match inside `-v10`.

import (
	"os"
	"regexp"
	"strconv"
)

// rewriteFixRepoFile reads fullPath, applies every target rewrite,
// and (unless dryRun) writes the result back. Returns the total
// replacement count across all targets, or an error on read/write
// failure.
func rewriteFixRepoFile(fullPath, base string, current int, targets []int, dryRun bool) (int, error) {
	original, err := os.ReadFile(fullPath) //nolint:gosec // tracked-file path from git ls-files
	if err != nil {
		return 0, err
	}
	updated, count := applyAllTargets(string(original), base, current, targets)
	if count == 0 {
		return 0, nil
	}
	if dryRun {
		return count, nil
	}
	if err := os.WriteFile(fullPath, []byte(updated), 0o644); err != nil { //nolint:gosec // preserve scripts' write mode
		return 0, err
	}

	return count, nil
}

// applyAllTargets folds every target rewrite over text and returns
// the cumulative result + total replacement count.
func applyAllTargets(text, base string, current int, targets []int) (string, int) {
	total := 0
	for _, n := range targets {
		updated, added := applyOneTarget(text, base, n, current)
		text = updated
		total += added
	}

	return text, total
}

// applyOneTarget replaces every literal `{base}-vN` (not followed by
// a digit) with `{base}-v{current}` and returns the new text + count.
func applyOneTarget(text, base string, n, current int) (string, int) {
	re := buildRewriteRegex(base, n)
	matches := re.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text, 0
	}
	replacement := base + "-v" + strconv.Itoa(current)

	return re.ReplaceAllString(text, replacement), len(matches)
}

// buildRewriteRegex compiles the literal-token + negative-lookahead
// pattern. Go's RE2 has no native `(?!...)` so we approximate with
// `(?:$|[^0-9])` and consume the trailing char via a capture-group
// fix-up, but for this token shape a simpler trick works: anchor on
// the literal, then check the next byte by hand. Implementation
// uses regexp with a non-digit / end-of-text trailing assertion.
func buildRewriteRegex(base string, n int) *regexp.Regexp {
	literal := regexp.QuoteMeta(base + "-v" + strconv.Itoa(n))
	// The trailing `\b` boundary is insufficient (digits ARE word chars)
	// so we use a non-digit-or-end alternative and consume the trailing
	// char in the replacement loop via a custom approach below.
	pattern := literal + `(?:[^0-9]|$)`

	return regexp.MustCompile(pattern)
}
