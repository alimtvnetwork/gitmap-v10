package clonenow

// Schema validation for clone-now JSON and CSV inputs.
//
// The downstream formatter parsers (formatter.ParseJSON /
// formatter.ParseCSV) are intentionally tolerant: unknown JSON keys
// are silently ignored, and CSVs with the wrong header row will
// happily mis-map columns into ScanRecord fields and yield garbage
// rows. That's the right call for the scan -> read pipeline (forward
// compatibility), but it's the wrong call for clone-now where the
// user is being told "feed me a scan artifact" and a typo in a key
// name should produce a clear error, not silent data loss.
//
// This file adds two pre-flight checks that run BEFORE the tolerant
// parsers and reject inputs whose shape doesn't match the documented
// schema:
//
//   - validateJSONSchema -- input must be a JSON array of objects;
//     every object key must be a known ScanRecord JSON tag; each
//     row must carry at least one of httpsUrl / sshUrl.
//   - validateCSVSchema  -- input must have a header row; every
//     header name must be a known ScanRecord CSV tag; the header
//     must include at least one of httpsUrl / sshUrl.
//
// Both checks return errors built from the constants in
// constants_clonenow.go so messages stay greppable and stable.

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/alimtvnetwork/gitmap-v8/gitmap/constants"
)

// knownScanFields is the authoritative set of JSON / CSV field names
// accepted by clone-now. Mirrors model.ScanRecord's struct tags --
// keep in sync if a new field is ever added there.
//
// We keep this as a package-level var (not a const map) so tests can
// reference it for golden-style assertions if they ever need to.
var knownScanFields = map[string]bool{
	"id":               true,
	"slug":             true,
	"repoId":           true,
	"repoName":         true,
	"httpsUrl":         true,
	"sshUrl":           true,
	"discoveredUrl":    true,
	"branch":           true,
	"branchSource":     true,
	"relativePath":     true,
	"absolutePath":     true,
	"cloneInstruction": true,
	"notes":            true,
	"depth":            true,
	"transport":        true,
}

// validateJSONSchema ensures the JSON input is an array of objects
// whose keys are all known ScanRecord field names and where every
// object carries at least one URL. Two-stage decode (array first,
// then per-element object) so a non-object element is reported with
// its 1-based row number instead of a raw decoder offset that the
// user has to hand-translate to a row.
func validateJSONSchema(data []byte) error {
	var elems []json.RawMessage
	if err := json.Unmarshal(data, &elems); err != nil {
		return fmt.Errorf(constants.ErrCloneNowJSONShape, err)
	}
	for i, raw := range elems {
		if err := validateJSONElement(i, raw); err != nil {
			return err
		}
	}

	return nil
}

// validateJSONElement decodes one array element as an object and
// runs the per-row checks. A non-object element (string, number,
// null, nested array) is reported with its 1-based index and the
// observed JSON kind so the user can jump straight to the line.
func validateJSONElement(i int, raw json.RawMessage) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fmt.Errorf(constants.ErrCloneNowJSONRowNotObject,
			i+1, jsonKind(raw))
	}

	return validateJSONRow(i, obj)
}

// validateJSONRow checks one decoded object: every key must be in
// knownScanFields and the row must carry at least one URL. The row
// index is 1-based in the error so it matches what a human reading
// the JSON file would count.
func validateJSONRow(i int, obj map[string]json.RawMessage) error {
	for k := range obj {
		if !knownScanFields[k] {
			return fmt.Errorf(constants.ErrCloneNowUnknownJSONField, i+1, k, knownFieldList())
		}
	}
	if !hasJSONURL(obj) {
		return fmt.Errorf(constants.ErrCloneNowMissingURL, i+1)
	}

	return nil
}

// jsonKind returns a short human label for a RawMessage's top-level
// JSON kind, used in error messages so users see "got string" rather
// than a parser offset. Falls back to "invalid" for malformed bytes.
func jsonKind(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if len(s) == 0 {
		return "empty"
	}
	switch s[0] {
	case '"':
		return "string"
	case '[':
		return "array"
	case '{':
		return "object"
	case 't', 'f':
		return "boolean"
	case 'n':
		return "null"
	}
	if (s[0] >= '0' && s[0] <= '9') || s[0] == '-' {
		return "number"
	}

	return "invalid"
}

// hasJSONURL reports whether the row carries a non-empty httpsUrl
// or sshUrl. Empty strings ("") count as missing -- the executor
// would skip the row anyway, and a clear pre-flight error is more
// useful than a silent drop.
func hasJSONURL(obj map[string]json.RawMessage) bool {
	return jsonStringNonEmpty(obj["httpsUrl"]) || jsonStringNonEmpty(obj["sshUrl"])
}

// jsonStringNonEmpty decodes a RawMessage as a string and reports
// whether the result is non-empty. Non-string values are treated as
// empty so a typo like `"httpsUrl": null` doesn't pass the URL gate.
func jsonStringNonEmpty(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return false
	}

	return len(strings.TrimSpace(s)) > 0
}

// validateCSVSchema ensures the CSV input has a recognizable header
// row whose every column is a known field and which includes at
// least one of httpsUrl / sshUrl, AND that every data row carries a
// non-empty URL. Per-row failures are reported with a 1-based DATA
// row number (header is row 0; data row 1 = first row after header,
// matching what a spreadsheet user sees as "row 2").
func validateCSVSchema(r io.Reader) error {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	header, err := cr.Read()
	if err == io.EOF {
		return fmt.Errorf(constants.ErrCloneNowEmptyCSV)
	}
	if err != nil {
		return fmt.Errorf(constants.ErrCloneNowCSVRead, err)
	}
	if err := validateCSVHeader(header); err != nil {
		return err
	}

	return validateCSVBody(cr, header)
}

// validateCSVBody streams the remaining CSV records (header already
// consumed) and asserts each data row has a non-empty URL in either
// httpsUrl or sshUrl. Field-count drift is intentionally NOT
// enforced here — the header validator already tolerates trailing
// empty columns, and uneven trailing emptiness is harmless. Stops
// at the first row-level failure so users fix one issue at a time.
func validateCSVBody(cr *csv.Reader, header []string) error {
	urlIdxs := urlColumnIndexes(header)
	for dataRow := 1; ; dataRow++ {
		rec, err := cr.Read()
		if err == io.EOF {

			return nil
		}
		if err != nil {
			return fmt.Errorf(constants.ErrCloneNowCSVRowRead, dataRow, err)
		}
		if !rowHasURL(rec, urlIdxs) {
			return fmt.Errorf(constants.ErrCloneNowCSVRowMissingURL, dataRow)
		}
	}
}

// urlColumnIndexes returns the positions of the httpsUrl / sshUrl
// columns in the (already validated) header. Empty slice cannot
// occur because validateCSVHeader rejects headers without one.
func urlColumnIndexes(header []string) []int {
	out := make([]int, 0, 2)
	for i, col := range header {
		name := normalizeHeaderName(col)
		if name == "httpsUrl" || name == "sshUrl" {
			out = append(out, i)
		}
	}

	return out
}

// rowHasURL reports whether at least one of the URL columns in this
// data row holds a non-whitespace value.
func rowHasURL(rec []string, urlIdxs []int) bool {
	for _, i := range urlIdxs {
		if i < len(rec) && len(strings.TrimSpace(rec[i])) > 0 {
			return true
		}
	}

	return false
}

// validateCSVHeader checks every header column against the known
// field set and confirms at least one URL column is present.
// Headers are normalized via normalizeHeaderName before lookup so
// real-world quirks (UTF-8 BOM on the first column, surrounding
// double-quotes preserved by some exporters, leading/trailing
// whitespace) don't trigger false "unknown column" rejections.
// Empty/whitespace-only header cells are tolerated and ignored —
// they're a common artifact of trailing commas in spreadsheet
// exports and the row payload would be empty for that column anyway.
func validateCSVHeader(header []string) error {
	hasURL := false
	for _, col := range header {
		name := normalizeHeaderName(col)
		if len(name) == 0 {

			continue
		}
		if !knownScanFields[name] {
			return fmt.Errorf(constants.ErrCloneNowUnknownCSVField, name, knownFieldList())
		}
		if name == "httpsUrl" || name == "sshUrl" {
			hasURL = true
		}
	}
	if !hasURL {
		return fmt.Errorf(constants.ErrCloneNowCSVMissingURLCol)
	}

	return nil
}

// normalizeHeaderName returns a header cell ready for the known-
// field lookup. Strips a UTF-8 BOM (only meaningful on the first
// cell, but cheap to apply uniformly), one optional pair of
// surrounding double-quotes (some exporters quote header names
// even when not required), and surrounding whitespace.
func normalizeHeaderName(col string) string {
	const utf8BOM = "\ufeff"
	s := strings.TrimPrefix(col, utf8BOM)
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	return strings.TrimSpace(s)
}

// knownFieldList returns the known field names sorted alphabetically
// and comma-joined for inclusion in error messages. Stable ordering
// makes test assertions deterministic.
func knownFieldList() string {
	names := make([]string, 0, len(knownScanFields))
	for k := range knownScanFields {
		names = append(names, k)
	}
	sort.Strings(names)

	return strings.Join(names, ", ")
}
