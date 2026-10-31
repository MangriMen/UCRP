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

func XLSXToCSV(data []byte, maxRows int) ([]byte, error) {
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
	sheet := sheets[0]
	rows, err := book.Rows(sheet)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCSVBadFormat, err)
	}
	defer func() { _ = rows.Close() }()

	dates := dateStyles{book: book, cache: map[int]bool{}}
	var out bytes.Buffer
	writer := csv.NewWriter(&out)
	written := 0
	for rowNumber := 1; rows.Next(); rowNumber++ {
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
		if rowNumber > 1 {
			for i, cell := range cells {
				cells[i] = dates.normalize(sheet, i+1, rowNumber, cell)
			}
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

type dateStyles struct {
	book  *excelize.File
	cache map[int]bool
}

func (d dateStyles) normalize(sheet string, column, row int, raw string) string {
	serial, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}
	cell, err := excelize.CoordinatesToCellName(column, row)
	if err != nil {
		return raw
	}
	styleID, err := d.book.GetCellStyle(sheet, cell)
	if err != nil || !d.isDate(styleID) {
		return raw
	}
	moment, err := excelize.ExcelDateToTime(serial, false)
	if err != nil {
		return raw
	}
	return moment.UTC().Format(time.RFC3339)
}

func (d dateStyles) isDate(styleID int) bool {
	if known, ok := d.cache[styleID]; ok {
		return known
	}
	style, err := d.book.GetStyle(styleID)
	result := err == nil && style != nil && isDateFormat(style)
	d.cache[styleID] = result
	return result
}

func isDateFormat(style *excelize.Style) bool {
	if style.CustomNumFmt != nil {
		return looksLikeDateFormat(*style.CustomNumFmt)
	}
	return (style.NumFmt >= 14 && style.NumFmt <= 22) || (style.NumFmt >= 45 && style.NumFmt <= 47)
}

func looksLikeDateFormat(format string) bool {
	var plain strings.Builder
	quoted, bracketed := false, false
	for _, r := range strings.ToLower(format) {
		switch {
		case r == '"':
			quoted = !quoted
		case r == '[':
			bracketed = true
		case r == ']':
			bracketed = false
		case !quoted && !bracketed:
			plain.WriteRune(r)
		}
	}
	return strings.ContainsAny(plain.String(), "ymdh")
}
