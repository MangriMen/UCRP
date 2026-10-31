package e2e

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestDataSourceTypesDefaults(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/data-source-types", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/data-source-types: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/data-source-types")
	s.RequireDeclaredStatus("GET", "/api/v1/data-source-types", resp)
	s.RequireSchema("GET", "/api/v1/data-source-types", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode data source types: %v", err)
	}
	items, ok := body.([]any)
	if !ok {
		t.Fatalf("data source types response is not an array: %s", resp.Snippet())
	}
	codes := map[string]bool{}
	for _, item := range items {
		obj, _ := item.(map[string]any)
		codes[harness.StringField(obj, "code")] = true
	}
	for _, want := range []string{"csv_excel", "crm", "pos", "api"} {
		if !codes[want] {
			t.Fatalf("data source types must contain %q, got %v", want, codes)
		}
	}
}

func TestDataSourceCreate(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "source-create")
	ctx, cancel := s.Ctx()
	defer cancel()

	types, err := s.API.Get(ctx, "/api/v1/data-source-types", nil)
	if err != nil {
		t.Fatalf("GET data source types: %v", err)
	}
	typesBody, err := types.JSON()
	if err != nil {
		t.Fatalf("decode types: %v", err)
	}
	items, _ := typesBody.([]any)
	var csvTypeID string
	for _, item := range items {
		obj, _ := item.(map[string]any)
		if harness.StringField(obj, "code") == "csv_excel" {
			csvTypeID = harness.StringField(obj, "id")
		}
	}
	if csvTypeID == "" {
		t.Fatalf("csv_excel data source type is missing")
	}

	resp, err := s.API.Post(ctx, "/api/v1/data-sources", map[string]any{
		"tenantId":     w.tenantA.ID,
		"typeId":       csvTypeID,
		"name":         s.Reg.Name("source-created"),
		"fieldMapping": map[string]string{"externalId": "external_id", "occurredAt": "occurred_at"},
	})
	if err != nil {
		t.Fatalf("POST /api/v1/data-sources: %v", err)
	}
	s.RequireStatus(resp, 201, "POST /api/v1/data-sources")
	s.RequireDeclaredStatus("POST", "/api/v1/data-sources", resp)
	s.RequireSchema("POST", "/api/v1/data-sources", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode created source: %v", err)
	}
	obj, _ := body.(map[string]any)
	sourceID := harness.StringField(obj, "id")
	if sourceID == "" {
		t.Fatalf("created data source has no id: %s", resp.Snippet())
	}
	trackDataSource(s, sourceID)

	if active, ok := obj["isActive"].(bool); !ok || !active {
		t.Fatalf("created data source must be active, got %v", obj["isActive"])
	}
	if code := harness.StringField(obj, "typeCode"); code != "csv_excel" {
		t.Fatalf("type_code: expected csv_excel, got %q", code)
	}
}

func TestDataSourceCreateRequiresName(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/data-sources", map[string]any{"typeId": "00000000-0000-0000-0000-000000000000"})
	if err != nil {
		t.Fatalf("POST /api/v1/data-sources without name: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/data-sources without name")
	s.RequireDeclaredStatus("POST", "/api/v1/data-sources", resp)
	s.RequireSchema("POST", "/api/v1/data-sources", resp)
}

func TestDataSourceCreateRejectsBadMapping(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "source-bad-mapping")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/data-sources", map[string]any{
		"tenantId":     w.tenantA.ID,
		"typeId":       w.source.ID,
		"name":         s.Reg.Name("source-bad-mapping"),
		"fieldMapping": "{not-json",
	})
	if err != nil {
		t.Fatalf("POST /api/v1/data-sources with a broken mapping: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/data-sources with a broken field_mapping")
	s.RequireDeclaredStatus("POST", "/api/v1/data-sources", resp)
	s.RequireSchema("POST", "/api/v1/data-sources", resp)
}

func TestDataSourceGetUnknownReturns404(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/data-sources/00000000-0000-0000-0000-000000000000", nil)
	if err != nil {
		t.Fatalf("GET unknown data source: %v", err)
	}
	s.RequireStatus(resp, 404, "GET unknown data source")
	s.RequireDeclaredStatus("GET", "/api/v1/data-sources/{id}", resp)
	s.RequireSchema("GET", "/api/v1/data-sources/{id}", resp)
}

func TestDataSourceSyncRejectsFileOnlySource(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "source-sync")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/data-sources/"+w.source.ID+"/sync", nil)
	if err != nil {
		t.Fatalf("POST /api/v1/data-sources/{id}/sync: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/data-sources/{id}/sync on a file-only source")
	s.RequireDeclaredStatus("POST", "/api/v1/data-sources/{id}/sync", resp)
	s.RequireSchema("POST", "/api/v1/data-sources/{id}/sync", resp)
}

func TestDataSourceListContainsFixture(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "source-list")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/data-sources", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/data-sources: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/data-sources")
	s.RequireSchema("GET", "/api/v1/data-sources", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode data sources: %v", err)
	}
	items, _ := harness.Items(body)
	found := false
	for _, item := range items {
		obj, _ := item.(map[string]any)
		if harness.StringField(obj, "id") == w.source.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("data source list must contain fixture source %s", w.source.ID)
	}
}

func trackDataSource(s *harness.Suite, sourceID string) {
	s.T.Helper()
	if sourceID == "" {
		return
	}
	s.T.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM data_source WHERE id = $1`, sourceID)
	})
}

func TestDataSourceListFiltersAndTotal(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "source-list")
	ctx, cancel := s.Ctx()
	defer cancel()

	list := func(query url.Values) (int, []any) {
		t.Helper()
		resp, err := s.API.Get(ctx, "/api/v1/data-sources", query)
		if err != nil {
			t.Fatalf("GET /api/v1/data-sources: %v", err)
		}
		s.RequireStatus(resp, 200, "GET /api/v1/data-sources")
		s.RequireSchema("GET", "/api/v1/data-sources", resp)
		body, err := resp.JSON()
		if err != nil {
			t.Fatalf("decode sources: %v", err)
		}
		page, _ := body.(map[string]any)
		meta, _ := page["meta"].(map[string]any)
		total, _ := harness.IntField(meta, "total")
		items, _ := page["items"].([]any)
		return total, items
	}

	total, items := list(url.Values{"tenant_id": {w.tenantA.ID}, "page_size": {"1"}})
	if total != 1 || len(items) != 1 {
		t.Fatalf("tenant filter: expected total 1 and 1 item, got total %d, %d items", total, len(items))
	}
	source, _ := items[0].(map[string]any)
	if harness.StringField(source, "id") != w.source.ID || harness.StringField(source, "tenantName") != w.tenantA.Name {
		t.Fatalf("tenant filter returned the wrong source: %v", source)
	}

	if total, _ := list(url.Values{"tenant_id": {w.tenantA.ID}, "import_status": {"none"}}); total != 1 {
		t.Fatalf("import_status=none: expected the source without imports, got total %d", total)
	}
	if total, _ := list(url.Values{"tenant_id": {w.tenantA.ID}, "import_status": {"failed"}}); total != 0 {
		t.Fatalf("import_status=failed: expected no sources, got total %d", total)
	}
	if total, _ := list(url.Values{"tenant_id": {w.tenantA.ID}, "is_active": {"false"}}); total != 0 {
		t.Fatalf("is_active=false: expected no sources, got total %d", total)
	}

	resp, err := s.API.Get(ctx, "/api/v1/data-sources", url.Values{"sort": {"bogus"}})
	if err != nil {
		t.Fatalf("GET /api/v1/data-sources with a bad sort: %v", err)
	}
	s.RequireStatus(resp, 400, "GET /api/v1/data-sources with a bad sort")
}
