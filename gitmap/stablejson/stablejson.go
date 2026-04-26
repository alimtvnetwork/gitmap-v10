// Package stablejson encodes JSON arrays of objects with a field
// order that is GUARANTEED stable across Go versions, encoding/json
// rewrites, and any future encoding/json/v2 transition.
//
// # Why this package exists
//
// gitmap publishes several `--format=json` outputs that downstream
// scripts (jq pipelines, CI dashboards, third-party importers) parse
// positionally or with key-order assumptions. The standard
// `encoding/json` package emits struct fields in DECLARATION order
// today — this is documented but informally so. Three forces could
// break that contract:
//
//   1. Go 2 / encoding/json/v2 has been actively discussed; an early
//      proposal floated alphabetical key ordering. Even if rejected,
//      relying on the v1 quirk leaves us exposed.
//   2. Reflection-based field walks change subtly when fields are
//      added, removed, embedded, or marked omitempty.
//   3. Code-mod tools (gofmt, IDE refactors, generated code) routinely
//      reorder struct fields without warning.
//
// stablejson sidesteps all three by NEVER reflecting on a struct.
// The caller hands in an ordered slice of (key, value) pairs and the
// encoder writes them verbatim, in the given order. The only
// reflection is `json.Marshal` on each individual VALUE, which is a
// well-defined per-leaf operation independent of object shape.
//
// # Output contract
//
// WriteArray emits exactly the bytes that `json.Encoder` with
// `SetIndent("", "  ")` would produce for an equivalent slice of
// structs — including:
//
//   - 2-space indentation per nested level
//   - empty array as the literal `[]` (NOT `null`)
//   - trailing `\n` (matches Encoder.Encode behavior)
//   - `, ` between sibling values is replaced by `,\n` + indent so
//     the pretty-printed shape matches Encoder output byte-for-byte
//
// Byte-compat with Encoder is verified in stablejson_test.go and is
// the reason existing golden fixtures continue to pass after a
// caller migrates from json.Encoder to stablejson.WriteArray.
//
// # Non-goals
//
// stablejson is intentionally minimal: it handles arrays-of-objects,
// the only shape gitmap needs for stable list outputs. Nested objects
// inside a value are still serialized through `encoding/json`, which
// is fine because gitmap's stable surfaces are flat (string/number
// leaves only). If a future caller needs nested-object stability, it
// should pass a pre-rendered `json.RawMessage` as the value.
package stablejson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Field is one key/value pair in a stable object. The Key is emitted
// verbatim (caller is responsible for choosing on-the-wire names —
// typically lowerCamel to match the rest of gitmap's JSON outputs).
// The Value goes through json.Marshal, so any json.Marshaler works.
type Field struct {
	Key   string
	Value any
}

// WriteArray writes `items` as a JSON array of objects to w. Each
// inner []Field is one object; field order within an object follows
// the slice order. Empty `items` writes the literal `[]\n` so jq
// pipelines that do `length` never have to special-case `null`.
//
// Indentation matches `json.Encoder.SetIndent("", "  ")` so callers
// migrating from the standard encoder get byte-identical output and
// existing golden fixtures keep passing without regeneration.
func WriteArray(w io.Writer, items [][]Field) error {
	if len(items) == 0 {
		_, err := io.WriteString(w, "[]\n")

		return err
	}
	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i, obj := range items {
		if err := writeObject(&buf, obj); err != nil {

			return err
		}
		if i < len(items)-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("]\n")
	_, err := w.Write(buf.Bytes())

	return err
}

// writeObject writes a single `{ ... }` block at array-item
// indentation (2 spaces outer, 4 spaces inner) into buf. Keys appear
// in the exact order given. Each value is JSON-marshalled in
// isolation so a malformed value fails the WHOLE call rather than
// emitting half a corrupt object.
func writeObject(buf *bytes.Buffer, fields []Field) error {
	buf.WriteString("  {\n")
	for i, f := range fields {
		buf.WriteString("    ")
		keyBytes, err := json.Marshal(f.Key)
		if err != nil {

			return fmt.Errorf("stablejson: encode key %q: %w", f.Key, err)
		}
		buf.Write(keyBytes)
		buf.WriteString(": ")
		valBytes, err := json.Marshal(f.Value)
		if err != nil {

			return fmt.Errorf("stablejson: encode value for key %q: %w", f.Key, err)
		}
		buf.Write(valBytes)
		if i < len(fields)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("  }")

	return nil
}
