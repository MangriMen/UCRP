package e2e

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

const csvHeader = "external_id,occurred_at,event_type,amount,phone,email,tenant_code\n"

func csvValid(tenantCode, phoneA, phoneB string) []byte {
	return []byte(csvHeader +
		"e2e-1,2026-10-01T10:00:00Z,purchase,1000," + phoneA + ",," + tenantCode + "\n" +
		"e2e-2,2026-10-02T11:00:00Z,visit,0," + phoneB + ",," + tenantCode + "\n")
}

func csvWithRejectedRows(tenantCode, phone string) []byte {
	return []byte(csvHeader +
		"e2e-ok,2026-10-01T10:00:00Z,purchase,1000," + phone + ",," + tenantCode + "\n" +
		"e2e-bad-date,not-a-date,purchase,500," + phone + ",," + tenantCode + "\n" +
		",2026-10-02T10:00:00Z,purchase,100,,," + tenantCode + "\n" +
		"e2e-bad-tenant,2026-10-03T10:00:00Z,purchase,100," + phone + ",,no-such-tenant\n")
}

func csvMissingColumns() []byte {
	return []byte("external_id,amount\n" +
		"e2e-1,1000\n")
}

func csvEmpty() []byte {
	return []byte("")
}

func csvDuplicateExternalID(tenantCode, phone string) []byte {
	return []byte(csvHeader +
		"e2e-dup,2026-10-01T10:00:00Z,purchase,1000," + phone + ",," + tenantCode + "\n" +
		"e2e-dup,2026-10-02T10:00:00Z,purchase,2000," + phone + ",," + tenantCode + "\n")
}

func xlsxValid(t *testing.T, tenantCode, phoneA, phoneB string) []byte {
	t.Helper()
	records := [][]any{
		{"external_id", "occurred_at", "event_type", "amount", "phone", "email", "tenant_code"},
		{"e2e-x1", "2026-10-01T10:00:00Z", "purchase", 1000, phoneA, "", tenantCode},
		{"e2e-x2", "2026-10-02T11:00:00Z", "visit", 0, phoneB, "", tenantCode},
	}
	book := excelize.NewFile()
	for i, record := range records {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			t.Fatal(err)
		}
		if err := book.SetSheetRow("Sheet1", cell, &record); err != nil {
			t.Fatal(err)
		}
	}
	var buf bytes.Buffer
	if err := book.Write(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
