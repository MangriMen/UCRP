package e2e

import (
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestDataQualitySuppressesSmallGroups(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "data-quality-access")
	ctx, cancel := s.Ctx()
	defer cancel()
	response, err := s.API.Get(ctx, "/api/v1/data-quality/sources", nil)
	if err != nil {
		t.Fatal(err)
	}
	s.RequireStatus(response, 200, "GET /api/v1/data-quality/sources")
	s.RequireSchema("GET", "/api/v1/data-quality/sources", response)
	body, err := response.JSON()
	if err != nil {
		t.Fatal(err)
	}
	items, ok := harness.Items(body)
	if !ok {
		t.Fatalf("missing data quality items: %s", response.Snippet())
	}
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok || harness.StringField(row, "dataSourceId") != w.source.ID {
			continue
		}
		if row["suppressed"] != true || row["pendingMergeCandidates"] != nil || row["rejectedShare30d"] != nil || row["hasWarning"] != nil {
			t.Fatalf("small source metrics exposed: %s", response.Snippet())
		}
		return
	}
	t.Fatalf("fixture source missing: %s", response.Snippet())
}
