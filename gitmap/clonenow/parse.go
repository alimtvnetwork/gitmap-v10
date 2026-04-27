package clonenow

// Parser entry point: ParseFile dispatches on the explicit format
// override (cfg.Format, when non-empty) or otherwise on file
// extension, returns a fully-validated Plan.
//
// Three formats are supported because `gitmap scan` itself produces
// all three (.json, .csv, .txt) and clone-now's contract is "feed me
// any scan output -- I will round-trip it back into a working tree."
//
// JSON and CSV piggyback on formatter.ParseJSON / formatter.ParseCSV
// so the schema (RepoName, HTTPSUrl, SSHUrl, Branch, RelativePath, ...)
// is read by exactly one parser per format -- if scan ever evolves
// the schema, clone-now picks up the change automatically.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/formatter"
	"github.com/alimtvnetwork/gitmap-v7/gitmap/model"
)

// ParseFile is the package's only public parser entry point.
// `format` may be "" (auto-detect from extension) or one of
// constants.CloneNowFormat{JSON,CSV,Text}. `mode` and `onExists`
// must already be validated by the caller -- ParseFile records
// them on the Plan but does not interpret them (that's the
// executor / renderer's job).
func ParseFile(path, format, mode, onExists string) (Plan, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Plan{}, fmt.Errorf(constants.ErrCloneNowAbsPath, path, err)
	}
	resolved := format
	if len(resolved) == 0 {
		resolved = detectFormat(abs)
	}
	rows, err := parseByFormat(abs, resolved)
	if err != nil {
		return Plan{}, err
	}
	rows = dedupRows(rows)
	if len(rows) == 0 {
		return Plan{}, fmt.Errorf(constants.ErrCloneNowEmpty, abs)
	}

	return Plan{
		Source: abs, Format: resolved,
		Mode: mode, OnExists: onExists,
		Rows: rows,
	}, nil
}

// detectFormat picks json / csv / text from the lowercased file
// extension. Anything unfamiliar falls back to text -- the most
// permissive parser -- so a malformed input surfaces as "0 rows"
// rather than a hard "unknown format" error.
func detectFormat(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return constants.CloneNowFormatJSON
	case ".csv":
		return constants.CloneNowFormatCSV
	}

	return constants.CloneNowFormatText
}

// parseByFormat opens the file once and dispatches the io.Reader to
// the format-specific parser. Centralized so all three branches
// share identical open / close / error-wrapping behavior.
func parseByFormat(path, format string) ([]Row, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrCloneNowOpen, path, err)
	}
	defer f.Close()

	switch format {
	case constants.CloneNowFormatJSON:
		recs, err := formatter.ParseJSON(f)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrCloneNowJSONDecode, err)
		}

		return rowsFromRecords(recs), nil
	case constants.CloneNowFormatCSV:
		recs, err := formatter.ParseCSV(f)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrCloneNowCSVRead, err)
		}

		return rowsFromRecords(recs), nil
	}

	return parseTextRows(f)
}

// rowsFromRecords lifts a slice of scan records into clone-now Rows,
// dropping any record that has neither URL set (truly nothing to
// clone) and back-filling RelativePath from the URL basename when
// the record didn't carry one (older scan exports may omit it).
func rowsFromRecords(recs []model.ScanRecord) []Row {
	out := make([]Row, 0, len(recs))
	for _, rec := range recs {
		if len(rec.HTTPSUrl) == 0 && len(rec.SSHUrl) == 0 {
			continue
		}
		dest := rec.RelativePath
		if len(dest) == 0 {
			dest = deriveDestFromRecord(rec)
		}
		out = append(out, Row{
			RepoName:     rec.RepoName,
			HTTPSUrl:     strings.TrimSpace(rec.HTTPSUrl),
			SSHUrl:       strings.TrimSpace(rec.SSHUrl),
			Branch:       strings.TrimSpace(rec.Branch),
			RelativePath: strings.TrimSpace(dest),
		})
	}

	return out
}

// deriveDestFromRecord picks the best non-empty URL on the record
// and returns its basename (sans .git). Used as the fallback when a
// scan record came in without a recorded RelativePath.
func deriveDestFromRecord(rec model.ScanRecord) string {
	url := rec.HTTPSUrl
	if len(url) == 0 {
		url = rec.SSHUrl
	}

	return DeriveDest(url)
}

// dedupRows collapses rows that share the same destination path.
// Later rows win so users can override an earlier row by re-listing
// the same dest -- mirrors clonefrom's "later wins" semantics so the
// two commands feel uniform when a user switches between them.
func dedupRows(rows []Row) []Row {
	seen := make(map[string]int, len(rows))
	out := make([]Row, 0, len(rows))
	for _, r := range rows {
		key := r.RelativePath
		if idx, ok := seen[key]; ok {
			out[idx] = r

			continue
		}
		seen[key] = len(out)
		out = append(out, r)
	}

	return out
}
