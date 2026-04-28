// Package cmd: audit-legacy scans the workspace for forbidden legacy
// strings (default: gitmap-v5 / gitmap-v6 / gitmap-v7) and exits 1 on
// any hit. Designed as a regression guard for remixes / rename commits.
package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alimtvnetwork/gitmap-v8/gitmap/constants"
)

// auditLegacyHit is one matched line.
type auditLegacyHit struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Pattern string `json:"pattern"`
	Text    string `json:"text"`
}

// auditLegacyOpts holds parsed CLI flags.
type auditLegacyOpts struct {
	Patterns []*regexp.Regexp
	Raw      []string
	Root     string
	AsJSON   bool
}

// runAuditLegacy is the dispatch entry point.
func runAuditLegacy(args []string) {
	checkHelp(constants.CmdAuditLegacy, args)
	opts, err := parseAuditLegacyArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	hits, n, walkErr := scanAuditLegacy(opts)
	if walkErr != nil {
		fmt.Fprintf(os.Stderr, constants.ErrAuditLegacyWalk, opts.Root, walkErr)
		os.Exit(2)
	}
	emitAuditLegacy(opts, hits, n)
	if len(hits) > 0 {
		os.Exit(1)
	}
}

// parseAuditLegacyArgs parses flags into an options struct.
func parseAuditLegacyArgs(args []string) (auditLegacyOpts, error) {
	fs := flag.NewFlagSet(constants.CmdAuditLegacy, flag.ContinueOnError)
	pats := fs.String(constants.FlagAuditLegacyPatterns,
		constants.DefaultAuditLegacyPatterns, constants.FlagDescAuditLegacyPatterns)
	root := fs.String(constants.FlagAuditLegacyPath, ".", constants.FlagDescAuditLegacyPath)
	asJSON := fs.Bool(constants.FlagAuditLegacyJSON, false, constants.FlagDescAuditLegacyJSON)
	if err := fs.Parse(args); err != nil {
		return auditLegacyOpts{}, err
	}
	compiled, raw, err := compileAuditPatterns(*pats)
	if err != nil {
		return auditLegacyOpts{}, err
	}

	return auditLegacyOpts{Patterns: compiled, Raw: raw, Root: *root, AsJSON: *asJSON}, nil
}

// compileAuditPatterns compiles a comma-separated pattern list.
func compileAuditPatterns(csv string) ([]*regexp.Regexp, []string, error) {
	parts := strings.Split(csv, ",")
	out := make([]*regexp.Regexp, 0, len(parts))
	raw := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		re, err := compileOnePattern(p)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, re)
		raw = append(raw, p)
	}

	return out, raw, nil
}

// compileOnePattern wraps regexp.Compile with a standardized error.
func compileOnePattern(p string) (*regexp.Regexp, error) {
	re, err := regexp.Compile(p)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrAuditLegacyRegex, p, err)
	}

	return re, nil
}

// scanAuditLegacy walks the root and returns hits + scanned-file count.
func scanAuditLegacy(opts auditLegacyOpts) ([]auditLegacyHit, int, error) {
	state := &auditWalkState{patterns: opts.Patterns}
	walkErr := filepath.WalkDir(opts.Root, state.visit)

	return state.hits, state.fileCount, walkErr
}

// auditWalkState accumulates results during filepath.WalkDir.
type auditWalkState struct {
	patterns  []*regexp.Regexp
	hits      []auditLegacyHit
	fileCount int
}

// visit is the WalkDir callback.
func (s *auditWalkState) visit(path string, d fs.DirEntry, err error) error {
	if err != nil || d == nil {
		return nil
	}
	if d.IsDir() {
		return skipAuditDir(d.Name())
	}
	if !isAuditScannable(path) {
		return nil
	}
	s.fileCount++
	s.hits = append(s.hits, scanAuditLegacyFile(path, s.patterns)...)

	return nil
}

// skipAuditDir returns fs.SkipDir for ignored directories.
func skipAuditDir(name string) error {
	switch name {
	case ".git", "node_modules", "dist", "build", "bin", ".next",
		".gitmap", "vendor", "coverage":
		return fs.SkipDir
	}

	return nil
}

// isAuditScannable filters out binary / non-text files by extension.
func isAuditScannable(path string) bool {
	skip := []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico",
		".pdf", ".zip", ".gz", ".tar", ".exe", ".dll", ".so",
		".dylib", ".bin", ".db", ".sqlite", ".woff", ".woff2", ".ttf"}
	low := strings.ToLower(path)
	for _, ext := range skip {
		if strings.HasSuffix(low, ext) {
			return false
		}
	}

	return true
}

// scanAuditLegacyFile reads one file and returns all matches.
func scanAuditLegacyFile(path string, pats []*regexp.Regexp) []auditLegacyHit {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	return collectAuditLineHits(path, scanner, pats)
}

// collectAuditLineHits iterates the scanner and gathers matches.
func collectAuditLineHits(path string, scanner *bufio.Scanner, pats []*regexp.Regexp) []auditLegacyHit {
	var hits []auditLegacyHit
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		hits = append(hits, matchAuditLine(path, lineNo, scanner.Text(), pats)...)
	}

	return hits
}

// matchAuditLine returns one hit per matching pattern on a single line.
func matchAuditLine(path string, lineNo int, line string, pats []*regexp.Regexp) []auditLegacyHit {
	var hits []auditLegacyHit
	for _, re := range pats {
		if re.MatchString(line) {
			hits = append(hits, auditLegacyHit{
				File: path, Line: lineNo, Pattern: re.String(), Text: line,
			})
		}
	}

	return hits
}
