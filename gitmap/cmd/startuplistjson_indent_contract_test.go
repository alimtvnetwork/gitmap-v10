package cmd

// Contract for `gitmap startup-list --json-indent=N`. Pins three
// invariants:
//
//  1. Indent value is whitespace-only — key names and key ORDER
//     are byte-identical across every accepted indent (0..8). A
//     future change that accidentally sorts keys when minifying
//     (e.g. by routing through map[string]any) would fail this.
//
//  2. Indent=0 produces a single-line minified document with no
//     inter-token whitespace. Matches `jq -c` framing.
//
//  3. Indent=2 (the default) produces output byte-identical to the
//     legacy encodeStartupListJSON pretty form — sibling golden
//     fixtures for the default indent continue to pass without
//     regeneration.
//
//  4. Empty list always emits `[]\n` regardless of indent — the
//     length-check contract every jq pipeline depends on.
//
// These are the FOUR things a downstream consumer needs to be true
// to safely set --json-indent in any pipeline. Anything else (exact
// space placement, escape behavior) is delegated to stablejson and
// covered by stablejson_test.go and the existing JSON byte-exact
// tests.

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/startup"
)

// TestStartupListJSONIndent_KeyOrderStableAcrossIndents walks
// every accepted indent value, parses the output back into a
// streamed token sequence, and asserts the keys appear in the
// canonical order (name, path, exec) in every variant. Catches any
// future regression where a minifier path accidentally walks fields
// in a different order than the pretty path.
func TestStartupListJSONIndent_KeyOrderStableAcrossIndents(t *testing.T) {
	entries := []startup.Entry{
		{Name: "gitmap-a", Path: "/p/a", Exec: "/bin/a"},
		{Name: "gitmap-b", Path: "/p/b", Exec: "/bin/b"},
	}
	wantKeys := []string{"name", "path", "exec"}
	for _, indent := range []int{0, 1, 2, 4, 8} {
		var buf bytes.Buffer
		if err := encodeStartupListJSONIndent(&buf, entries, indent); err != nil {
			t.Fatalf("indent=%d: encode: %v", indent, err)
		}
		assertJSONArrayKeyOrder(t, indent, buf.Bytes(), wantKeys, len(entries))
	}
}

// assertJSONArrayKeyOrder walks a JSON array and asserts each
// object's keys appear in the expected order. Uses Decoder.Token
// because a regular Unmarshal into map[string]any loses order.
// Split out so the parent test stays under the 15-line budget.
func assertJSONArrayKeyOrder(t *testing.T, indent int, data []byte, want []string, wantObjects int) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	if _, err := dec.Token(); err != nil { // opening `[`
		t.Fatalf("indent=%d: open array: %v", indent, err)
	}
	for objIdx := 0; objIdx < wantObjects; objIdx++ {
		if _, err := dec.Token(); err != nil { // opening `{`
			t.Fatalf("indent=%d obj %d: open: %v", indent, objIdx, err)
		}
		for keyIdx, wantKey := range want {
			tok, err := dec.Token()
			if err != nil {
				t.Fatalf("indent=%d obj %d key %d: %v", indent, objIdx, keyIdx, err)
			}
			if tok != wantKey {
				t.Fatalf("indent=%d obj %d key %d: want %q got %v",
					indent, objIdx, keyIdx, wantKey, tok)
			}
			if _, err := dec.Token(); err != nil { // value
				t.Fatalf("indent=%d obj %d val %d: %v", indent, objIdx, keyIdx, err)
			}
		}
		if _, err := dec.Token(); err != nil { // closing `}`
			t.Fatalf("indent=%d obj %d: close: %v", indent, objIdx, err)
		}
	}
}

// TestStartupListJSONIndent_MinifiedByteExact pins the exact bytes
// for indent=0. Catches any drift in inter-token whitespace (a
// stray space after `:` or `,` would silently bloat a high-volume
// pipeline) and the trailing-newline rule.
func TestStartupListJSONIndent_MinifiedByteExact(t *testing.T) {
	entries := []startup.Entry{
		{Name: "gitmap-a", Path: "/p/a", Exec: "/bin/a"},
		{Name: "gitmap-b", Path: "/p/b", Exec: ""},
	}
	var buf bytes.Buffer
	if err := encodeStartupListJSONIndent(&buf, entries, 0); err != nil {
		t.Fatalf("encode: %v", err)
	}
	want := `[{"name":"gitmap-a","path":"/p/a","exec":"/bin/a"},` +
		`{"name":"gitmap-b","path":"/p/b","exec":""}]` + "\n"

	if got := buf.String(); got != want {
		t.Fatalf("byte mismatch\nwant: %q\ngot:  %q", want, got)
	}
}

// TestStartupListJSONIndent_DefaultMatchesLegacy proves indent=2
// (the new default) produces output byte-identical to the no-arg
// encodeStartupListJSON wrapper. Without this guarantee the
// existing 7+ JSON golden fixtures would all need regeneration —
// the whole point of the wrapper is to keep them passing untouched.
func TestStartupListJSONIndent_DefaultMatchesLegacy(t *testing.T) {
	entries := []startup.Entry{
		{Name: "gitmap-a", Path: "/p/a", Exec: "/bin/a"},
		{Name: "gitmap-b", Path: "/p/b", Exec: "/bin/b --flag"},
	}
	var legacyBuf, indentBuf bytes.Buffer
	if err := encodeStartupListJSON(&legacyBuf, entries); err != nil {
		t.Fatalf("legacy: %v", err)
	}
	if err := encodeStartupListJSONIndent(&indentBuf, entries, 2); err != nil {
		t.Fatalf("indent=2: %v", err)
	}
	if !bytes.Equal(legacyBuf.Bytes(), indentBuf.Bytes()) {
		t.Fatalf("default-indent drift\n--- legacy\n%s--- indent=2\n%s",
			legacyBuf.String(), indentBuf.String())
	}
}

// TestStartupListJSONIndent_EmptyAlwaysBracketsNewline asserts the
// `[]\n` empty-list contract holds across every indent setting.
// Critical for jq pipelines that do `length` — they must work the
// same whether the user passed --json-indent=0 or =4.
func TestStartupListJSONIndent_EmptyAlwaysBracketsNewline(t *testing.T) {
	for _, indent := range []int{0, 1, 2, 4, 8} {
		var buf bytes.Buffer
		if err := encodeStartupListJSONIndent(&buf, nil, indent); err != nil {
			t.Fatalf("indent=%d: %v", indent, err)
		}
		if got := buf.String(); got != "[]\n" {
			t.Fatalf("indent=%d empty: want %q got %q",
				indent, "[]\n", got)
		}
	}
}

// TestStartupListJSONIndent_IndentWidthIsCountedSpaces verifies
// the per-line prefix at indent=N really is N spaces (not tabs,
// not N-1, not the json.Encoder N+1 quirk). Uses a single-entry
// document so the indented `"name"` line has a predictable prefix.
func TestStartupListJSONIndent_IndentWidthIsCountedSpaces(t *testing.T) {
	entries := []startup.Entry{
		{Name: "gitmap-a", Path: "/p/a", Exec: "/bin/a"},
	}
	for _, indent := range []int{1, 2, 4, 8} {
		var buf bytes.Buffer
		if err := encodeStartupListJSONIndent(&buf, entries, indent); err != nil {
			t.Fatalf("indent=%d: %v", indent, err)
		}
		// Per writeObject contract: outer brace at 1×indent, key
		// lines at 2×indent. Look for the key line specifically.
		wantPrefix := "\n" + strings.Repeat(" ", indent*2) + `"name"`
		if !strings.Contains(buf.String(), wantPrefix) {
			t.Fatalf("indent=%d: missing %q in:\n%s",
				indent, wantPrefix, buf.String())
		}
	}
}

// TestStartupListJSONIndent_FlagParsing covers the two cases the
// flag parser must reject (negative + too-large) and the boundary
// values that must succeed. Validates that --json-indent fails fast
// at parse time, not silently at render time.
func TestStartupListJSONIndent_FlagParsing(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"default", []string{}, false},
		{"zero", []string{"--json-indent=0"}, false},
		{"max", []string{"--json-indent=8"}, false},
		{"negative", []string{"--json-indent=-1"}, true},
		{"too_large", []string{"--json-indent=9"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseStartupListFlags(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("args=%v: wantErr=%v got %v",
					tc.args, tc.wantErr, err)
			}
		})
	}
}
