package ingestion

import (
	"bytes"
	"errors"
	"testing"

	"github.com/xuri/excelize/v2"
)

func buildXLSX(t *testing.T, rows [][]any) []byte {
	t.Helper()
	book := excelize.NewFile()
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			t.Fatal(err)
		}
		if err := book.SetSheetRow("Sheet1", cell, &row); err != nil {
			t.Fatal(err)
		}
	}
	var buf bytes.Buffer
	if err := book.Write(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestXLSXToCSV_ConvertsFirstSheetToRows_WhenFileIsValid(t *testing.T) {
	data := buildXLSX(t, [][]any{{"phone", "name"}, {"+79001112233", "Anna"}})

	if !IsXLSX(data) {
		t.Fatal("expected zip signature")
	}
	csvData, err := XLSXToCSV(data)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(csvData), "phone,name\n+79001112233,Anna\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXLSXToCSV_RejectsBadFormat_WhenFileIsNotXLSX(t *testing.T) {
	_, err := XLSXToCSV([]byte("PK\x03\x04garbage"))
	if !errors.Is(err, ErrCSVBadFormat) {
		t.Fatalf("expected bad format, got %v", err)
	}
}

func TestIsXLSX_ReturnsFalse_WhenFileIsCSV(t *testing.T) {
	if IsXLSX([]byte("phone,name\n")) {
		t.Fatal("csv must not be detected as xlsx")
	}
}
