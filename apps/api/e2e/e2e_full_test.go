package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestE2EImportToExport(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "e2e-full")
	ctx, cancel := s.Ctx()
	defer cancel()

	phone := s.Reg.Phone(17)
	content := []byte(csvHeader +
		"e2e-full-1,2026-10-01T10:00:00Z,purchase,1000," + phone + ",," + iw.world.tenantA.Code + "\n")

	resp, err := uploadImport(ctx, s, iw.source.ID, "full.csv", content)
	if err != nil {
		t.Fatalf("step 1 POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "step 1: upload import")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("step 1 decode import: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batch_id", "batchId")
	trackImport(s, batchID)

	var guestID string
	s.WaitFor(s.Cfg.ImportTimeout, "step 2 identity resolution to create the guest", func(ctx context.Context) (bool, error) {
		var id string
		err := s.DB.QueryRowContext(ctx, `
			SELECT guest_id FROM guest_identifier WHERE id_type = 'phone' AND value_normalized = $1`, phone).Scan(&id)
		if err != nil {
			return false, err
		}
		guestID = id
		return true, nil
	})

	profile, err := s.API.Get(ctx, "/api/v1/guests/"+guestID, nil)
	if err != nil {
		t.Fatalf("step 3 GET profile: %v", err)
	}
	s.RequireStatus(profile, 200, "step 3: read the resolved profile")
	s.RequireSchema("GET", "/api/v1/guests/{id}", profile)

	events, err := s.API.Get(ctx, "/api/v1/guests/"+guestID+"/events", nil)
	if err != nil {
		t.Fatalf("step 4 GET events: %v", err)
	}
	s.RequireStatus(events, 200, "step 4: read the imported event")
	eventsBody, err := events.JSON()
	if err != nil {
		t.Fatalf("step 4 decode events: %v", err)
	}
	eventItems, _ := harness.Items(eventsBody)
	if len(eventItems) != 1 {
		t.Fatalf("step 4: the imported event must be attached to the resolved guest, got %d", len(eventItems))
	}

	segmentName := s.Reg.Name("e2e-full-segment")
	create, err := s.PostContractShape(ctx, "/api/v1/segments",
		map[string]any{"name": segmentName, "rule": map[string]any{"was": []string{iw.world.tenantA.ID}, "wasNot": []string{iw.world.tenantC.ID}, "periodDays": 365}},
		map[string]any{"name": segmentName, "was": []string{iw.world.tenantA.ID}, "was_not": []string{iw.world.tenantC.ID}, "period_days": 365})
	if err != nil {
		t.Fatalf("step 5 create segment: %v", err)
	}
	s.RequireStatusOrDrift(create, 201, "step 5: create segment")
	createBody, err := create.JSON()
	if err != nil {
		t.Fatalf("step 5 decode segment: %v", err)
	}
	createdObj, _ := createBody.(map[string]any)
	segmentID := harness.StringField(createdObj, "id")
	if segmentID == "" {
		t.Fatalf("step 5: created segment has no id: %s", create.Snippet())
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cleanupCancel()
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM segment_member WHERE segment_version_id IN (SELECT id FROM segment_version WHERE segment_id = $1)`, segmentID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM segment_version WHERE segment_id = $1`, segmentID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM segment WHERE id = $1`, segmentID)
	})

	refresh, err := s.API.Post(ctx, "/api/v1/guests/"+guestID+"/recommendations/refresh", nil)
	if err != nil {
		t.Fatalf("step 6 refresh recommendations: %v", err)
	}
	s.RequireStatus(refresh, 200, "step 6: refresh recommendations")

	marketer := s.Marketer()
	preview, err := exportPreview(ctx, marketer, segmentID, "sms")
	if err != nil {
		t.Fatalf("step 7 preview export: %v", err)
	}
	s.RequireStatus(preview, 200, "step 7: preview export")
	previewBody, err := preview.JSON()
	if err != nil {
		t.Fatalf("step 7 decode preview: %v", err)
	}
	previewObj, _ := previewBody.(map[string]any)
	assertExportCountsSuppressed(t, previewObj, "in_segment", "with_contact", "blocked_by_limit", "control", "to_send")

	exportResp, err := exportCreate(ctx, marketer, segmentID, s.Reg.Name("e2e-full-key"))
	if err != nil {
		t.Fatalf("step 8 create export: %v", err)
	}
	s.RequireStatus(exportResp, 201, "step 8: create export")
	exportBody, err := exportResp.JSON()
	if err != nil {
		t.Fatalf("step 8 decode export: %v", err)
	}
	exportObj, _ := exportBody.(map[string]any)
	exportID := harness.StringField(exportObj, "id")
	trackExport(s, exportID)

	download, err := marketer.Get(ctx, "/api/v1/exports/"+exportID+"/download", nil)
	if err != nil {
		t.Fatalf("step 9 download export: %v", err)
	}
	s.RequireStatus(download, 200, "step 9: download export")
	csvText := string(download.Body)
	if !strings.Contains(csvText, guestID) {
		t.Fatalf("step 9: the downloaded CSV must contain the imported guest %s, got %q", guestID, csvText)
	}
	if !strings.Contains(csvText, phone) {
		t.Fatalf("step 9: the downloaded CSV must contain the phone contact %s, got %q", phone, csvText)
	}
}

func TestE2ENewEventChangesSegmentAndRecommendations(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "e2e-delta")
	ctx, cancel := s.Ctx()
	defer cancel()

	phone := s.Reg.Phone(18)
	first := []byte(csvHeader +
		"e2e-delta-1,2026-10-01T10:00:00Z,purchase,1000," + phone + ",," + iw.world.tenantA.Code + "\n")

	resp, err := uploadImport(ctx, s, iw.source.ID, "delta-1.csv", first)
	if err != nil {
		t.Fatalf("step 1 upload first import: %v", err)
	}
	s.RequireStatus(resp, 202, "step 1: first import")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("step 1 decode import: %v", err)
	}
	obj, _ := body.(map[string]any)
	trackImport(s, harness.StringField(obj, "batch_id", "batchId"))

	var guestID string
	s.WaitFor(s.Cfg.ImportTimeout, "step 2 guest resolution", func(ctx context.Context) (bool, error) {
		var id string
		if err := s.DB.QueryRowContext(ctx, `
			SELECT guest_id FROM guest_identifier WHERE id_type = 'phone' AND value_normalized = $1`, phone).Scan(&id); err != nil {
			return false, err
		}
		guestID = id
		return true, nil
	})

	before, err := s.API.Get(ctx, "/api/v1/guests/"+guestID, nil)
	if err != nil {
		t.Fatalf("step 3 GET profile: %v", err)
	}
	s.RequireStatus(before, 200, "step 3: profile before the second import")
	beforeBody, err := before.JSON()
	if err != nil {
		t.Fatalf("step 3 decode profile: %v", err)
	}
	beforeObj, _ := beforeBody.(map[string]any)
	beforeTenants, _ := harness.IntField(beforeObj, "tenantCount", "tenant_count")

	second := []byte(csvHeader +
		"e2e-delta-2,2026-10-02T10:00:00Z,purchase,2000," + phone + ",," + iw.world.tenantB.Code + "\n")
	resp2, err := uploadImport(ctx, s, iw.source.ID, "delta-2.csv", second)
	if err != nil {
		t.Fatalf("step 4 upload second import: %v", err)
	}
	s.RequireStatus(resp2, 202, "step 4: second import")
	body2, err := resp2.JSON()
	if err != nil {
		t.Fatalf("step 4 decode import: %v", err)
	}
	obj2, _ := body2.(map[string]any)
	batch2 := harness.StringField(obj2, "batch_id", "batchId")
	trackImport(s, batch2)

	s.WaitFor(s.Cfg.ImportTimeout, "step 5 second event to attach", func(ctx context.Context) (bool, error) {
		n, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM event WHERE import_batch_id = $1 AND guest_id = $2`, batch2, guestID)
		if err != nil {
			return false, err
		}
		return n == 1, nil
	})

	after, err := s.API.Get(ctx, "/api/v1/guests/"+guestID, nil)
	if err != nil {
		t.Fatalf("step 6 GET profile: %v", err)
	}
	s.RequireStatus(after, 200, "step 6: profile after the second import")
	afterBody, err := after.JSON()
	if err != nil {
		t.Fatalf("step 6 decode profile: %v", err)
	}
	afterObj, _ := afterBody.(map[string]any)
	afterTenants, _ := harness.IntField(afterObj, "tenantCount", "tenant_count")
	if afterTenants <= beforeTenants {
		t.Fatalf("step 6: a purchase in a new tenant must grow tenantCount: before=%d after=%d", beforeTenants, afterTenants)
	}

	events, err := s.API.Get(ctx, "/api/v1/guests/"+guestID+"/events", nil)
	if err != nil {
		t.Fatalf("step 7 GET events: %v", err)
	}
	s.RequireStatus(events, 200, "step 7: event feed")
	eventsBody, err := events.JSON()
	if err != nil {
		t.Fatalf("step 7 decode events: %v", err)
	}
	eventItems, _ := harness.Items(eventsBody)
	if len(eventItems) != 2 {
		t.Fatalf("step 7: both imports must be visible on the profile, got %d events", len(eventItems))
	}

	refresh, err := s.API.Post(ctx, "/api/v1/guests/"+guestID+"/recommendations/refresh", nil)
	if err != nil {
		t.Fatalf("step 8 refresh recommendations: %v", err)
	}
	s.RequireStatus(refresh, 200, "step 8: refresh recommendations")
	refreshBody, err := refresh.JSON()
	if err != nil {
		t.Fatalf("step 8 decode recommendations: %v", err)
	}
	items, _ := harness.Items(refreshBody)
	for _, item := range items {
		reco, _ := item.(map[string]any)
		tenantID := harness.StringField(reco, "tenantId", "tenant_id")
		if tenantID == iw.world.tenantA.ID || tenantID == iw.world.tenantB.ID {
			t.Fatalf("step 8: recommendations must exclude visited tenants, got %s", tenantID)
		}
	}
}
