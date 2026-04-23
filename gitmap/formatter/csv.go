package formatter

import (
	"encoding/csv"
	"io"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
)

// WriteCSV writes records to the given writer in CSV format.
//
// Records are validated first; per-issue warnings are emitted to the
// configured sink (default os.Stderr) but the write always proceeds.
// See validate.go for the warn-and-write policy.
func WriteCSV(w io.Writer, records []model.ScanRecord) error {
	issueCount := emitValidationWarnings(records)

	cw := csv.NewWriter(w)
	err := cw.Write(constants.ScanCSVHeaders)
	if err != nil {
		return err
	}
	err = writeCSVRows(cw, records)
	if err != nil {
		return err
	}
	emitWriteSummary("csv", len(records), issueCount)

	return nil
}

// writeCSVRows writes each record as a CSV row and flushes.
func writeCSVRows(cw *csv.Writer, records []model.ScanRecord) error {
	for _, r := range records {
		err := writeCSVRow(cw, r)
		if err != nil {
			return err
		}
	}
	cw.Flush()

	return cw.Error()
}

// writeCSVRow converts a single record to a CSV row.
func writeCSVRow(cw *csv.Writer, r model.ScanRecord) error {
	row := []string{
		r.RepoName, r.HTTPSUrl, r.SSHUrl, r.Branch, r.BranchSource,
		r.RelativePath, r.AbsolutePath, r.CloneInstruction, r.Notes,
	}

	return cw.Write(row)
}

// ParseCSV reads records from a CSV reader.
func ParseCSV(reader io.Reader) ([]model.ScanRecord, error) {
	cr := csv.NewReader(reader)
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	return parseCSVRows(rows), nil
}

// parseCSVRows converts raw CSV rows (skipping header) into records.
// Supports both the legacy 8-column layout and the current 9-column layout
// (which adds branchSource after branch).
func parseCSVRows(rows [][]string) []model.ScanRecord {
	records := make([]model.ScanRecord, 0, len(rows))
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) >= 8 {
			records = append(records, rowToRecord(row))
		}
	}

	return records
}

// rowToRecord maps a CSV row to a ScanRecord. Accepts both legacy (8 cols)
// and current (9 cols including branchSource) layouts.
func rowToRecord(row []string) model.ScanRecord {
	if len(row) >= 9 {
		notes := ""
		if len(row) > 8 {
			notes = row[8]
		}

		return model.ScanRecord{
			RepoName: row[0], HTTPSUrl: row[1], SSHUrl: row[2],
			Branch: row[3], BranchSource: row[4],
			RelativePath: row[5], AbsolutePath: row[6],
			CloneInstruction: row[7], Notes: notes,
		}
	}

	notes := ""
	if len(row) > 7 {
		notes = row[7]
	}

	return model.ScanRecord{
		RepoName: row[0], HTTPSUrl: row[1], SSHUrl: row[2],
		Branch: row[3], RelativePath: row[4], AbsolutePath: row[5],
		CloneInstruction: row[6], Notes: notes,
	}
}
