package export

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/xuri/excelize/v2"
)

const sheetName = "Export"

func CSV(rows [][]string) ([]byte, error) {
	var out bytes.Buffer
	writer := csv.NewWriter(&out)
	if err := writer.WriteAll(rows); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func XLSX(rows [][]string) ([]byte, error) {
	book := excelize.NewFile()
	defer func() { _ = book.Close() }()
	if err := book.SetSheetName("Sheet1", sheetName); err != nil {
		return nil, err
	}
	stream, err := book.NewStreamWriter(sheetName)
	if err != nil {
		return nil, err
	}
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			return nil, err
		}
		values := make([]any, len(row))
		for j, value := range row {
			values[j] = value
		}
		if err := stream.SetRow(cell, values); err != nil {
			return nil, fmt.Errorf("write row %d: %w", i+1, err)
		}
	}
	if err := stream.Flush(); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := book.Write(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
