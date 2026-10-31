package ingestion

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func writeBook(t *testing.T, book *excelize.File) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := book.Write(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

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
	return writeBook(t, book)
}

func TestXLSXToCSV_ConvertsFirstSheetToRows_WhenFileIsValid(t *testing.T) {
	data := buildXLSX(t, [][]any{{"phone", "name"}, {"+79001112233", "Anna"}})

	if !IsXLSX(data) {
		t.Fatal("expected zip signature")
	}
	csvData, err := XLSXToCSV(data, 100, "occurred_at")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(csvData), "phone,name\n+79001112233,Anna\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXLSXToCSV_WritesRFC3339_WhenCellIsDate(t *testing.T) {
	moment := time.Date(2026, 10, 1, 10, 30, 0, 0, time.UTC)
	data := buildXLSX(t, [][]any{{"occurred_at"}, {moment}})

	csvData, err := XLSXToCSV(data, 100, "occurred_at")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(csvData), "occurred_at\n2026-10-01T10:30:00Z\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXLSXToCSV_KeepsNumbers_WhenColumnIsNotDate(t *testing.T) {
	data := buildXLSX(t, [][]any{{"amount", "occurred_at"}, {2, "2026-10-01T10:00:00Z"}})

	csvData, err := XLSXToCSV(data, 100, "occurred_at")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(csvData), "amount,occurred_at\n2,2026-10-01T10:00:00Z\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXLSXToCSV_MatchesDateColumn_WhenHeaderCaseDiffers(t *testing.T) {
	moment := time.Date(2026, 10, 1, 10, 30, 0, 0, time.UTC)
	data := buildXLSX(t, [][]any{{"Occurred_At"}, {moment}})

	csvData, err := XLSXToCSV(data, 100, "occurred_at")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(csvData), "Occurred_At\n2026-10-01T10:30:00Z\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXLSXToCSV_RejectsFile_WhenRowsExceedLimit(t *testing.T) {
	data := buildXLSX(t, [][]any{{"a"}, {"1"}, {"2"}, {"3"}})

	_, err := XLSXToCSV(data, 2, "occurred_at")
	if !errors.Is(err, ErrCSVFileTooLarge) {
		t.Fatalf("expected file too large, got %v", err)
	}
}

func TestXLSXToCSV_AcceptsFile_WhenRowsEqualLimit(t *testing.T) {
	data := buildXLSX(t, [][]any{{"a"}, {"1"}, {"2"}})

	if _, err := XLSXToCSV(data, 2, "occurred_at"); err != nil {
		t.Fatal(err)
	}
}

func TestXLSXToCSV_RejectsFile_WhenSheetIsEmpty(t *testing.T) {
	data := writeBook(t, excelize.NewFile())

	_, err := XLSXToCSV(data, 100, "occurred_at")
	if !errors.Is(err, ErrCSVEmpty) {
		t.Fatalf("expected empty file, got %v", err)
	}
}

func TestXLSXToCSV_RejectsBadFormat_WhenFileIsNotXLSX(t *testing.T) {
	_, err := XLSXToCSV([]byte("PK\x03\x04garbage"), 100, "occurred_at")
	if !errors.Is(err, ErrCSVBadFormat) {
		t.Fatalf("expected bad format, got %v", err)
	}
}

func TestIsXLSX_ReturnsFalse_WhenFileIsCSV(t *testing.T) {
	if IsXLSX([]byte("phone,name\n")) {
		t.Fatal("csv must not be detected as xlsx")
	}
}
