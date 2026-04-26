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

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
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

// startupListJSONEntry mirrors startup.Entry with explicit JSON tags
// so the on-the-wire field names are stable even if the internal
// struct gets renamed. lower_snake would be more idiomatic for JSON
// but we use lowerCamel to match the rest of gitmap's JSON outputs.
//
// CONTRACT: the field set, JSON tag names, and field DECLARATION
// order are pinned by gitmap/cmd/startuplistjson_contract_test.go.
// Reordering, renaming, or adding/removing fields will fail those
// tests — intentional changes require regenerating the golden
// fixtures and bumping the consumer-facing changelog.
type startupListJSONEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Exec string `json:"exec"`
}

// encodeStartupListJSON writes a JSON array to w. Always emits a
// non-nil slice so empty results encode as `[]` (NOT `null`).
func encodeStartupListJSON(w io.Writer, entries []startup.Entry) error {
	out := make([]startupListJSONEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, startupListJSONEntry{Name: e.Name, Path: e.Path, Exec: e.Exec})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", constants.JSONIndent)

	return enc.Encode(out)
}

// encodeStartupListCSV writes a header row followed by one row per
// entry. encoding/csv handles quoting of values containing commas,
// quotes, or newlines — important because Exec lines can include
// shell quoting and embedded spaces.
func encodeStartupListCSV(w io.Writer, entries []startup.Entry) error {
	cw := csv.NewWriter(w)
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
