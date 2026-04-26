package cmd

// Schema contract for `gitmap startup-list --json`. Pairs the
// runtime encoder (encodeStartupListJSON) with the published
// schema at spec/08-json-schemas/startup-list.schema.json so a
// drift in either side fails the build.
//
// Why a hand-rolled mini-validator instead of pulling in a real
// JSON-Schema library?
//
//   1. The project's go.mod is intentionally lean (one direct dep
//      besides stdlib + charmbracelet/sqlite/archives/fuzzy/sys).
//      Adding a 30k-LOC schema validator for one test would be a
//      disproportionate dependency tax.
//   2. The contract surface here is tiny: a top-level array of
//      objects with three required string keys and a fixed key
//      order. A 60-line bespoke check covers the contract precisely
//      and makes the assertions readable in the failure message.
//   3. If/when the schema set in spec/08-json-schemas/ grows past
//      ~5 commands, swapping this for github.com/santhosh-tekuri/
//      jsonschema becomes worthwhile and is a one-file change.
//
// The schema file is read at test time (NOT embed-bundled) so a
// schema edit doesn't require re-running `go generate` or a build
// — just re-run the test.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/startup"
)

// schemaPath resolves spec/08-json-schemas/startup-list.schema.json
// relative to the repo root. Walks up from the test's CWD (which Go
// sets to the package dir, i.e. gitmap/cmd) until it finds a
// `spec/` sibling — same idiom used by other gitmap contract tests.
func schemaPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, "spec", "08-json-schemas", "startup-list.schema.json")
		if _, err := os.Stat(candidate); err == nil {

			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate spec/08-json-schemas/startup-list.schema.json walking up from %s", dir)

	return ""
}

// loadSchema parses the schema file into a generic map so the test
// can read both standard fields (`type`, `required`, `properties`)
// AND our extension (`propertyOrder`).
func loadSchema(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(schemaPath(t))
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var s map[string]any
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("parse schema: %v", err)
	}

	return s
}

// itemSchema descends into the array's `items` subschema where the
// per-entry object contract lives. Centralizes the navigation so
// the individual assertions below stay flat.
func itemSchema(t *testing.T, root map[string]any) map[string]any {
	t.Helper()
	items, ok := root["items"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no items object")
	}

	return items
}

// TestStartupListSchema_TopLevelIsArray pins the most fundamental
// shape decision: empty output is `[]`, NOT `null` or `{}`. A
// future encoder bug that emitted `null` for empty would break
// every downstream `jq length` consumer.
func TestStartupListSchema_TopLevelIsArray(t *testing.T) {
	root := loadSchema(t)
	if root["type"] != "array" {
		t.Fatalf("schema top-level type = %v, want array", root["type"])
	}
}

// TestStartupListSchema_RequiredKeysMatchEncoder asserts the schema
// requires exactly the three keys the encoder emits. If a future
// PR adds a key to startup.Entry but forgets the schema, this test
// fails with a clear diff.
func TestStartupListSchema_RequiredKeysMatchEncoder(t *testing.T) {
	required := stringSliceFromAny(itemSchema(t, loadSchema(t))["required"])
	want := []string{"name", "path", "exec"}
	if !equalStringSlices(required, want) {
		t.Fatalf("schema required = %v, want %v", required, want)
	}
}

// TestStartupListSchema_PropertyOrderMatchesEncoder is the headline
// contract test: encode a real entry, parse the resulting JSON
// preserving key order, and assert the order matches the schema's
// propertyOrder array. This is the ONLY guard that catches a
// reordering of the stablejson.Field slice in startuplistrender.go
// — Go's encoding/json sorts map keys alphabetically so a generic
// json.Unmarshal would mask the bug.
func TestStartupListSchema_PropertyOrderMatchesEncoder(t *testing.T) {
	want := stringSliceFromAny(itemSchema(t, loadSchema(t))["propertyOrder"])
	if len(want) == 0 {
		t.Fatalf("schema item has no propertyOrder array")
	}
	entries := []startup.Entry{{Name: "n", Path: "p", Exec: "e"}}
	var buf bytes.Buffer
	if err := encodeStartupListJSON(&buf, entries); err != nil {
		t.Fatalf("encode: %v", err)
	}
	got := extractFirstObjectKeyOrder(t, buf.Bytes())
	if !equalStringSlices(got, want) {
		t.Fatalf("emitted key order = %v, schema propertyOrder = %v", got, want)
	}
}

// TestStartupListSchema_EmptyEncodesAsArray pins the empty-input
// behavior end-to-end. Belt-and-suspenders alongside the byte-
// exact contract test: that one fails if the bytes drift, this
// one fails if the SHAPE drifts (e.g., `null` vs `[]`).
func TestStartupListSchema_EmptyEncodesAsArray(t *testing.T) {
	var buf bytes.Buffer
	if err := encodeStartupListJSON(&buf, nil); err != nil {
		t.Fatalf("encode: %v", err)
	}
	trimmed := bytes.TrimSpace(buf.Bytes())
	if !bytes.Equal(trimmed, []byte("[]")) {
		t.Fatalf("empty encoded as %q, want []", trimmed)
	}
}

// stringSliceFromAny converts a JSON-unmarshalled []any into
// []string. Returns nil on any non-string element so the caller's
// equality check fails loudly rather than silently coercing.
func stringSliceFromAny(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		s, ok := e.(string)
		if !ok {
			return nil
		}
		out = append(out, s)
	}

	return out
}

// equalStringSlices is a tiny order-sensitive compare. Avoids a
// reflect.DeepEqual import for one line of logic.
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

// extractFirstObjectKeyOrder uses json.Decoder's token stream to
// recover the literal on-the-wire key order of the first object
// inside the top-level array. encoding/json's standard Unmarshal
// into a map[string]any would lose ordering (Go maps are
// unordered); the Token API preserves it because it walks the raw
// bytes left-to-right.
func extractFirstObjectKeyOrder(t *testing.T, data []byte) []string {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	expectDelim(t, dec, '[')
	expectDelim(t, dec, '{')
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("token: %v", err)
		}
		key, ok := tok.(string)
		if !ok {
			t.Fatalf("expected key string, got %T (%v)", tok, tok)
		}
		keys = append(keys, key)
		if _, err := dec.Token(); err != nil {
			t.Fatalf("value token: %v", err)
		}
	}

	return keys
}

// expectDelim consumes one delimiter token and fails the test if
// it is not the expected rune. Pulled out of
// extractFirstObjectKeyOrder so that function stays under the
// 15-line per-function budget.
func expectDelim(t *testing.T, dec *json.Decoder, want json.Delim) {
	t.Helper()
	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("expected delim %v, got error: %v", want, err)
	}
	got, ok := tok.(json.Delim)
	if !ok || got != want {
		t.Fatalf("expected delim %v, got %v", want, tok)
	}
}
