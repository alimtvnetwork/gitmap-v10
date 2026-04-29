package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alimtvnetwork/gitmap-v9/gitmap/constants"
)

// TestBuildAuditNeedles verifies the dual-form contract for the audit
// scanner: every target version produces exactly two needles
// (`<base>-vN` and `<base>/vN`) in deterministic order.
func TestBuildAuditNeedles(t *testing.T) {
	got := buildAuditNeedles("gitmap", []int{4, 8})
	want := [][]byte{
		[]byte("gitmap-v4"), []byte("gitmap/v4"),
		[]byte("gitmap-v9"), []byte("gitmap/v9"),
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if !bytes.Equal(got[i], want[i]) {
			t.Errorf("needle[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestLineContainsAny is the primitive the audit scanner uses to decide
// whether a line is reportable. Tested independently so future
// optimizations cannot regress the contract.
func TestLineContainsAny(t *testing.T) {
	needles := [][]byte{[]byte("gitmap-v4"), []byte("gitmap/v9")}
	if !lineContainsAny([]byte("import gitmap-v4/foo"), needles) {
		t.Error("expected dash-form match")
	}
	if !lineContainsAny([]byte("module github.com/x/gitmap/v9"), needles) {
		t.Error("expected slash-form match")
	}
	if lineContainsAny([]byte("nothing relevant"), needles) {
		t.Error("unexpected match on clean line")
	}
}

// TestScanAuditFileFormatting locks the printed format to the spec's
// `path:line: matched-text` shape and confirms the per-file hit count.
func TestScanAuditFileFormatting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	body := "intro line\nsee gitmap-v4 for details\nclean line\nuse gitmap/v9 too\n"
	mustWriteFile(t, path, []byte(body))

	needles := [][]byte{[]byte("gitmap-v4"), []byte("gitmap/v9")}

	stdout, hits := captureStdout(t, func() int {
		return scanAuditFile(path, needles)
	})

	if hits != 2 {
		t.Fatalf("hits = %d, want 2", hits)
	}

	want2 := fmt.Sprintf(constants.MsgReplaceAuditMatch, path, 2, "see gitmap-v4 for details")
	want4 := fmt.Sprintf(constants.MsgReplaceAuditMatch, path, 4, "use gitmap/v9 too")
	if !strings.Contains(stdout, want2) {
		t.Errorf("missing line-2 hit\n got: %q\nwant substring: %q", stdout, want2)
	}
	if !strings.Contains(stdout, want4) {
		t.Errorf("missing line-4 hit\n got: %q\nwant substring: %q", stdout, want4)
	}
}

// captureStdout temporarily redirects os.Stdout so we can assert on
// fmt.Fprintf(os.Stdout, ...) output without coupling tests to a
// global writer.
func captureStdout(t *testing.T, fn func() int) (string, int) {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	result := fn()

	w.Close()
	os.Stdout = orig

	buf, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return string(buf), result
}
