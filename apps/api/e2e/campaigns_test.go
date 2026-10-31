package e2e

import (
	"context"
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestCampaignCreateRequiresTenantAndSegment(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	w := buildWorld(t, s, "campaign-create")
	ctx, cancel := s.Ctx()
	defer cancel()

	segment := mustSegment(t, s, w, "campaign-create")

	resp, err := s.API.Post(ctx, "/api/v1/campaigns", map[string]any{
		"name":             s.Reg.Name("campaign"),
		"tenant_id":        w.tenantC.ID,
		"segment_id":       segment.ID,
		"attribution_days": 14,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/campaigns: %v", err)
	}
	s.RequireStatusOrDrift(resp, 201, "POST /api/v1/campaigns")
	s.RequireDeclaredStatus("POST", "/api/v1/campaigns", resp)
	s.RequireSchema("POST", "/api/v1/campaigns", resp)

	created := jsonObject(t, resp, "created campaign")
	if got := harness.StringField(created, "tenant_id", "target_tenant_id"); got != w.tenantC.ID {
		t.Fatalf("campaign target tenant: expected %s, got %q", w.tenantC.ID, got)
	}
	if days, _ := harness.IntField(created, "attribution_days"); days != 14 {
		t.Fatalf("attribution_days: expected 14, got %d", days)
	}
	trackCampaign(s, harness.StringField(created, "id"))
}

func TestCampaignCreateDefaultsAttributionWindow(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	w := buildWorld(t, s, "campaign-default")
	ctx, cancel := s.Ctx()
	defer cancel()

	segment := mustSegment(t, s, w, "campaign-default")
	resp, err := s.API.Post(ctx, "/api/v1/campaigns", map[string]any{
		"name":       s.Reg.Name("campaign-default"),
		"tenant_id":  w.tenantC.ID,
		"segment_id": segment.ID,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/campaigns: %v", err)
	}
	s.RequireStatusOrDrift(resp, 201, "POST /api/v1/campaigns without attribution_days")
	if days, _ := harness.IntField(jsonObject(t, resp, "created campaign"), "attribution_days"); days != 14 {
		t.Fatalf("the default attribution window must be 14 days, got %d", days)
	}
	trackCampaign(s, harness.StringField(jsonObject(t, resp, "created campaign"), "id"))
}

func TestCampaignDeleteWithExportsReturns409(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	w := buildWorld(t, s, "campaign-delete")
	ctx, cancel := s.Ctx()
	defer cancel()

	segment := mustSegment(t, s, w, "campaign-delete")
	create, err := s.API.Post(ctx, "/api/v1/campaigns", map[string]any{
		"name":       s.Reg.Name("campaign-delete"),
		"tenant_id":  w.tenantC.ID,
		"segment_id": segment.ID,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/campaigns: %v", err)
	}
	s.RequireStatusOrDrift(create, 201, "POST /api/v1/campaigns")
	campaignID := harness.StringField(jsonObject(t, create, "created campaign"), "id")
	trackCampaign(s, campaignID)

	export, err := s.API.Post(ctx, "/api/v1/exports", map[string]any{
		"segment_id":      segment.ID,
		"channel":         "sms",
		"format":          "csv",
		"campaign_id":     campaignID,
		"idempotency_key": s.Reg.Name("campaign-export"),
	})
	if err != nil {
		t.Fatalf("POST /api/v1/exports: %v", err)
	}
	s.RequireStatusOrDrift(export, 201, "POST /api/v1/exports for the campaign")
	trackExport(s, harness.StringField(jsonObject(t, export, "created export"), "id"))

	resp, err := s.API.Delete(ctx, "/api/v1/campaigns/"+campaignID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/campaigns/{id}: %v", err)
	}
	s.RequireStatus(resp, 409, "DELETE /api/v1/campaigns/{id} that owns exports")
	s.RequireDeclaredStatus("DELETE", "/api/v1/campaigns/{id}", resp)
}

func TestCampaignDeleteWithoutExports(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	w := buildWorld(t, s, "campaign-clean")
	ctx, cancel := s.Ctx()
	defer cancel()

	segment := mustSegment(t, s, w, "campaign-clean")
	create, err := s.API.Post(ctx, "/api/v1/campaigns", map[string]any{
		"name":       s.Reg.Name("campaign-clean"),
		"tenant_id":  w.tenantC.ID,
		"segment_id": segment.ID,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/campaigns: %v", err)
	}
	s.RequireStatusOrDrift(create, 201, "POST /api/v1/campaigns")
	campaignID := harness.StringField(jsonObject(t, create, "created campaign"), "id")

	resp, err := s.API.Delete(ctx, "/api/v1/campaigns/"+campaignID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/campaigns/{id}: %v", err)
	}
	s.RequireStatus(resp, 204, "DELETE /api/v1/campaigns/{id} without exports")
	s.RequireDeclaredStatus("DELETE", "/api/v1/campaigns/{id}", resp)
}

func TestCampaignCreateForbiddenForManagerAndAnalyst(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "campaign-roles")
	ctx, cancel := s.Ctx()
	defer cancel()

	segment := mustSegment(t, s, w, "campaign-roles")
	for _, role := range []string{"manager", "analyst"} {
		c := actor(t, s, role)
		resp, err := c.Post(ctx, "/api/v1/campaigns", map[string]any{
			"name":       s.Reg.Name("campaign-" + role),
			"tenant_id":  w.tenantC.ID,
			"segment_id": segment.ID,
		})
		if err != nil {
			t.Fatalf("POST /api/v1/campaigns as %s: %v", role, err)
		}
		s.RequireStatus(resp, 403, "POST /api/v1/campaigns as "+role)
	}
}

func mustSegment(t *testing.T, s *harness.Suite, w *world, suffix string) *segmentRef {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
	defer cancel()

	members := make([]string, 0, 2)
	for i := 0; i < 2; i++ {
		guest, err := s.Reg.Guest(ctx, suffix+"-g"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("fixture guest %d: %v", i, err)
		}
		if _, err := s.Reg.Identifier(ctx, guest.ID, "phone", s.Reg.Phone(30+i), w.source.ID); err != nil {
			t.Fatalf("fixture identifier %d: %v", i, err)
		}
		members = append(members, guest.ID)
	}
	created, err := s.Reg.Segment(ctx, suffix, `{"visitedTenantIds":[],"excluded_tenant_ids":[],"period_days":90}`, members)
	if err != nil {
		t.Fatalf("fixture segment: %v", err)
	}
	return &segmentRef{ID: created.ID}
}

func trackCampaign(s *harness.Suite, campaignID string) {
	s.T.Helper()
	if campaignID == "" {
		return
	}
	s.T.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cancel()
		_ = s.DB.Exec(ctx, `DELETE FROM campaign WHERE id = $1`, campaignID)
	})
}

type segmentRef struct {
	ID   string
	Name string
}
