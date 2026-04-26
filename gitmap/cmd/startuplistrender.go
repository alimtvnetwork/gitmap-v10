package cmd

// Renderers for `gitmap startup-list --format=...`. Split out from
// startup.go so the per-format encoder logic doesn't push the parent
// file over the 200-line code-style budget. Three formats supported:
//
//   - table (default): human-readable, identical to the pre-flag
//     output so existing users see no change.
//   - json: array of {name, path, exec} objects. Empty list renders
//     as `[]` (NOT `null`) so jq-based pipelines never have to
//     special-case missing data.
//   - csv: RFC4180 via encoding/csv. Header row is always written so
//     downstream tools can self-discover columns. Empty list still
//     emits the header so spreadsheet imports get consistent shape.
//
// JSON / CSV encoders take an io.Writer (rather than hardcoding
// os.Stdout) so contract tests can capture the bytes into a buffer
// for byte-exact comparison against committed golden fixtures.
//
// JSON encoding goes through gitmap/stablejson rather than
// encoding/json directly: stablejson builds each object key-by-key
// in caller-declared order and CANNOT be reordered by a future Go
// release or encoding/json/v2. The output is byte-identical to the
// previous Encoder-based code, so the existing golden fixtures pass
// unchanged. See gitmap/stablejson/stablejson.go for full rationale.

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/stablejson"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/startup"
)

// renderStartupList dispatches to the per-format encoder.
func renderStartupList(format, dir string, entries []startup.Entry) error {
	switch format {
	case constants.OutputJSON:

		return encodeStartupListJSON(os.Stdout, entries)
	case constants.OutputCSV:

		return encodeStartupListCSV(os.Stdout, entries)
	default:
		renderStartupListTable(dir, entries)

		return nil
	}
}

// renderStartupListTable is the legacy human-readable rendering —
// extracted verbatim from the original runStartupList so behavior
// for users who don't pass --format is byte-identical.
func renderStartupListTable(dir string, entries []startup.Entry) {
	fmt.Printf(constants.MsgStartupListHeader, dir)
	if len(entries) == 0 {
		fmt.Print(constants.MsgStartupListEmpty)

		return
	}
	for _, e := range entries {
		fmt.Printf(constants.MsgStartupListRow, e.Name, renderExec(e.Exec))
	}
	fmt.Printf(constants.MsgStartupListFooter, len(entries))
}

// startupListJSONFields names — single source of truth for both the
// on-the-wire field labels and their order. Centralized as constants
// so any consumer-facing rename or reorder is one diff to review.
//
// CONTRACT: the names AND order here are pinned by
// gitmap/cmd/startuplistjson_contract_test.go (golden bytes for
// empty + single, structural key-order check for multi-entry).
// Reordering, renaming, or adding fields will fail those tests —
// intentional changes require regenerating the golden fixtures and
// bumping the consumer-facing changelog.
const (
	startupListJSONKeyName = "name"
	startupListJSONKeyPath = "path"
	startupListJSONKeyExec = "exec"
)

// encodeStartupListJSON writes a JSON array to w using stablejson,
// which builds each object key-by-key in caller-declared order
// instead of relying on encoding/json's reflection-based struct
// field iteration. This guarantees field order CANNOT drift even if
// a future Go release (or encoding/json/v2) changes how struct
// fields are walked. See gitmap/stablejson/stablejson.go for the
// full rationale and byte-compat guarantee.
//
// Empty input still encodes as `[]\n` (NOT `null`) so jq pipelines
// that do `length` work without conditionals.
func encodeStartupListJSON(w io.Writer, entries []startup.Entry) error {
	items := make([][]stablejson.Field, 0, len(entries))
	for _, e := range entries {
		items = append(items, []stablejson.Field{
			{Key: startupListJSONKeyName, Value: e.Name},
			{Key: startupListJSONKeyPath, Value: e.Path},
			{Key: startupListJSONKeyExec, Value: e.Exec},
		})
	}

	return stablejson.WriteArray(w, items)
}

// encodeStartupListCSV writes a header row followed by one row per
// entry. encoding/csv handles quoting of values containing commas,
// quotes, or newlines — important because Exec lines can include
// shell quoting and embedded spaces.
func encodeStartupListCSV(w io.Writer, entries []startup.Entry) error {
	cw := csv.NewWriter(w)
	// CRLF for cross-platform byte-identical output (RFC 4180).
	// Pinned by gitmap/cmd/csvcrlf_contract_test.go.
	cw.UseCRLF = true
	if err := cw.Write([]string{"name", "path", "exec"}); err != nil {

		return err
	}
	for _, e := range entries {
		if err := cw.Write([]string{e.Name, e.Path, e.Exec}); err != nil {

			return err
		}
	}
	cw.Flush()

	return cw.Error()
}

// renderExec keeps long Exec lines from making the list table noisy
// — falls back to a placeholder when the .desktop file omits Exec.
func renderExec(exec string) string {
	if len(exec) == 0 {

		return "(no Exec line)"
	}

	return exec
}
