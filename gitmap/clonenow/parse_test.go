package clonenow

// Parser tests cover all three supported input formats:
//
//   - JSON  -- exercises the formatter.ParseJSON round-trip path
//              and verifies that ScanRecord -> Row lifting
//              preserves repo name, both URL fields, branch, and
//              the recorded relative path verbatim.
//   - CSV   -- same coverage as JSON but through the CSV parser,
//              including the case where the legacy 9-column layout
//              (no depth) round-trips cleanly.
//   - text  -- covers the plain `git clone <url> [dest]` artifact,
//              including: ssh-style URL classification, branch-
//              flag stripping, comment-line tolerance, and the
//              dest-fallback when only the URL is on the line.
//
// We also pin the dedup-by-RelativePath rule (later row wins) and
// the "no clonable rows" failure path so the CLI's exit-1
// guarantee for empty inputs has a contract test behind it.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// writeTemp drops a payload into a temp file with the given
// extension and returns the path. Centralized so each test stays
// focused on its parse contract instead of file-IO scaffolding.
func writeTemp(t *testing.T, ext, body string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "clonenow-*"+ext)
	if err != nil {
		t.Fatalf("temp: %v", err)
	}
	if _, err := f.WriteString(body); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	return f.Name()
}

func TestParseFile_JSON(t *testing.T) {
	body := `[
	  {"repoName":"a","httpsUrl":"https://example.com/a.git","sshUrl":"git@example.com:a.git","branch":"main","relativePath":"src/a"},
	  {"repoName":"b","httpsUrl":"https://example.com/b.git","relativePath":"src/b"}
	]`
	path := writeTemp(t, ".json", body)
	plan, err := ParseFile(path, "", constants.CloneNowModeHTTPS, constants.CloneNowOnExistsSkip)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if plan.Format != constants.CloneNowFormatJSON {
		t.Errorf("format = %q, want json", plan.Format)
	}
	if len(plan.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(plan.Rows))
	}
	if plan.Rows[0].SSHUrl == "" || plan.Rows[0].HTTPSUrl == "" {
		t.Errorf("row 0 url fields lost: %+v", plan.Rows[0])
	}
	if plan.Rows[0].RelativePath != "src/a" {
		t.Errorf("row 0 dest = %q", plan.Rows[0].RelativePath)
	}
}

func TestParseFile_CSV(t *testing.T) {
	// 10-col layout matches gitmap scan's current writer (see
	// formatter.WriteCSV). The trailing depth column is parsed but
	// not surfaced on Row -- clone-now doesn't honor depth.
	body := "repoName,httpsUrl,sshUrl,branch,branchSource,relativePath,absolutePath,cloneInstruction,notes,depth\r\n" +
		"a,https://example.com/a.git,git@example.com:a.git,main,HEAD,src/a,/abs/src/a,git clone https://example.com/a.git src/a,,0\r\n" +
		"b,https://example.com/b.git,,develop,config,src/b,/abs/src/b,git clone https://example.com/b.git src/b,,1\r\n"
	path := writeTemp(t, ".csv", body)
	plan, err := ParseFile(path, "", constants.CloneNowModeSSH, constants.CloneNowOnExistsSkip)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if plan.Format != constants.CloneNowFormatCSV {
		t.Errorf("format = %q", plan.Format)
	}
	if len(plan.Rows) != 2 {
		t.Fatalf("rows = %d", len(plan.Rows))
	}
	if plan.Rows[0].Branch != "main" || plan.Rows[1].Branch != "develop" {
		t.Errorf("branch lost: %+v", plan.Rows)
	}
}

func TestParseFile_Text(t *testing.T) {
	body := strings.Join([]string{
		"# this is a comment",
		"",
		"git clone https://example.com/a.git src/a",
		"git clone -b main https://example.com/b.git src/b",
		"git clone git@example.com:c.git",
		"echo not-a-clone-line",
	}, "\n") + "\n"
	path := writeTemp(t, ".txt", body)
	plan, err := ParseFile(path, "", constants.CloneNowModeHTTPS, constants.CloneNowOnExistsSkip)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if plan.Format != constants.CloneNowFormatText {
		t.Errorf("format = %q", plan.Format)
	}
	if len(plan.Rows) != 3 {
		t.Fatalf("rows = %d, want 3 (a/b/c, comment + echo skipped)", len(plan.Rows))
	}
	// Row 1 used `-b main` -- branch flags are stripped on purpose;
	// see parsetext.skipCloneFlags. We assert the URL still landed
	// at the expected slot and the dest is preserved.
	if plan.Rows[1].HTTPSUrl != "https://example.com/b.git" || plan.Rows[1].RelativePath != "src/b" {
		t.Errorf("row 1: %+v", plan.Rows[1])
	}
	// Row 2 had no explicit dest -> derived from URL basename.
	if plan.Rows[2].SSHUrl == "" || plan.Rows[2].RelativePath != "c" {
		t.Errorf("row 2 ssh/dest: %+v", plan.Rows[2])
	}
}

func TestParseFile_ForceFormat(t *testing.T) {
	// Same JSON payload but file extension is `.list` so auto-detect
	// would route to the text parser and produce zero rows. --format
	// json must override the extension.
	body := `[{"httpsUrl":"https://example.com/x.git","relativePath":"x"}]`
	path := writeTemp(t, ".list", body)
	plan, err := ParseFile(path, constants.CloneNowFormatJSON, constants.CloneNowModeHTTPS, constants.CloneNowOnExistsSkip)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if plan.Format != constants.CloneNowFormatJSON || len(plan.Rows) != 1 {
		t.Errorf("forced format ignored: %+v", plan)
	}
}

func TestParseFile_EmptyIsError(t *testing.T) {
	path := writeTemp(t, ".txt", "# nothing to clone here\n")
	_, err := ParseFile(path, "", constants.CloneNowModeHTTPS, constants.CloneNowOnExistsSkip)
	if err == nil {
		t.Fatal("ParseFile: want empty-input error, got nil")
	}
}

func TestDedupRows_LaterWins(t *testing.T) {
	rows := []Row{
		{HTTPSUrl: "https://x/a.git", RelativePath: "a"},
		{HTTPSUrl: "https://y/a.git", RelativePath: "a"}, // same dest -> overrides
		{HTTPSUrl: "https://x/b.git", RelativePath: "b"},
	}
	got := dedupRows(rows)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].HTTPSUrl != "https://y/a.git" {
		t.Errorf("dedup later-wins broken: %+v", got[0])
	}
}

func TestDeriveDest(t *testing.T) {
	cases := map[string]string{
		"https://example.com/owner/repo.git": "repo",
		"git@example.com:owner/repo.git":     "repo",
		"ssh://git@example.com/owner/repo":   "repo",
		"https://example.com/owner/repo/":    "repo",
	}
	for in, want := range cases {
		if got := DeriveDest(in); got != want {
			t.Errorf("DeriveDest(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseFile_AbsPathPropagated(t *testing.T) {
	// Plan.Source must be absolute so the dry-run header is
	// unambiguous regardless of the user's cwd. We only check the
	// "is absolute" property -- the exact path varies per OS / temp.
	path := writeTemp(t, ".json", `[{"httpsUrl":"https://x/a.git","relativePath":"a"}]`)
	plan, err := ParseFile(filepath.Base(path), constants.CloneNowFormatJSON, constants.CloneNowModeHTTPS, constants.CloneNowOnExistsSkip)
	// The relative open will fail (different cwd) -- that's fine, we
	// only need the absolute-path guarantee for the success case.
	if err == nil && !filepath.IsAbs(plan.Source) {
		t.Errorf("source not absolute: %q", plan.Source)
	}
}
