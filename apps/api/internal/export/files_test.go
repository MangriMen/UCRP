package export

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"
)

var sampleRows = [][]string{
	{"guest_id", "contact"},
	{"g-1", "+79001112233"},
	{"g-2", "a,b@example.com"},
}

func TestCSV_WritesQuotedRows_WhenValueContainsComma(t *testing.T) {
	data, err := CSV(sampleRows)
	if err != nil {
		t.Fatal(err)
	}
	want := "guest_id,contact\ng-1,+79001112233\ng-2,\"a,b@example.com\"\n"
	if string(data) != want {
		t.Fatalf("got %q, want %q", data, want)
	}
}

func TestXLSX_RoundTripsRowsAsText_WhenValuesLookNumeric(t *testing.T) {
	data, err := XLSX(sampleRows)
	if err != nil {
		t.Fatal(err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = book.Close() }()
	got, err := book.GetRows(sheetName)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, sampleRows) {
		t.Fatalf("got %v, want %v", got, sampleRows)
	}
}

func TestXLSX_WritesHeaderOnly_WhenThereAreNoMembers(t *testing.T) {
	data, err := XLSX(sampleRows[:1])
	if err != nil {
		t.Fatal(err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = book.Close() }()
	got, err := book.GetRows(sheetName)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d rows, want 1", len(got))
	}
}
