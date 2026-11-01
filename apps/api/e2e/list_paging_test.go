package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/harness"
)

func listPage(t *testing.T, s *harness.Suite, api *client.Client, path string, query map[string][]string) ([]any, int) {
	t.Helper()
	ctx, cancel := s.Ctx()
	defer cancel()
	resp, err := api.Get(ctx, path, query)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	s.RequireStatus(resp, 200, "GET "+path)
	s.RequireSchema("GET", path, resp)
	body := jsonObject(t, resp, path)
	items, ok := harness.Items(body)
	if !ok {
		t.Fatalf("GET %s has no items array: %s", path, resp.Snippet())
	}
	meta, _ := body["meta"].(map[string]any)
	total, ok := harness.IntField(meta, "total")
	if !ok {
		t.Fatalf("GET %s has no meta.total: %s", path, resp.Snippet())
	}
	return items, total
}

func TestCampaignListFiltersAndReportsTotal(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	w := buildWorld(t, s, "campaign-list")
	ctx, cancel := s.Ctx()
	defer cancel()

	first := mustSegment(t, s, w, "campaign-list-a")
	second := mustSegment(t, s, w, "campaign-list-b")
	for i, segmentID := range []string{first.ID, first.ID, second.ID} {
		resp, err := s.API.Post(ctx, "/api/v1/campaigns", map[string]any{
			"name":           s.Reg.Name("campaign-list-" + string(rune('a'+i))),
			"targetTenantId": w.tenantC.ID,
			"segmentId":      segmentID,
		})
		if err != nil {
			t.Fatalf("POST /api/v1/campaigns: %v", err)
		}
		s.RequireStatusOrDrift(resp, 201, "POST /api/v1/campaigns")
		trackCampaign(s, harness.StringField(jsonObject(t, resp, "created campaign"), "id"))
	}

	items, total := listPage(t, s, s.API, "/api/v1/campaigns", map[string][]string{"segment_id": {first.ID}, "page_size": {"1"}})
	if total != 2 || len(items) != 1 {
		t.Fatalf("segment filter: expected total 2 with one item on the page, got total=%d items=%d", total, len(items))
	}
	items, total = listPage(t, s, s.API, "/api/v1/campaigns", map[string][]string{"segment_id": {first.ID}, "page": {"2"}, "page_size": {"1"}})
	if total != 2 || len(items) != 1 {
		t.Fatalf("second page: expected total 2 with one item, got total=%d items=%d", total, len(items))
	}
	items, total = listPage(t, s, s.API, "/api/v1/campaigns", map[string][]string{"segment_id": {first.ID}, "status": {"archived"}})
	if total != 0 || len(items) != 0 {
		t.Fatalf("status filter: expected no archived campaigns, got total=%d items=%d", total, len(items))
	}
}

func TestSegmentListFiltersAndReportsTotal(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	first, err := s.Reg.Segment(ctx, "list-a", `{"visitedTenantIds":[],"excluded_tenant_ids":[],"period_days":90}`, nil)
	if err != nil {
		t.Fatalf("fixture segment: %v", err)
	}
	if _, err := s.Reg.Segment(ctx, "list-b", `{"visitedTenantIds":[],"excluded_tenant_ids":[],"period_days":90}`, nil); err != nil {
		t.Fatalf("fixture segment: %v", err)
	}
	archived, err := s.Reg.Segment(ctx, "list-c", `{"visitedTenantIds":[],"excluded_tenant_ids":[],"period_days":90}`, nil)
	if err != nil {
		t.Fatalf("fixture segment: %v", err)
	}
	if err := s.DB.Exec(ctx, `UPDATE segment SET status = 'archived' WHERE id = $1`, archived.ID); err != nil {
		t.Fatalf("archive fixture segment: %v", err)
	}
	prefix := s.Reg.Name("segment-list-")

	items, total := listPage(t, s, s.API, "/api/v1/segments", map[string][]string{"q": {prefix}, "page_size": {"1"}})
	if total != 2 || len(items) != 1 {
		t.Fatalf("q filter: expected total 2 (archived hidden) with one item, got total=%d items=%d", total, len(items))
	}
	items, total = listPage(t, s, s.API, "/api/v1/segments", map[string][]string{"q": {prefix}, "status": {"archived"}})
	if total != 1 || len(items) != 1 || harness.StringField(items[0].(map[string]any), "id") != archived.ID {
		t.Fatalf("archived filter: expected only %s, got total=%d items=%v", archived.ID, total, items)
	}
	items, total = listPage(t, s, s.API, "/api/v1/segments", map[string][]string{"q": {prefix + "a"}, "status": {"ready"}})
	if total != 1 || len(items) != 1 || harness.StringField(items[0].(map[string]any), "id") != first.ID {
		t.Fatalf("ready filter: expected only %s, got total=%d items=%v", first.ID, total, items)
	}
}

func TestImportListReportsTotal(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-list")
	ctx, cancel := s.Ctx()
	defer cancel()

	for _, phones := range [][2]string{{s.Reg.Phone(1), s.Reg.Phone(2)}, {s.Reg.Phone(3), s.Reg.Phone(4)}} {
		resp, err := uploadImport(ctx, s, iw.source.ID, "valid.csv", csvValid(iw.world.tenantA.Code, phones[0], phones[1]))
		if err != nil {
			t.Fatalf("POST /api/v1/imports: %v", err)
		}
		s.RequireStatus(resp, 202, "POST /api/v1/imports")
		trackImport(s, harness.StringField(jsonObject(t, resp, "import"), "batchId"))
	}

	items, total := listPage(t, s, s.API, "/api/v1/imports", map[string][]string{"data_source_id": {iw.source.ID}, "page_size": {"1"}})
	if total != 2 || len(items) != 1 {
		t.Fatalf("data source filter: expected total 2 with one item, got total=%d items=%d", total, len(items))
	}
	items, total = listPage(t, s, s.API, "/api/v1/imports", map[string][]string{"data_source_id": {iw.source.ID}, "status": {"failed"}})
	if total != 0 || len(items) != 0 {
		t.Fatalf("status filter: expected no batches, got total=%d items=%d", total, len(items))
	}
}

func TestAuditLogListFiltersAndReportsTotal(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	entityID := s.Reg.NewID()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM audit_log WHERE entity_id = $1`, entityID)
	})
	now := time.Now().UTC().Truncate(time.Second)
	for _, row := range []struct {
		action string
		when   time.Time
	}{
		{"e2e.list.first", now.Add(-2 * time.Hour)},
		{"e2e.list.second", now.Add(-time.Hour)},
		{"e2e.list.second", now},
	} {
		if err := s.DB.Exec(ctx, `
			INSERT INTO audit_log (id, actor_role, action, entity_type, entity_id, details, created_at)
			VALUES ($1,'analyst',$2,'e2e_list',$3,'{}',$4)`, s.Reg.NewID(), row.action, entityID, row.when); err != nil {
			t.Fatalf("insert audit fixture: %v", err)
		}
	}

	items, total := listPage(t, s, s.API, "/api/v1/audit-logs", map[string][]string{"entity_id": {entityID}, "page_size": {"2"}})
	if total != 3 || len(items) != 2 {
		t.Fatalf("entity filter: expected total 3 with two items, got total=%d items=%d", total, len(items))
	}
	items, total = listPage(t, s, s.API, "/api/v1/audit-logs", map[string][]string{"entity_id": {entityID}, "action": {"e2e.list.second"}})
	if total != 2 || len(items) != 2 {
		t.Fatalf("action filter: expected 2, got total=%d items=%d", total, len(items))
	}
	_, total = listPage(t, s, s.API, "/api/v1/audit-logs", map[string][]string{
		"entity_id": {entityID},
		"from":      {now.Add(-90 * time.Minute).Format(time.RFC3339)},
		"to":        {now.Add(-30 * time.Minute).Format(time.RFC3339)},
	})
	if total != 1 {
		t.Fatalf("period filter: expected 1, got total=%d", total)
	}

	bad, err := s.API.Get(ctx, "/api/v1/audit-logs", map[string][]string{"from": {"yesterday"}})
	if err != nil {
		t.Fatalf("GET /api/v1/audit-logs: %v", err)
	}
	s.RequireStatus(bad, 400, "GET /api/v1/audit-logs with a malformed from")
}
