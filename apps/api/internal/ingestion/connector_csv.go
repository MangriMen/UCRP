package ingestion

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Row map[string]string

type CSVConnector struct {
	MaxRows int
}

func NewCSVConnector(maxRows int) *CSVConnector {
	if maxRows <= 0 {
		maxRows = 100000
	}
	return &CSVConnector{MaxRows: maxRows}
}

var (
	ErrCSVFileTooLarge   = errors.New("file_too_large")
	ErrCSVMissingColumns = errors.New("missing_columns")
	ErrCSVBadFormat      = errors.New("bad_format")
	ErrCSVEmpty          = errors.New("empty_file")
)

func (c *CSVConnector) Read(r io.Reader, mapping FieldMapping) ([]Row, error) {
	cr := csvReader(r)
	header, err := cr.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, ErrCSVEmpty
		}
		return nil, ErrCSVBadFormat
	}
	cols := headerCols(header)
	if err := validateHeader(cols, mapping); err != nil {
		return nil, err
	}

	out := make([]Row, 0, 64)
	for {
		record, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, ErrCSVBadFormat
		}
		if len(out)+1 > c.MaxRows {
			return nil, ErrCSVFileTooLarge
		}
		row := make(Row, len(cols))
		for i, col := range cols {
			if i < len(record) {
				row[col] = strings.TrimSpace(record[i])
			}
		}
		out = append(out, row)
	}
	return out, nil
}

func ValidateHeader(r io.Reader, mapping FieldMapping) error {
	cr := csvReader(r)
	header, err := cr.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return ErrCSVEmpty
		}
		return ErrCSVBadFormat
	}
	return validateHeader(headerCols(header), mapping)
}

func csvReader(r io.Reader) *csv.Reader {
	cr := csv.NewReader(r)
	cr.Comma = ','
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true
	return cr
}

func headerCols(header []string) []string {
	cols := make([]string, 0, len(header))
	for _, h := range header {
		cols = append(cols, normalizeHeader(h))
	}
	return cols
}

func normalizeHeader(h string) string {
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "")
	return strings.ToLower(h)
}

func validateHeader(cols []string, mapping FieldMapping) error {
	have := make(map[string]bool, len(cols))
	for _, c := range cols {
		have[c] = true
	}
	for _, need := range mapping.requiredColumns() {
		if need == "" {
			continue
		}
		if !have[strings.ToLower(strings.TrimSpace(need))] {
			return fmt.Errorf("%w: %s", ErrCSVMissingColumns, need)
		}
	}
	return nil
}
