package cmd

// E2E-style CSV serialization tests for `gitmap cn --all` / `--csv`
// batch mode under concurrency (v3.126.0+). These tests exercise
// the production CSV writers (writeReportRows + writeBatchReport)
// to prove byte-identical output regardless of pool size.
//
// Stub helpers + helpers live in
// clonenextbatchconcurrent_e2e_helpers_test.go.

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestE2E_BatchConcurrency_ByteIdenticalAcrossPoolSizes is the
// strongest determinism guarantee: regardless of worker count the
// CSV bytes produced by writeReportRows are identical.
func TestE2E_BatchConcurrency_ByteIdenticalAcrossPoolSizes(t *testing.T) {
	installStubProcessor(t)
	repos := makeRepoPaths(50)

	baseline := runAndSerialize(t, repos, 1)
	for _, workers := range []int{2, 4, 8, 16} {
		got := runAndSerialize(t, repos, workers)
		if !bytes.Equal(baseline, got) {
			t.Fatalf("CSV bytes differ at workers=%d (sequential vs parallel)\n--- want ---\n%s\n--- got ---\n%s",
				workers, baseline, got)
		}
	}
}

// runAndSerialize runs the concurrent pool then writes the CSV via
// the production writeReportRows into a temp file (exercising the
// real *os.File path).
func runAndSerialize(t *testing.T, repos []string, workers int) []byte {
	t.Helper()
	results := processBatchReposConcurrent(repos, workers, nil)

	tmp, err := os.CreateTemp(t.TempDir(), "cn-batch-*.csv")
	if err != nil {
		t.Fatalf("temp csv: %v", err)
	}
	writeReportRows(tmp, results)
	tmp.Close()

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("read back csv: %v", err)
	}
	return data
}

// TestE2E_BatchConcurrency_FullWriteBatchReport drives the
// production writeBatchReport (the real cn entrypoint helper) end
// to end: temp CWD, real file creation with the unix-second name,
// real CSV bytes. Asserts header + one row per repo + input
// ordering preserved.
func TestE2E_BatchConcurrency_FullWriteBatchReport(t *testing.T) {
	installStubProcessor(t)
	t.Chdir(t.TempDir())

	repos := makeRepoPaths(20)
	results := processBatchReposConcurrent(repos, 4, nil)

	reportPath := writeBatchReport(results)
	if reportPath == "" {
		t.Fatal("writeBatchReport returned empty path")
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != len(repos)+1 {
		t.Fatalf("line count: got %d, want %d (1 header + %d rows)",
			len(lines), len(repos)+1, len(repos))
	}
	if !strings.HasPrefix(lines[0], "repo,from,to,status,detail") {
		t.Fatalf("header line: got %q", lines[0])
	}
	for i := 0; i < len(repos); i++ {
		want := fmt.Sprintf("repo-%d", i)
		if !strings.Contains(lines[i+1], want) {
			t.Fatalf("row %d should contain %q, got %q", i, want, lines[i+1])
		}
	}
}
