package service

import (
	"encoding/json"
	"testing"

	"github.com/MangriMen/UCRP/internal/domain"
)

func TestExportPreviewSuppression(t *testing.T) {
	cases := []struct {
		name                                       string
		total, withContact, blocked, control, send int
		suppressed                                 bool
	}{
		{"all groups at threshold", 20, 15, 5, 5, 5, false},
		{"derived missing contacts below threshold", 19, 15, 5, 5, 5, true},
		{"blocked below threshold", 21, 16, 4, 6, 6, true},
		{"zero control group", 20, 15, 5, 0, 10, true},
		{"small segment", 4, 4, 0, 0, 4, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := exportPreviewResponse(tc.total, tc.withContact, tc.blocked, tc.control, tc.send)
			counts := []*int{result.InSegment, result.WithContact, result.BlockedByLimit, result.Control, result.ToSend}
			if result.Suppressed != tc.suppressed {
				t.Fatalf("suppressed = %t", result.Suppressed)
			}
			for _, count := range counts {
				if (count == nil) != tc.suppressed {
					t.Fatalf("count = %v, suppressed = %t", count, tc.suppressed)
				}
			}
			if !tc.suppressed && (*result.InSegment != tc.total || *result.WithContact != tc.withContact || *result.BlockedByLimit != tc.blocked || *result.Control != tc.control || *result.ToSend != tc.send) {
				t.Fatal("unsuppressed counts changed")
			}
			assertExportNulls(t, result, []string{"in_segment", "with_contact", "blocked_by_limit", "control", "to_send"}, tc.suppressed)
		})
	}
}

func TestExportDTOSuppression(t *testing.T) {
	cases := []struct {
		name       string
		job        domain.ExportJob
		suppressed bool
	}{
		{"all groups at threshold", domain.ExportJob{TotalInSegment: 20, NoContactCount: 5, BlockedByLimitCount: 5, ControlCount: 5, ExportedCount: 5}, false},
		{"small contact exclusion", domain.ExportJob{TotalInSegment: 19, NoContactCount: 4, BlockedByLimitCount: 5, ControlCount: 5, ExportedCount: 5}, true},
		{"zero blocked count", domain.ExportJob{TotalInSegment: 15, NoContactCount: 5, BlockedByLimitCount: 0, ControlCount: 5, ExportedCount: 5}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := toExportDTO(tc.job)
			if result.Suppressed != tc.suppressed {
				t.Fatalf("suppressed = %t", result.Suppressed)
			}
			assertExportNulls(t, result, []string{"total_in_segment", "no_contact_count", "blocked_by_limit_count", "control_count", "exported_count"}, tc.suppressed)
			if !tc.suppressed && (*result.TotalInSegment != tc.job.TotalInSegment || *result.NoContactCount != tc.job.NoContactCount || *result.BlockedByLimitCount != tc.job.BlockedByLimitCount || *result.ControlCount != tc.job.ControlCount || *result.ExportedCount != tc.job.ExportedCount) {
				t.Fatal("unsuppressed counts changed")
			}
		})
	}
}

func assertExportNulls(t *testing.T, value any, fields []string, suppressed bool) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, field := range fields {
		value, exists := decoded[field]
		if !exists || (value == nil) != suppressed {
			t.Fatalf("%s = %v, present = %t, suppressed = %t", field, value, exists, suppressed)
		}
	}
	if decoded["suppressed"] != suppressed {
		t.Fatalf("suppressed = %v", decoded["suppressed"])
	}
}
