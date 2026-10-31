package e2e

import (
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestContactLimitBlocksRepeatExport(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "limit-blocks", 3)
	ctx, cancel := s.Ctx()
	defer cancel()

	if _, err := s.Reg.ContactLimit(ctx, "sms", 1, 7); err != nil {
		t.Fatalf("fixture contact limit: %v", err)
	}

	first, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("limit-key-1"))
	if err != nil {
		t.Fatalf("first export: %v", err)
	}
	s.RequireStatus(first, 201, "first export")
	firstBody, err := first.JSON()
	if err != nil {
		t.Fatalf("decode first export: %v", err)
	}
	firstObj, _ := firstBody.(map[string]any)
	trackExport(s, harness.StringField(firstObj, "id"))

	preview, err := exportPreview(ctx, s.API, ew.segment.ID, "sms")
	if err != nil {
		t.Fatalf("preview after export: %v", err)
	}
	s.RequireStatus(preview, 200, "preview after the first export")
	previewBody, err := preview.JSON()
	if err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	previewObj, _ := previewBody.(map[string]any)
	assertExportCountsSuppressed(t, previewObj, "in_segment", "with_contact", "blocked_by_limit", "control", "to_send")
	second, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("limit-key-2"))
	if err != nil {
		t.Fatalf("second export: %v", err)
	}
	s.RequireStatus(second, 409, "repeat export blocked by contact limit")
}

func TestContactLimitStrictestRuleWins(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "limit-strictest", 2)
	ctx, cancel := s.Ctx()
	defer cancel()

	if _, err := s.Reg.ContactLimit(ctx, "", 5, 30); err != nil {
		t.Fatalf("fixture broad limit: %v", err)
	}
	if _, err := s.Reg.ContactLimit(ctx, "sms", 1, 7); err != nil {
		t.Fatalf("fixture strict limit: %v", err)
	}

	first, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("strict-key-1"))
	if err != nil {
		t.Fatalf("first export: %v", err)
	}
	s.RequireStatus(first, 201, "first export")
	firstBody, err := first.JSON()
	if err != nil {
		t.Fatalf("decode first export: %v", err)
	}
	firstObj, _ := firstBody.(map[string]any)
	trackExport(s, harness.StringField(firstObj, "id"))

	preview, err := exportPreview(ctx, s.API, ew.segment.ID, "sms")
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	s.RequireStatus(preview, 200, "preview after the first export")
	previewBody, err := preview.JSON()
	if err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	previewObj, _ := previewBody.(map[string]any)
	assertExportCountsSuppressed(t, previewObj, "in_segment", "with_contact", "blocked_by_limit", "control", "to_send")
	second, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("strict-key-2"))
	if err != nil {
		t.Fatalf("second export: %v", err)
	}
	s.RequireStatus(second, 409, "strictest contact limit blocks repeat export")
}

func TestContactLimitIsChannelScoped(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "limit-channel", 2)
	ctx, cancel := s.Ctx()
	defer cancel()

	if _, err := s.Reg.ContactLimit(ctx, "sms", 1, 7); err != nil {
		t.Fatalf("fixture sms limit: %v", err)
	}

	first, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("channel-key-1"))
	if err != nil {
		t.Fatalf("first export: %v", err)
	}
	s.RequireStatus(first, 201, "first export")
	firstBody, err := first.JSON()
	if err != nil {
		t.Fatalf("decode first export: %v", err)
	}
	firstObj, _ := firstBody.(map[string]any)
	trackExport(s, harness.StringField(firstObj, "id"))

	preview, err := exportPreview(ctx, s.API, ew.segment.ID, "sms")
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	s.RequireStatus(preview, 200, "preview")
	previewBody, err := preview.JSON()
	if err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	previewObj, _ := previewBody.(map[string]any)
	assertExportCountsSuppressed(t, previewObj, "in_segment", "with_contact", "blocked_by_limit", "control", "to_send")
	second, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("channel-key-2"))
	if err != nil {
		t.Fatalf("second export: %v", err)
	}
	s.RequireStatus(second, 409, "sms limit blocks repeat sms export")
}

func TestContactLimitControlGroupNotLogged(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "limit-control", 4)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/exports", map[string]any{
		"segment_id":      ew.segment.ID,
		"channel":         "sms",
		"format":          "csv",
		"control_percent": 50,
		"idempotency_key": s.Reg.Name("control-log-key"),
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

	logs, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM contact_log WHERE export_id = $1`, exportID)
	if err != nil {
		t.Fatalf("count contact logs: %v", err)
	}
	if logs != exported {
		t.Fatalf("contact_log must contain only recipients: expected %d, got %d", exported, logs)
	}
	if control == 0 {
		t.Fatalf("control_count: expected a non-empty control group at 50%%, got 0")
	}
}
