package ingestion

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	xlsxUnzipLimit    = 200 << 20
	xlsxUnzipXMLLimit = 64 << 20
)

var zipSignature = []byte("PK\x03\x04")

func IsXLSX(data []byte) bool {
	return bytes.HasPrefix(data, zipSignature)
}

func XLSXToCSV(data []byte, maxRows int, dateColumn string) ([]byte, error) {
	book, err := excelize.OpenReader(bytes.NewReader(data), excelize.Options{
		UnzipSizeLimit:    xlsxUnzipLimit,
		UnzipXMLSizeLimit: xlsxUnzipXMLLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCSVBadFormat, err)
	}
	defer func() { _ = book.Close() }()

	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrCSVEmpty
	}
	rows, err := book.Rows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCSVBadFormat, err)
	}
	defer func() { _ = rows.Close() }()

	date1904 := book.WorkBook != nil && book.WorkBook.WorkbookPr != nil && book.WorkBook.WorkbookPr.Date1904
	dateIndex := -1
	var out bytes.Buffer
	writer := csv.NewWriter(&out)
	written := 0
	for rows.Next() {
		cells, err := rows.Columns(excelize.Options{RawCellValue: true})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCSVBadFormat, err)
		}
		if isBlankRow(cells) {
			continue
		}
		if written > maxRows {
			return nil, ErrCSVFileTooLarge
		}
		if written == 0 {
			dateIndex = columnIndex(cells, dateColumn)
		} else if dateIndex >= 0 && dateIndex < len(cells) {
			cells[dateIndex] = serialToTimestamp(cells[dateIndex], date1904)
		}
		if err := writer.Write(cells); err != nil {
			return nil, ErrCSVBadFormat
		}
		written++
	}
	if written == 0 {
		return nil, ErrCSVEmpty
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, ErrCSVBadFormat
	}
	return out.Bytes(), nil
}

func isBlankRow(cells []string) bool {
	for _, cell := range cells {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func columnIndex(header []string, name string) int {
	want := normalizeHeader(name)
	for i, title := range header {
		if normalizeHeader(title) == want {
			return i
		}
	}
	return -1
}

func serialToTimestamp(raw string, date1904 bool) string {
	serial, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return raw
	}
	moment, err := excelize.ExcelDateToTime(serial, date1904)
	if err != nil {
		return raw
	}
	return moment.UTC().Format(time.RFC3339)
}
