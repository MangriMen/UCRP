package e2e

import (
	"strings"
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestPenetrationBucketsContract(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "penetration")
	ctx, cancel := s.Ctx()
	defer cancel()

	single, err := s.Reg.Guest(ctx, "penetration-single")
	if err != nil {
		t.Fatalf("fixture single-tenant guest: %v", err)
	}
	pair, err := s.Reg.Guest(ctx, "penetration-pair")
	if err != nil {
		t.Fatalf("fixture two-tenant guest: %v", err)
	}
	w.visit(t, s, single.ID, w.tenantA.ID, "pen-single", 3)
	w.visit(t, s, pair.ID, w.tenantA.ID, "pen-pair-a", 3)
	w.visit(t, s, pair.ID, w.tenantB.ID, "pen-pair-b", 2)

	resp, err := s.API.Get(ctx, "/api/v1/dashboard/penetration", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/dashboard/penetration: %v", err)
	}
	s.RequireStatusOrDrift(resp, 200, "GET /api/v1/dashboard/penetration")
	s.RequireDeclaredStatus("GET", "/api/v1/dashboard/penetration", resp)
	s.RequireSchema("GET", "/api/v1/dashboard/penetration", resp)

	report := jsonObject(t, resp, "penetration report")
	one, _ := harness.IntField(report, "one")
	two, _ := harness.IntField(report, "two")
	threePlus, _ := harness.IntField(report, "three_plus")
	buckets, ok := report["buckets"].([]any)
	if !ok || len(buckets) != 3 {
		t.Fatalf("the penetration report must expose three buckets: %s", resp.Snippet())
	}

	var bucketTotal int
	for _, raw := range buckets {
		bucket, _ := raw.(map[string]any)
		count, _ := harness.IntField(bucket, "count")
		bucketTotal += count
	}
	if bucketTotal != one+two+threePlus {
		t.Fatalf("bucket counts (%d) must add up to one+two+three_plus (%d)", bucketTotal, one+two+threePlus)
	}
	if one < 1 || two < 1 {
		t.Fatalf("the fixture guests must land in the 1-tenant and 2-tenant buckets, got one=%d two=%d", one, two)
	}
}

func TestPenetrationRejectsInvertedPeriod(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/dashboard/penetration", map[string][]string{
		"from": {"2026-12-01T00:00:00Z"},
		"to":   {"2026-10-01T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("GET /api/v1/dashboard/penetration: %v", err)
	}
	s.RequireStatus(resp, 400, "GET /api/v1/dashboard/penetration with from > to")
	s.RequireDeclaredStatus("GET", "/api/v1/dashboard/penetration", resp)
	s.RequireSchema("GET", "/api/v1/dashboard/penetration", resp)
}

func TestPenetrationRejectsOversizedPeriod(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/dashboard/penetration", map[string][]string{
		"from": {"2024-01-01T00:00:00Z"},
		"to":   {"2026-10-01T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("GET /api/v1/dashboard/penetration: %v", err)
	}
	s.RequireStatus(resp, 400, "GET /api/v1/dashboard/penetration with a period longer than 24 months")
	s.RequireDeclaredStatus("GET", "/api/v1/dashboard/penetration", resp)
}

func TestDashboardMetricsFields(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/dashboard/metrics", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/dashboard/metrics: %v", err)
	}
	s.RequireStatusOrDrift(resp, 200, "GET /api/v1/dashboard/metrics")
	s.RequireDeclaredStatus("GET", "/api/v1/dashboard/metrics", resp)
	s.RequireSchema("GET", "/api/v1/dashboard/metrics", resp)

	metrics := jsonObject(t, resp, "dashboard metrics")
	for _, key := range []string{"match_rate", "contested_merge_rate", "avg_segment_to_export_minutes"} {
		value, ok := metrics[key]
		if !ok {
			t.Fatalf("the metrics payload is missing %q: %s", key, resp.Snippet())
		}
		rate, ok := value.(float64)
		if !ok {
			t.Fatalf("%q must be numeric, got %T", key, value)
		}
		if rate < 0 || rate > 1 {
			t.Fatalf("%q must stay within 0..1, got %v", key, rate)
		}
	}
}

func TestTenantAggregatesRespectsKThreshold(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "aggregates-k")
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "aggregates-k")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "aggregates-k", 3)

	resp, err := s.API.Get(ctx, "/api/v1/tenants/"+w.tenantA.ID+"/aggregates", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/tenants/{id}/aggregates: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/tenants/{id}/aggregates")
	s.RequireDeclaredStatus("GET", "/api/v1/tenants/{id}/aggregates", resp)
	s.RequireSchema("GET", "/api/v1/tenants/{id}/aggregates", resp)

	aggregate := jsonObject(t, resp, "tenant aggregate")
	hidden, ok := aggregate["hidden"].(bool)
	if !ok {
		t.Fatalf("the aggregate must report a hidden flag: %s", resp.Snippet())
	}
	if !hidden {
		t.Fatalf("a single guest is below the k = 5 threshold and must be suppressed")
	}
	if value, exists := aggregate["new_guests"]; exists {
		if n, _ := value.(float64); n != 0 {
			t.Fatalf("a suppressed aggregate must not leak the value, got %v", value)
		}
	}
}

func TestTenantAggregatesWithoutGuestIDs(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "aggregates-shape")
	ctx, cancel := s.Ctx()
	defer cancel()

	guests := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		guest, err := s.Reg.Guest(ctx, "aggregates-shape-"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("fixture guest %d: %v", i, err)
		}
		guests = append(guests, guest.ID)
		w.visit(t, s, guest.ID, w.tenantB.ID, "aggregates-shape-"+string(rune('a'+i)), 2)
	}

	resp, err := s.API.Get(ctx, "/api/v1/tenants/"+w.tenantB.ID+"/aggregates", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/tenants/{id}/aggregates: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/tenants/{id}/aggregates above the threshold")

	raw := string(resp.Body)
	for _, guestID := range guests {
		if contains(raw, guestID) {
			t.Fatalf("the aggregate response must not expose guest ids, found %s in %s", guestID, raw)
		}
	}
	for _, forbidden := range []string{"guest_id", "guestId"} {
		if contains(raw, forbidden) {
			t.Fatalf("the aggregate response must not carry a %q field: %s", forbidden, raw)
		}
	}
}

func TestTenantAggregatesTenantRoleIsolation(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "aggregates-role")
	ctx, cancel := s.Ctx()
	defer cancel()

	user, err := s.Reg.User(ctx, "aggregates-tenant", "tenant", w.tenantA.ID)
	if err != nil {
		t.Fatalf("fixture tenant user: %v", err)
	}
	tenantClient := s.As(user.Email, user.Password)

	own, err := tenantClient.Get(ctx, "/api/v1/tenants/"+w.tenantA.ID+"/aggregates", nil)
	if err != nil {
		t.Fatalf("GET own aggregates as tenant: %v", err)
	}
	s.RequireStatus(own, 200, "GET /api/v1/tenants/{own}/aggregates as tenant")

	foreign, err := tenantClient.Get(ctx, "/api/v1/tenants/"+w.tenantC.ID+"/aggregates", nil)
	if err != nil {
		t.Fatalf("GET foreign aggregates as tenant: %v", err)
	}
	s.RequireStatus(foreign, 403, "GET /api/v1/tenants/{foreign}/aggregates as tenant")
	s.RequireDeclaredStatus("GET", "/api/v1/tenants/{id}/aggregates", foreign)
}

func TestDataQualitySourcesList(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	iw := buildImportWorld(t, s, "dq-sources")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/data-quality/sources", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/data-quality/sources: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/data-quality/sources")
	s.RequireDeclaredStatus("GET", "/api/v1/data-quality/sources", resp)
	s.RequireSchema("GET", "/api/v1/data-quality/sources", resp)

	items, ok := harness.Items(mustJSON(t, resp, "data quality list"))
	if !ok {
		t.Fatalf("the data quality payload must expose items: %s", resp.Snippet())
	}
	sourceIndex := -1
	for i, item := range items {
		row, _ := item.(map[string]any)
		if harness.StringField(row, "data_source_id") == iw.source.ID {
			sourceIndex = i
			break
		}
	}
	if sourceIndex < 0 {
		t.Fatalf("the fixture data source %s is absent from the data quality list", iw.source.ID)
	}
	row, _ := items[sourceIndex].(map[string]any)
	if _, exists := row["warning"]; !exists {
		t.Fatalf("a data quality row must carry a warning flag: %s", resp.Snippet())
	}
	if _, exists := row["rejected_rate"]; !exists {
		t.Fatalf("a data quality row must carry rejected_rate: %s", resp.Snippet())
	}
}

func TestDataQualityForbiddenForMarketer(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	marketer := actor(t, s, "marketer")
	resp, err := marketer.Get(ctx, "/api/v1/data-quality/sources", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/data-quality/sources as marketer: %v", err)
	}
	s.RequireStatus(resp, 403, "GET /api/v1/data-quality/sources as marketer")
}

func contains(haystack, needle string) bool {
	return needle != "" && strings.Contains(haystack, needle)
}
