package cmd

// Reusable helpers for JSON snapshot tests (extracted from
// startuplistjson_snapshot_test.go to keep both files under the
// 200-line code-style budget). Future `--format=json` snapshot
// tests should land their helpers here too rather than spawning
// per-feature helper files.
//
// Design contract:
//
//   - Token-stream parsing (NOT json.Unmarshal into map[string]any),
//     so on-the-wire key ORDER is observable. A map-based parse
//     would silently lose the very property these tests exist to
//     pin down.
//
//   - Per-object failure messages name the offending object index
//     and the missing/unexpected/reordered key. A schema regression
//     in row 17 of a 100-row output points at row 17, not at the
//     whole blob.
//
//   - No dependency on any specific encoder or schema — these
//     helpers operate on `[]byte` of already-rendered JSON, so they
//     work for stablejson, encoding/json, or any future encoder.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/startup"
)

// mustEncodeStartupList runs encodeStartupListJSON on `entries` and
// returns a COPY of the produced bytes (so a caller comparing two
// runs cannot accidentally alias a reused buffer). Lives in the
// shared helpers file rather than per-test so any future
// startup-list-shaped snapshot can reuse it without duplicating
// the encoder-call + buffer-copy boilerplate.
func mustEncodeStartupList(t *testing.T, entries []startup.Entry) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := encodeStartupListJSON(&buf, entries); err != nil {
		t.Fatalf("encode: %v", err)
	}

	return append([]byte(nil), buf.Bytes()...)
}

// assertEveryObjectKeysExact parses `raw` as a top-level JSON array
// and asserts that EVERY object in it has exactly `want` keys, in
// `want` order. Reports per-object failures so a schema regression
// in row 17 of a 100-row output points at row 17, not at the whole
// blob. Uses the same json.Decoder.Token() path as the existing
// assertObjectKeyOrder helper so on-the-wire ordering is checked,
// never a map-shuffled view of it.
func assertEveryObjectKeysExact(t *testing.T, raw []byte, want []string) {
	t.Helper()
	keysPerObject := readEveryObjectKeys(t, raw)
	if len(keysPerObject) == 0 {
		t.Fatalf("expected at least one object, got zero")
	}
	for i, got := range keysPerObject {
		assertObjectKeysExactAt(t, i, got, want)
	}
}

// assertObjectKeysExactAt does the per-object comparison with three
// targeted failure modes (missing, unexpected, reordered) so the
// test output tells the developer exactly what kind of schema drift
// happened. Split out so assertEveryObjectKeysExact stays under the
// 15-line function budget.
func assertObjectKeysExactAt(t *testing.T, idx int, got, want []string) {
	t.Helper()
	wantSet := stringSet(want)
	gotSet := stringSet(got)
	for _, k := range want {
		if !gotSet[k] {
			t.Errorf("missing key %q in object[%d] (got keys %v)", k, idx, got)
		}
	}
	for _, k := range got {
		if !wantSet[k] {
			t.Errorf("unexpected key %q in object[%d] (want keys %v)", k, idx, want)
		}
	}
	if !equalStringSlices(got, want) && len(got) == len(want) {
		t.Errorf("key order drift in object[%d]\n  want: %v\n  got:  %v", idx, want, got)
	}
}

// readEveryObjectKeys streams the entire top-level array and returns
// one []string of keys per object. Returns an empty outer slice for
// `[]` so callers can distinguish "no objects" from "malformed".
func readEveryObjectKeys(t *testing.T, raw []byte) [][]string {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := expectDelim(dec, '['); err != nil {
		t.Fatalf("expected top-level array: %v", err)
	}
	var out [][]string
	for dec.More() {
		if err := expectDelim(dec, '{'); err != nil {
			t.Fatalf("expected object at index %d: %v", len(out), err)
		}
		out = append(out, collectObjectKeys(t, dec))
	}

	return out
}

// expectDelim reads the next token and confirms it's the requested
// JSON delimiter (`[`, `]`, `{`, or `}`). One helper instead of
// per-delimiter wrappers keeps the helpers file under budget while
// keeping failure messages specific (the caller adds context).
func expectDelim(dec *json.Decoder, want byte) error {
	tok, err := dec.Token()
	if err != nil {

		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok || rune(delim) != rune(want) {

		return fmt.Errorf("want delim %q, got %v (%T)", want, tok, tok)
	}

	return nil
}

// collectObjectKeys reads key-value pairs from dec until the
// closing `}` is consumed, returning just the key names in
// wire order. dec must already have consumed the opening `{`.
func collectObjectKeys(t *testing.T, dec *json.Decoder) []string {
	t.Helper()
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("reading object key: %v", err)
		}
		key, ok := tok.(string)
		if !ok {
			t.Fatalf("expected string key, got %v (%T)", tok, tok)
		}
		keys = append(keys, key)
		// Skip the value without decoding its type.
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			t.Fatalf("skipping value for key %q: %v", key, err)
		}
	}
	// Consume the closing '}'.
	if _, err := dec.Token(); err != nil {
		t.Fatalf("expected closing '}': %v", err)
	}
	return keys
}

// equalStringSlices returns true when a and b have identical length
// and identical element values at every index.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// stringSet is the tiny set helper used by the missing/unexpected
// passes. Inline construction keeps the schema check zero-allocation
// on the success path (Go maps with <8 entries fit in one bucket).
func stringSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}

	return m
}
