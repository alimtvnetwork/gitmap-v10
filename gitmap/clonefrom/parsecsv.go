package clonefrom

// CSV parser half of ParseFile. Split from parse.go so each file
// stays under the 200-line per-file budget. Public surface is
// unchanged: callers go through ParseFile in parse.go which
// dispatches to parseCSV here based on file extension.

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/alimtvnetwork/gitmap-v8/gitmap/constants"
)

// parseCSV expects a header row of `url,dest,branch,depth` (case-
// insensitive, only `url` required). Missing optional columns
// default to empty/zero. Extra columns past `depth` are ignored.
func parseCSV(r io.Reader) ([]Row, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1 // tolerate ragged rows
	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf(constants.ErrCloneFromCSVHeader, err)
	}
	idx := indexCSVHeader(header)
	if idx.url < 0 {
		return nil, fmt.Errorf(constants.ErrCloneFromCSVNoURL)
	}

	return readCSVRows(cr, idx)
}

// readCSVRows is the inner loop split out so parseCSV stays under
// the function-length budget.
func readCSVRows(cr *csv.Reader, idx csvIndex) ([]Row, error) {
	var out []Row
	rowNum := 1 // header was row 1; first data row is 2
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		rowNum++
		if err != nil {
			return nil, fmt.Errorf(constants.ErrCloneFromCSVRow, rowNum, err)
		}
		row, err := csvRow(rec, idx)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrCloneFromCSVRow, rowNum, err)
		}
		out = append(out, row)
	}

	return out, nil
}

// csvIndex maps logical column names to record indices. Negative
// value means "column absent" — only `url` is required.
type csvIndex struct{ url, dest, branch, depth, checkout int }

// indexCSVHeader walks the header row once and records each
// column's position. Case-insensitive so spreadsheet exports with
// "URL"/"Url" headers work without preprocessing.
func indexCSVHeader(header []string) csvIndex {
	idx := csvIndex{url: -1, dest: -1, branch: -1, depth: -1, checkout: -1}
	for i, name := range header {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "url":
			idx.url = i
		case "dest":
			idx.dest = i
		case "branch":
			idx.branch = i
		case "depth":
			idx.depth = i
		case "checkout":
			idx.checkout = i
		}
	}

	return idx
}

// csvRow extracts one Row from a parsed CSV record using the
// pre-computed column index. Returns a wrapped error on bad depth.
func csvRow(rec []string, idx csvIndex) (Row, error) {
	row := Row{
		URL:      strings.TrimSpace(get(rec, idx.url)),
		Dest:     strings.TrimSpace(get(rec, idx.dest)),
		Branch:   strings.TrimSpace(get(rec, idx.branch)),
		Checkout: strings.ToLower(strings.TrimSpace(get(rec, idx.checkout))),
	}
	if depthStr := strings.TrimSpace(get(rec, idx.depth)); len(depthStr) > 0 {
		d, err := strconv.Atoi(depthStr)
		if err != nil {
			return row, fmt.Errorf(constants.ErrCloneFromBadDepth, depthStr)
		}
		row.Depth = d
	}

	return row, validateRow(row)
}

// get is a bounds-safe slice accessor. Returns empty string when
// the index is negative (column absent in header) or past the
// record's end (ragged row).
func get(rec []string, i int) string {
	if i < 0 || i >= len(rec) {
		return ""
	}

	return rec[i]
}
