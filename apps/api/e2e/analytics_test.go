package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestCrossTenantDistributionsContract(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "cross-tenant")
	ctx, cancel := s.Ctx()
	defer cancel()

	for i := 0; i < 5; i++ {
		g, err := s.Reg.Guest(ctx, "cross-tenant-single-"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("fixture single-tenant guest %d: %v", i, err)
		}
		w.visit(t, s, g.ID, w.tenantA.ID, "pen-single-"+string(rune('a'+i)), 3)
	}
	for i := 0; i < 5; i++ {
		g, err := s.Reg.Guest(ctx, "cross-tenant-pair-"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("fixture two-tenant guest %d: %v", i, err)
		}
		w.visit(t, s, g.ID, w.tenantA.ID, "pen-pair-a-"+string(rune('a'+i)), 3)
		w.visit(t, s, g.ID, w.tenantB.ID, "pen-pair-b-"+string(rune('a'+i)), 2)
	}
	for i := 0; i < 5; i++ {
		g, err := s.Reg.Guest(ctx, "cross-tenant-triple-"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("fixture three-tenant guest %d: %v", i, err)
		}
		w.visit(t, s, g.ID, w.tenantA.ID, "pen-triple-a-"+string(rune('a'+i)), 3)
		w.visit(t, s, g.ID, w.tenantB.ID, "pen-triple-b-"+string(rune('a'+i)), 2)
		w.visit(t, s, g.ID, w.tenantC.ID, "pen-triple-c-"+string(rune('a'+i)), 1)
	}

	resp, err := s.API.Get(ctx, "/api/v1/dashboard/cross-tenant", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/dashboard/cross-tenant: %v", err)
	}
	s.RequireStatusOrDrift(resp, 200, "GET /api/v1/dashboard/cross-tenant")
	s.RequireDeclaredStatus("GET", "/api/v1/dashboard/cross-tenant", resp)
	s.RequireSchema("GET", "/api/v1/dashboard/cross-tenant", resp)

	report := jsonObject(t, resp, "cross-tenant report")
	distribution, ok := report["distribution"].(map[string]any)
	if !ok {
		t.Fatalf("the cross-tenant report must expose a distribution object: %s", resp.Snippet())
	}
	one, oneOK := harness.IntField(distribution, "one")
	two, twoOK := harness.IntField(distribution, "two")
	threePlus, threeOK := harness.IntField(distribution, "threePlus")
	if !oneOK || !twoOK || !threeOK {
		t.Fatalf("raw distribution buckets must be visible for five or more guests each: %s", resp.Snippet())
	}
	if one < 5 || two < 5 || threePlus < 5 {
		t.Fatalf("distribution buckets must each carry at least five fixture guests, got one=%d two=%d threePlus=%d", one, two, threePlus)
	}
	series, ok := report["series"].([]any)
	if !ok || len(series) == 0 {
		t.Fatalf("the cross-tenant report must expose a series: %s", resp.Snippet())
	}
}

func TestCrossTenantRejectsInvertedPeriod(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/dashboard/cross-tenant", map[string][]string{
		"from": {"2026-12-01T00:00:00Z"},
		"to":   {"2026-10-01T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("GET /api/v1/dashboard/cross-tenant: %v", err)
	}
	s.RequireStatus(resp, 400, "GET /api/v1/dashboard/cross-tenant with from > to")
	s.RequireDeclaredStatus("GET", "/api/v1/dashboard/cross-tenant", resp)
	s.RequireSchema("GET", "/api/v1/dashboard/cross-tenant", resp)
}

func TestCrossTenantRejectsOversizedPeriod(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/dashboard/cross-tenant", map[string][]string{
		"from": {"2024-01-01T00:00:00Z"},
		"to":   {"2026-10-01T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("GET /api/v1/dashboard/cross-tenant: %v", err)
	}
	s.RequireStatus(resp, 400, "GET /api/v1/dashboard/cross-tenant with a period longer than 24 months")
	s.RequireDeclaredStatus("GET", "/api/v1/dashboard/cross-tenant", resp)
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
	for _, key := range []string{"matchRate", "disputedMergeRate"} {
		value, ok := metrics[key]
		if !ok {
			t.Fatalf("the metrics payload is missing %q: %s", key, resp.Snippet())
		}
		if value == nil {
			continue
		}
		rate, ok := value.(float64)
		if !ok {
			t.Fatalf("%q must be numeric or null, got %T", key, value)
		}
		if rate < 0 || rate > 1 {
			t.Fatalf("%q must stay within 0..1, got %v", key, rate)
		}
	}
	for _, key := range []string{"segmentToExportMedianSeconds", "segmentToExportP95Seconds"} {
		value, ok := metrics[key]
		if !ok {
			t.Fatalf("the metrics payload is missing %q: %s", key, resp.Snippet())
		}
		if value == nil {
			continue
		}
		seconds, ok := value.(float64)
		if !ok {
			t.Fatalf("%q must be numeric or null, got %T", key, value)
		}
		if seconds < 0 {
			t.Fatalf("%q must not be negative, got %v", key, seconds)
		}
	}
	sampleSizes, ok := metrics["sampleSizes"].(map[string]any)
	if !ok {
		t.Fatalf("the metrics payload is missing sampleSizes: %s", resp.Snippet())
	}
	for _, key := range []string{"events", "mergeCandidates", "exports", "contacts"} {
		size, ok := harness.IntField(sampleSizes, key)
		if !ok || size < 0 {
			t.Fatalf("sampleSizes.%s must be a non-negative integer: %s", key, resp.Snippet())
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
	suppressed, ok := aggregate["suppressed"].(bool)
	if !ok {
		t.Fatalf("the aggregate must report a suppressed flag: %s", resp.Snippet())
	}
	if !suppressed {
		t.Fatalf("a single guest is below the k = 5 threshold and must be suppressed")
	}
	if value, exists := aggregate["newGuests"]; exists {
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
	manager, err := s.Reg.User(ctx, "aggregates-manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}
	managerClient := s.As(manager.Email, manager.Password)

	withoutGrant, err := s.As(user.Email, user.Password).Get(ctx, "/api/v1/tenants/"+w.tenantA.ID+"/aggregates", nil)
	if err != nil {
		t.Fatalf("GET own aggregates as tenant without a grant: %v", err)
	}
	s.RequireStatus(withoutGrant, 403, "GET /api/v1/tenants/{own}/aggregates as tenant without a grant")

	grant, err := managerClient.Post(ctx, "/api/v1/access/grants", map[string]any{
		"subjectType": "tenant",
		"subjectId":   w.tenantA.ID,
		"status":      "active",
		"validFrom":   time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		"validTo":     time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		"reason":      "E2E aggregates isolation",
		"items": []map[string]string{
			{"permission": "tenant.aggregates.own", "effect": "allow"},
			{"permission": "metric.visits_count.read", "effect": "allow"},
		},
		"scopes": []map[string]string{{"type": "tenant", "value": w.tenantA.ID}},
	})
	if err != nil {
		t.Fatalf("create aggregates grant: %v", err)
	}
	s.RequireStatus(grant, 201, "manager grants tenant.aggregates.own")

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
		if harness.StringField(row, "dataSourceId") == iw.source.ID {
			sourceIndex = i
			break
		}
	}
	if sourceIndex < 0 {
		t.Fatalf("the fixture data source %s is absent from the data quality list", iw.source.ID)
	}
	row, _ := items[sourceIndex].(map[string]any)
	if _, exists := row["hasWarning"]; !exists {
		t.Fatalf("a data quality row must carry a warning flag: %s", resp.Snippet())
	}
	if _, exists := row["rejectedShare30d"]; !exists {
		t.Fatalf("a data quality row must carry rejectedShare30d: %s", resp.Snippet())
	}
	if _, exists := row["suppressed"]; !exists {
		t.Fatalf("a data quality row must carry a suppressed flag: %s", resp.Snippet())
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

func TestDataQualitySourceWithoutImports(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	iw := buildImportWorld(t, s, "dq-source-empty")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/data-quality/sources/"+iw.source.ID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/data-quality/sources/{id}: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/data-quality/sources/{id}")
	s.RequireDeclaredStatus("GET", "/api/v1/data-quality/sources/{id}", resp)
	s.RequireSchema("GET", "/api/v1/data-quality/sources/{id}", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode data quality source: %v", err)
	}
	obj, _ := body.(map[string]any)
	if name := harness.StringField(obj, "dataSourceName"); name != iw.source.Name {
		t.Fatalf("data source name: expected %q, got %q", iw.source.Name, name)
	}
	if series, ok := obj["series"].([]any); !ok || len(series) != 0 {
		t.Fatalf("a source without imports must expose an empty series: %s", resp.Snippet())
	}
	if batches, ok := obj["recentBatches"].([]any); !ok || len(batches) != 0 {
		t.Fatalf("a source without imports must expose empty recentBatches: %s", resp.Snippet())
	}
}

func TestDataQualitySourceDetail(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	iw := buildImportWorld(t, s, "dq-source-detail")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := uploadImport(ctx, s, iw.source.ID, "valid.csv", csvValid(iw.world.tenantA.Code, s.Reg.Phone(11), s.Reg.Phone(12)))
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batchId")
	trackImport(s, batchID)

	detail, err := s.API.Get(ctx, "/api/v1/data-quality/sources/"+iw.source.ID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/data-quality/sources/{id}: %v", err)
	}
	s.RequireStatus(detail, 200, "GET /api/v1/data-quality/sources/{id}")
	s.RequireSchema("GET", "/api/v1/data-quality/sources/{id}", detail)

	decoded, err := detail.JSON()
	if err != nil {
		t.Fatalf("decode data quality source: %v", err)
	}
	row, _ := decoded.(map[string]any)
	batches, ok := row["recentBatches"].([]any)
	if !ok || len(batches) == 0 {
		t.Fatalf("a source with an import must expose recentBatches: %s", detail.Snippet())
	}
	first, _ := batches[0].(map[string]any)
	if id := harness.StringField(first, "id"); id != batchID {
		t.Fatalf("recentBatches[0].id: expected %q, got %q", batchID, id)
	}
	if series, ok := row["series"].([]any); !ok || len(series) == 0 {
		t.Fatalf("a source with an import must expose a non-empty series: %s", detail.Snippet())
	}
}

func TestDataQualitySourceNotFound(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/data-quality/sources/00000000-0000-0000-0000-000000000000", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/data-quality/sources/{id}: %v", err)
	}
	s.RequireStatus(resp, 404, "GET /api/v1/data-quality/sources/{id} for an unknown source")
	s.RequireDeclaredStatus("GET", "/api/v1/data-quality/sources/{id}", resp)
}

func contains(haystack, needle string) bool {
	return needle != "" && strings.Contains(haystack, needle)
}
