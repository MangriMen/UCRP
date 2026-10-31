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
	blocked, _ := harness.IntField(previewObj, "blockedByLimit")
	toSend, _ := harness.IntField(previewObj, "toSend")

	if blocked != 3 {
		t.Fatalf("blocked_by_limit: expected all 3 guests blocked by maxContacts=1, got %d", blocked)
	}
	if toSend != 0 {
		t.Fatalf("to_send: expected 0 when every guest is blocked, got %d", toSend)
	}
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
	blocked, _ := harness.IntField(previewObj, "blockedByLimit")
	if blocked != 2 {
		t.Fatalf("the strictest rule (maxContacts=1) must block both guests, got %d blocked", blocked)
	}
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
	blocked, _ := harness.IntField(previewObj, "blockedByLimit")
	if blocked != 2 {
		t.Fatalf("sms limit must block both guests on the sms channel, got %d", blocked)
	}
}

func TestContactLimitControlGroupNotLogged(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "limit-control", 4)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/exports", map[string]any{
		"segmentId":      ew.segment.ID,
		"channel":        "sms",
		"format":         "csv",
		"controlPercent": 50,
		"idempotencyKey": s.Reg.Name("control-log-key"),
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

	control, _ := harness.IntField(obj, "controlCount")
	exported, _ := harness.IntField(obj, "exportedCount")

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
