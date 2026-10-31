package ingestion

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/xuri/excelize/v2"
)

var zipSignature = []byte("PK\x03\x04")

func IsXLSX(data []byte) bool {
	return bytes.HasPrefix(data, zipSignature)
}

func XLSXToCSV(data []byte) ([]byte, error) {
	book, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCSVBadFormat, err)
	}
	defer func() { _ = book.Close() }()

	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrCSVEmpty
	}
	rows, err := book.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCSVBadFormat, err)
	}
	if len(rows) == 0 {
		return nil, ErrCSVEmpty
	}

	var out bytes.Buffer
	writer := csv.NewWriter(&out)
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, ErrCSVBadFormat
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, ErrCSVBadFormat
	}
	return out.Bytes(), nil
}
