package handler

import (
	"testing"

	"github.com/MangriMen/UCRP/internal/service"
)

func exportWithCounts(total, noContact, blocked, control, exported int) service.ExportDTO {
	withContact := total - noContact
	return service.ExportDTO{
		TotalInSegment: &total, WithContactCount: &withContact, NoContactCount: &noContact,
		BlockedByLimitCount: &blocked, ControlCount: &control, ExportedCount: &exported,
	}
}

func TestSuppressSmallExportCounts_HidesWholeGroupWhenAnyCountIsBelowK(t *testing.T) {
	item := exportWithCounts(40, 10, 3, 9, 18)
	suppressSmallExportCounts(&item)
	if !item.Suppressed {
		t.Fatal("expected suppressed")
	}
	for name, count := range map[string]*int{
		"total": item.TotalInSegment, "withContact": item.WithContactCount, "noContact": item.NoContactCount,
		"blocked": item.BlockedByLimitCount, "control": item.ControlCount, "exported": item.ExportedCount,
	} {
		if count != nil {
			t.Fatalf("%s = %d, want hidden", name, *count)
		}
	}
}

func TestSuppressSmallExportCounts_KeepsZerosAndLargeCounts(t *testing.T) {
	item := exportWithCounts(40, 10, 0, 6, 24)
	suppressSmallExportCounts(&item)
	if item.Suppressed || item.TotalInSegment == nil || *item.BlockedByLimitCount != 0 {
		t.Fatalf("unexpected suppression: %+v", item)
	}
}
