package cmd

// E2E-style ordering + count tests for `gitmap cn --all` / `--csv`
// batch mode under concurrency (v3.126.0+). These tests stub
// processOneBatchRepoFn (see clonenextbatchconcurrent_e2e_helpers_test.go)
// so they run in milliseconds without real git repos.
//
// CSV byte-equivalence + full-write-report tests live in the
// sibling clonenextbatchconcurrent_e2e_csv_test.go file.

import (
	"sync/atomic"
	"testing"
)

// TestE2E_BatchConcurrency_DeterministicOrdering proves that under
// --max-concurrency=8 over 50 repos, the result slice is still in
// input order even though the stub deliberately makes workers
// finish out-of-order.
func TestE2E_BatchConcurrency_DeterministicOrdering(t *testing.T) {
	peak := installStubProcessor(t)
	repos := makeRepoPaths(50)

	results := processBatchReposConcurrent(repos, 8, nil)

	if len(results) != len(repos) {
		t.Fatalf("results length %d != input %d", len(results), len(repos))
	}
	for i, r := range results {
		if r.RepoPath != repos[i] {
			t.Fatalf("ordering drift at index %d: got %s, want %s",
				i, r.RepoPath, repos[i])
		}
	}
	if atomic.LoadInt64(peak) < 2 {
		t.Fatalf("pool never went parallel (peak inflight = %d) — test is invalid",
			atomic.LoadInt64(peak))
	}
}

// TestE2E_BatchConcurrency_StatusCountsExact verifies the per-bucket
// totals exactly match what the stub emitted: 50 repos with last
// digits 0-9 cycling 5 times → 20 ok, 20 failed, 10 skipped.
func TestE2E_BatchConcurrency_StatusCountsExact(t *testing.T) {
	installStubProcessor(t)
	repos := makeRepoPaths(50)

	results := processBatchReposConcurrent(repos, 4, nil)
	ok, failed, skipped := tallyBatch(results)

	const wantOK, wantFailed, wantSkipped = 20, 20, 10
	if ok != wantOK || failed != wantFailed || skipped != wantSkipped {
		t.Fatalf("counts: ok=%d failed=%d skipped=%d, want %d/%d/%d",
			ok, failed, skipped, wantOK, wantFailed, wantSkipped)
	}
}

// TestE2E_BatchConcurrency_ProgressCallbackFires verifies the
// onResult callback handed to the collector receives exactly one
// invocation per repo, with the same total count as the returned
// slice — proving the live progress reporter sees a faithful view
// even under heavy concurrency.
func TestE2E_BatchConcurrency_ProgressCallbackFires(t *testing.T) {
	installStubProcessor(t)
	repos := makeRepoPaths(30)

	var seenCount int64
	cb := func(_ batchRowResult) {
		atomic.AddInt64(&seenCount, 1)
	}

	results := processBatchReposConcurrent(repos, 6, cb)

	if got := atomic.LoadInt64(&seenCount); int(got) != len(repos) {
		t.Fatalf("callback fired %d times, want %d", got, len(repos))
	}
	if len(results) != len(repos) {
		t.Fatalf("results length: got %d, want %d", len(results), len(repos))
	}
}
