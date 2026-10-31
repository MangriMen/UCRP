package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/fixtures"
	"github.com/MangriMen/UCRP/e2e/harness"
)

type exportWorld struct {
	world   *world
	segment *fixtures.Segment
	guests  []*fixtures.Guest
}

func buildExportWorld(t *testing.T, s *harness.Suite, suffix string, guestCount int) *exportWorld {
	t.Helper()
	w := buildWorld(t, s, suffix)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	guests := make([]*fixtures.Guest, 0, guestCount)
	for i := 0; i < guestCount; i++ {
		g, err := s.Reg.Guest(ctx, suffix+"-g"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("fixture guest %d: %v", i, err)
		}
		guests = append(guests, g)
	}
	segment, err := s.Reg.SegmentWithContacts(ctx, suffix, guests, "phone", w.source.ID)
	if err != nil {
		t.Fatalf("fixture segment with contacts: %v", err)
	}
	return &exportWorld{world: w, segment: segment, guests: guests}
}

func exportPreview(ctx context.Context, c *client.Client, segmentID, channel string) (*client.Response, error) {
	return c.Post(ctx, "/api/v1/exports/preview", map[string]any{
		"segment_id": segmentID,
		"channel":    channel,
	})
}

func exportCreate(ctx context.Context, c *client.Client, segmentID, key string) (*client.Response, error) {
	return c.Post(ctx, "/api/v1/exports", map[string]any{
		"segment_id":      segmentID,
		"channel":         "sms",
		"format":          "csv",
		"idempotency_key": key,
	})
}

func TestExportPreviewCounters(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "export-preview", 3)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := exportPreview(ctx, s.API, ew.segment.ID, "sms")
	if err != nil {
		t.Fatalf("POST /api/v1/exports/preview: %v", err)
	}
	s.RequireStatus(resp, 200, "POST /api/v1/exports/preview")
	s.RequireDeclaredStatus("POST", "/api/v1/exports/preview", resp)
	s.RequireSchema("POST", "/api/v1/exports/preview", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	obj, _ := body.(map[string]any)
	assertExportCountsSuppressed(t, obj, "in_segment", "with_contact", "blocked_by_limit", "control", "to_send")
}

func TestExportPreviewChannelContactFilter(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "export-channel", 2)
	ctx, cancel := s.Ctx()
	defer cancel()

	sms, err := exportPreview(ctx, s.API, ew.segment.ID, "sms")
	if err != nil {
		t.Fatalf("preview sms: %v", err)
	}
	s.RequireStatus(sms, 200, "preview sms")
	smsBody, err := sms.JSON()
	if err != nil {
		t.Fatalf("decode sms preview: %v", err)
	}
	smsObj, _ := smsBody.(map[string]any)
	assertExportCountsSuppressed(t, smsObj, "in_segment", "with_contact", "blocked_by_limit", "control", "to_send")

	email, err := exportPreview(ctx, s.API, ew.segment.ID, "email")
	if err != nil {
		t.Fatalf("preview email: %v", err)
	}
	s.RequireStatus(email, 200, "preview email")
	emailBody, err := email.JSON()
	if err != nil {
		t.Fatalf("decode email preview: %v", err)
	}
	emailObj, _ := emailBody.(map[string]any)
	assertExportCountsSuppressed(t, emailObj, "in_segment", "with_contact", "blocked_by_limit", "control", "to_send")
}

func TestExportCreateReady(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "export-create", 2)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("key-create"))
	if err != nil {
		t.Fatalf("POST /api/v1/exports: %v", err)
	}
	s.RequireStatus(resp, 201, "POST /api/v1/exports")
	s.RequireDeclaredStatus("POST", "/api/v1/exports", resp)
	s.RequireSchema("POST", "/api/v1/exports", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode export: %v", err)
	}
	obj, _ := body.(map[string]any)
	exportID := harness.StringField(obj, "id")
	if exportID == "" {
		t.Fatalf("export response has no id: %s", resp.Snippet())
	}
	trackExport(s, exportID)

	if status := harness.StringField(obj, "status"); status != "ready" {
		t.Fatalf("export status: expected ready, got %q", status)
	}
	assertExportCountsSuppressed(t, obj, "total_in_segment", "no_contact_count", "blocked_by_limit_count", "control_count", "exported_count")
	var exported int
	if err := s.DB.QueryRowContext(ctx, `SELECT exported_count FROM export_job WHERE id = $1`, exportID).Scan(&exported); err != nil {
		t.Fatalf("read exported count: %v", err)
	}
	if exported != 2 {
		t.Fatalf("exported_count: expected 2, got %d", exported)
	}
}

func TestExportCreateRequiresIdempotencyKey(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "export-no-key", 1)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := exportCreate(ctx, s.API, ew.segment.ID, "")
	if err != nil {
		t.Fatalf("POST /api/v1/exports without a key: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/exports without Idempotency-Key")
	s.RequireDeclaredStatus("POST", "/api/v1/exports", resp)
	s.RequireSchema("POST", "/api/v1/exports", resp)
}

func TestExportCreateIsIdempotent(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "export-idempotent", 2)
	ctx, cancel := s.Ctx()
	defer cancel()

	key := s.Reg.Name("key-idempotent")
	first, err := exportCreate(ctx, s.API, ew.segment.ID, key)
	if err != nil {
		t.Fatalf("first export: %v", err)
	}
	s.RequireStatus(first, 201, "first export")
	firstBody, err := first.JSON()
	if err != nil {
		t.Fatalf("decode first export: %v", err)
	}
	firstObj, _ := firstBody.(map[string]any)
	firstID := harness.StringField(firstObj, "id")
	trackExport(s, firstID)

	second, err := exportCreate(ctx, s.API, ew.segment.ID, key)
	if err != nil {
		t.Fatalf("second export: %v", err)
	}
	s.RequireStatus(second, 201, "second export with the same key")
	secondBody, err := second.JSON()
	if err != nil {
		t.Fatalf("decode second export: %v", err)
	}
	secondObj, _ := secondBody.(map[string]any)
	secondID := harness.StringField(secondObj, "id")
	if secondID != firstID {
		t.Fatalf("repeated Idempotency-Key must return the same export: first=%s second=%s", firstID, secondID)
	}

	logs, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM contact_log WHERE export_id = $1`, firstID)
	if err != nil {
		t.Fatalf("count contact logs: %v", err)
	}
	if logs != 2 {
		t.Fatalf("repeated export must not add contact_log rows: expected 2, got %d", logs)
	}
}

func TestExportDownloadExcludesControlGroup(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "export-control", 4)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/exports", map[string]any{
		"segment_id":      ew.segment.ID,
		"channel":         "sms",
		"format":          "csv",
		"control_percent": 50,
		"idempotency_key": s.Reg.Name("key-control"),
	})
	if err != nil {
		t.Fatalf("POST /api/v1/exports: %v", err)
	}
	s.RequireStatus(resp, 201, "POST /api/v1/exports with a control group")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode export: %v", err)
	}
	obj, _ := body.(map[string]any)
	exportID := harness.StringField(obj, "id")
	trackExport(s, exportID)

	assertExportCountsSuppressed(t, obj, "total_in_segment", "no_contact_count", "blocked_by_limit_count", "control_count", "exported_count")
	var control, exported int
	if err := s.DB.QueryRowContext(ctx, `SELECT control_count, exported_count FROM export_job WHERE id = $1`, exportID).Scan(&control, &exported); err != nil {
		t.Fatalf("read export counts: %v", err)
	}
	if control == 0 {
		t.Fatalf("control_count: expected a non-empty control group at 50%%, got 0")
	}
	if control+exported != 4 {
		t.Fatalf("control(%d) + exported(%d) must equal the segment size 4", control, exported)
	}

	download, err := s.API.Get(ctx, "/api/v1/exports/"+exportID+"/download", nil)
	if err != nil {
		t.Fatalf("GET export download: %v", err)
	}
	s.RequireStatus(download, 200, "GET /api/v1/exports/{id}/download")

	lines := strings.Split(strings.TrimSpace(string(download.Body)), "\n")
	if len(lines) != exported+1 {
		t.Fatalf("CSV must contain a header plus %d recipients (control group excluded), got %d lines",
			exported, len(lines))
	}
	if !strings.HasPrefix(lines[0], "guest_id") {
		t.Fatalf("CSV header: expected guest_id column, got %q", lines[0])
	}
}

func TestExportCountersInvariant(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "export-counters", 3)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("key-counters"))
	if err != nil {
		t.Fatalf("POST /api/v1/exports: %v", err)
	}
	s.RequireStatus(resp, 201, "POST /api/v1/exports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode export: %v", err)
	}
	obj, _ := body.(map[string]any)
	exportID := harness.StringField(obj, "id")
	trackExport(s, exportID)

	get, err := s.API.Get(ctx, "/api/v1/exports/"+exportID, nil)
	if err != nil {
		t.Fatalf("GET export: %v", err)
	}
	s.RequireStatus(get, 200, "GET /api/v1/exports/{id}")
	s.RequireSchema("GET", "/api/v1/exports/{id}", get)

	decoded, err := get.JSON()
	if err != nil {
		t.Fatalf("decode export: %v", err)
	}
	job, _ := decoded.(map[string]any)
	assertExportCountsSuppressed(t, job, "total_in_segment", "no_contact_count", "blocked_by_limit_count", "control_count", "exported_count")
	var total, noContact, blocked, control, exported int
	if err := s.DB.QueryRowContext(ctx, `SELECT total_in_segment, no_contact_count, blocked_by_limit_count, control_count, exported_count FROM export_job WHERE id = $1`, exportID).Scan(&total, &noContact, &blocked, &control, &exported); err != nil {
		t.Fatalf("read export counts: %v", err)
	}

	if total != 3 {
		t.Fatalf("total_in_segment: expected 3, got %d", total)
	}
	if total-noContact-blocked != control+exported {
		t.Fatalf("counter invariant violated: total(%d) - noContact(%d) - blocked(%d) != control(%d) + exported(%d)",
			total, noContact, blocked, control, exported)
	}
}

func TestExportListJournal(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "export-journal", 1)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("key-journal"))
	if err != nil {
		t.Fatalf("POST /api/v1/exports: %v", err)
	}
	s.RequireStatus(resp, 201, "POST /api/v1/exports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode export: %v", err)
	}
	obj, _ := body.(map[string]any)
	exportID := harness.StringField(obj, "id")
	trackExport(s, exportID)

	list, err := s.API.Get(ctx, "/api/v1/exports", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/exports: %v", err)
	}
	s.RequireStatus(list, 200, "GET /api/v1/exports")
	s.RequireSchema("GET", "/api/v1/exports", list)

	decoded, err := list.JSON()
	if err != nil {
		t.Fatalf("decode export list: %v", err)
	}
	items, _ := harness.Items(decoded)
	found := false
	for _, item := range items {
		o, _ := item.(map[string]any)
		if harness.StringField(o, "id") == exportID {
			found = true
		}
	}
	if !found {
		t.Fatalf("export journal must contain export %s", exportID)
	}
}

func TestExportGetUnknownReturns404(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/exports/00000000-0000-0000-0000-000000000000", nil)
	if err != nil {
		t.Fatalf("GET unknown export: %v", err)
	}
	s.RequireStatus(resp, 404, "GET unknown export")
	s.RequireDeclaredStatus("GET", "/api/v1/exports/{id}", resp)
	s.RequireSchema("GET", "/api/v1/exports/{id}", resp)
}

func TestExportPreviewRejectsUnknownChannel(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "export-bad-channel", 1)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := exportPreview(ctx, s.API, ew.segment.ID, "telegram")
	if err != nil {
		t.Fatalf("preview with an unknown channel: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/exports/preview with an unknown channel")
	s.RequireDeclaredStatus("POST", "/api/v1/exports/preview", resp)
	s.RequireSchema("POST", "/api/v1/exports/preview", resp)
}

func TestExportPreviewRejectsUnknownSegment(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := exportPreview(ctx, s.API, "00000000-0000-0000-0000-000000000000", "sms")
	if err != nil {
		t.Fatalf("preview with an unknown segment: %v", err)
	}
	s.RequireStatus(resp, 404, "POST /api/v1/exports/preview with an unknown segment")
	s.RequireDeclaredStatus("POST", "/api/v1/exports/preview", resp)
	s.RequireSchema("POST", "/api/v1/exports/preview", resp)
}

func trackExport(s *harness.Suite, exportID string) {
	s.T.Helper()
	if exportID == "" {
		return
	}
	s.T.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM contact_log WHERE export_id = $1`, exportID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM export_member WHERE export_id = $1`, exportID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM export_job WHERE id = $1`, exportID)
	})
}

func assertExportCountsSuppressed(t *testing.T, obj map[string]any, fields ...string) {
	t.Helper()
	if obj["suppressed"] != true {
		t.Fatalf("expected suppressed response, got %v", obj)
	}
	for _, field := range fields {
		value, exists := obj[field]
		if !exists || value != nil {
			t.Fatalf("%s must be null, got %v (present: %t)", field, value, exists)
		}
	}
}
