// Package cloner re-clones repos from structured files.
package cloner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/formatter"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
)

// CloneFromFile reads a source file and clones all repos under targetDir.
func CloneFromFile(sourcePath, targetDir string, safePull bool) (model.CloneSummary, error) {
	records, err := loadRecords(sourcePath)
	if err != nil {
		return model.CloneSummary{}, err
	}

	return cloneAll(records, targetDir, safePull, false), nil
}

// CloneFromFileQuiet reads a source file and clones with suppressed progress.
func CloneFromFileQuiet(sourcePath, targetDir string, safePull bool) (model.CloneSummary, error) {
	records, err := loadRecords(sourcePath)
	if err != nil {
		return model.CloneSummary{}, err
	}

	return cloneAll(records, targetDir, safePull, true), nil
}

// loadRecords detects file format and parses records.
//
// Errors are wrapped with the source path so the CLI can surface
// "which file failed" without callers needing to know the original
// argument. Parser errors additionally carry their own line context
// (see parseTextFile) so users can jump straight to the offending row.
func loadRecords(path string) ([]model.ScanRecord, error) {
	ext := strings.ToLower(filepath.Ext(path))
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open clone source %q: %w", path, err)
	}
	defer file.Close()

	records, err := parseByExtension(ext, file)
	if err != nil {
		return nil, fmt.Errorf("parse clone source %q: %w", path, err)
	}

	return records, nil
}

// parseByExtension dispatches to the correct parser.
func parseByExtension(ext string, r io.Reader) ([]model.ScanRecord, error) {
	if ext == constants.ExtCSV {
		return formatter.ParseCSV(r)
	}
	if ext == constants.ExtJSON {
		return formatter.ParseJSON(r)
	}

	return parseTextFile(r)
}

// parseTextFile reads one git clone command per line. Scanner errors
// are wrapped with the line number of the last successfully read line
// so users can locate malformed input in long clone manifests.
func parseTextFile(r io.Reader) ([]model.ScanRecord, error) {
	var records []model.ScanRecord
	sc := bufio.NewScanner(r)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := strings.TrimSpace(sc.Text())
		if len(line) > 0 {
			records = append(records, parseCloneLine(line))
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read line %d: %w", lineNum+1, err)
	}

	return records, nil
}

// parseCloneLine extracts url, branch, path from a git clone command.
func parseCloneLine(line string) model.ScanRecord {
	parts := strings.Fields(line)
	rec := model.ScanRecord{CloneInstruction: line}
	if len(parts) >= 5 {
		rec.Branch = parts[3]
		rec.HTTPSUrl = parts[4]
	}
	if len(parts) >= 6 {
		rec.RelativePath = parts[5]
	}

	return rec
}

// cloneAll iterates records and clones each one with progress tracking.
func cloneAll(records []model.ScanRecord, targetDir string, safePull, quiet bool) model.CloneSummary {
	if !safePull && hasExistingRepos(records, targetDir) {
		safePull = true
		fmt.Print(constants.MsgAutoSafePull)
	}

	cache := LoadCloneCache(targetDir)
	progress := NewProgress(len(records), quiet)
	summary := model.CloneSummary{}

	for _, rec := range records {
		progress.Begin(repoDisplayName(rec))

		dest := filepath.Join(targetDir, rec.RelativePath)
		if cache.IsUpToDate(rec, dest) {
			result := model.CloneResult{Record: rec, Success: true}
			progress.Skip(result)
			summary = updateSummarySkipped(summary, result)
			continue
		}

		result := cloneOrPullOne(rec, targetDir, safePull)
		trackResult(progress, result, rec, targetDir, safePull)
		summary = updateSummary(summary, result)

		if result.Success {
			cache.Record(rec, dest)
		}
	}

	// Best-effort cache persistence — never fail the run on write errors.
	_ = cache.Save()

	progress.PrintSummary()

	return summary
}

// repoDisplayName returns a display name for progress output.
func repoDisplayName(rec model.ScanRecord) string {
	if len(rec.RepoName) > 0 {
		return rec.RepoName
	}

	return rec.RelativePath
}

// trackResult updates progress based on clone/pull outcome.
func trackResult(p *Progress, result model.CloneResult, rec model.ScanRecord, targetDir string, safePull bool) {
	if result.Success {
		pulled := safePull && isGitRepo(filepath.Join(targetDir, rec.RelativePath))
		p.Done(result, pulled)

		return
	}

	p.Fail(result)
}

// hasExistingRepos checks if any target repo directories already exist.
func hasExistingRepos(records []model.ScanRecord, targetDir string) bool {
	for _, rec := range records {
		dest := filepath.Join(targetDir, rec.RelativePath)
		if isGitRepo(dest) {
			return true
		}
	}

	return false
}

// cloneOne clones a single repository. Errors include the destination
// path and the record's RelativePath/RepoName so failures point straight
// at the offending row in the source manifest.
func cloneOne(rec model.ScanRecord, targetDir string) model.CloneResult {
	dest := filepath.Join(targetDir, rec.RelativePath)
	err := os.MkdirAll(filepath.Dir(dest), constants.DirPermission)
	if err != nil {
		msg := fmt.Sprintf("mkdir %q for %s: %v", filepath.Dir(dest), recordTag(rec), err)

		return model.CloneResult{Record: rec, Success: false, Error: msg}
	}

	return runClone(rec, dest)
}

// runClone executes the git clone command.
//
// The branch-selection strategy is driven by ScanRecord.BranchSource so
// that records captured in a detached or unknown state never produce
// "Remote branch not found" errors. When the source is trusted (HEAD,
// remote-tracking, default) the recorded branch is passed via -b; when it
// is untrusted (detached, unknown) git clone is invoked without -b and
// the remote's default HEAD decides the checkout.
//
// Failures are formatted with the URL, branch, destination, and record
// tag so a single error line is enough to identify which manifest row
// failed and why — no cross-referencing the source file required.
func runClone(rec model.ScanRecord, dest string) model.CloneResult {
	url := pickURL(rec)
	strategy := pickCloneStrategy(rec)

	args := []string{constants.GitClone}
	if strategy.useBranch {
		args = append(args, constants.GitBranchFlag, strategy.branch)
	}
	args = append(args, url, dest)

	cmd := exec.Command(constants.GitBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf(
			"git clone failed for %s: url=%q branch=%q dest=%q: %v: %s",
			recordTag(rec), url, strategy.branch, dest, err, strings.TrimSpace(string(out)),
		)

		return model.CloneResult{Record: rec, Success: false, Error: msg, Notes: strategy.reason}
	}

	return model.CloneResult{Record: rec, Success: true, Notes: strategy.reason}
}

// recordTag returns a short, log-friendly identifier for a record using
// the most specific field available. Used in error messages so users can
// locate the failing row in their clone manifest at a glance.
func recordTag(rec model.ScanRecord) string {
	switch {
	case len(rec.RepoName) > 0 && len(rec.RelativePath) > 0:
		return fmt.Sprintf("%s (%s)", rec.RepoName, rec.RelativePath)
	case len(rec.RepoName) > 0:
		return rec.RepoName
	case len(rec.RelativePath) > 0:
		return rec.RelativePath
	case len(rec.HTTPSUrl) > 0:
		return rec.HTTPSUrl
	case len(rec.SSHUrl) > 0:
		return rec.SSHUrl
	default:
		return "<unnamed record>"
	}
}

// pickURL selects the best available URL from a record.
func pickURL(rec model.ScanRecord) string {
	if len(rec.HTTPSUrl) > 0 {
		return rec.HTTPSUrl
	}

	return rec.SSHUrl
}

// updateSummary increments counters and collects results.
func updateSummary(s model.CloneSummary, r model.CloneResult) model.CloneSummary {
	if r.Success {
		s.Succeeded++
		s.Cloned = append(s.Cloned, r)

		return s
	}
	s.Failed++
	s.Errors = append(s.Errors, r)

	return s
}

// updateSummarySkipped records a cache-skipped repo: it counts toward
// Succeeded (the desired state is achieved) and is also tracked in
// Cloned + Skipped so downstream consumers (GitHub Desktop registration,
// reporting) treat it the same as a fresh clone.
func updateSummarySkipped(s model.CloneSummary, r model.CloneResult) model.CloneSummary {
	s.Succeeded++
	s.Cloned = append(s.Cloned, r)
	s.Skipped = append(s.Skipped, r)

	return s
}
