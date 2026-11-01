package e2e

import (
	"context"
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestSegmentTemplatesFive(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/segment-templates", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/segment-templates: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/segment-templates")
	s.RequireDeclaredStatus("GET", "/api/v1/segment-templates", resp)
	s.RequireSchema("GET", "/api/v1/segment-templates", resp)

	items, ok := harness.Items(mustJSON(t, resp, "segment templates"))
	if !ok {
		t.Fatalf("the template payload must expose items: %s", resp.Snippet())
	}
	if len(items) != 5 {
		t.Fatalf("exactly five seeded templates are expected, got %d", len(items))
	}
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if code := harness.StringField(item, "code"); code == "" {
			t.Fatalf("every template must carry a code: %v", item)
		}
		if name := harness.StringField(item, "name"); name == "" {
			t.Fatalf("every template must carry a name: %v", item)
		}
		if rule, ok := item["rule"]; !ok || rule == nil {
			t.Fatalf("every template must carry a rule: %v", item)
		}
	}
}

func TestSegmentUpdateCreatesVersion(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "segment-update")
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "segment-update")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "segment-update", 5)
	segment := createSegmentViaAPI(t, s, w.tenantA.ID)

	update, err := s.API.Put(ctx, "/api/v1/segments/"+segment.ID, map[string]any{
		"name": segment.Name,
		"rule": map[string]any{"visitedTenantIds": []string{w.tenantB.ID}, "excludedTenantIds": []string{}, "periodDays": 90},
	})
	if err != nil {
		t.Fatalf("PUT /api/v1/segments/{id}: %v", err)
	}
	s.RequireStatus(update, 200, "PUT /api/v1/segments/{id}")
	s.RequireDeclaredStatus("PUT", "/api/v1/segments/{id}", update)
	s.RequireSchema("PUT", "/api/v1/segments/{id}", update)

	updated := jsonObject(t, update, "updated segment")
	if version, _ := harness.IntField(updated, "currentVersionNo"); version < 2 {
		t.Fatalf("editing the rule must bump the version, got %d", version)
	}

	versions, err := s.API.Get(ctx, "/api/v1/segments/"+segment.ID+"/versions", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/segments/{id}/versions: %v", err)
	}
	s.RequireStatus(versions, 200, "GET /api/v1/segments/{id}/versions")
	s.RequireDeclaredStatus("GET", "/api/v1/segments/{id}/versions", versions)
	s.RequireSchema("GET", "/api/v1/segments/{id}/versions", versions)

	items, _ := harness.Items(mustJSON(t, versions, "segment versions"))
	if len(items) < 2 {
		t.Fatalf("the version history must keep both revisions, got %d", len(items))
	}
}

func TestSegmentMembersPaged(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "segment-members")
	ctx, cancel := s.Ctx()
	defer cancel()

	guests := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		guest, err := s.Reg.Guest(ctx, "segment-members-"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("fixture guest %d: %v", i, err)
		}
		w.visit(t, s, guest.ID, w.tenantA.ID, "segment-members-"+string(rune('a'+i)), 4)
		guests = append(guests, guest.ID)
	}
	segment := createSegmentViaAPI(t, s, w.tenantA.ID)

	resp, err := s.API.Get(ctx, "/api/v1/segments/"+segment.ID+"/members", map[string][]string{"limit": {"2"}})
	if err != nil {
		t.Fatalf("GET /api/v1/segments/{id}/members: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/segments/{id}/members")
	s.RequireDeclaredStatus("GET", "/api/v1/segments/{id}/members", resp)
	s.RequireSchema("GET", "/api/v1/segments/{id}/members", resp)

	items, ok := harness.Items(mustJSON(t, resp, "segment members"))
	if !ok {
		t.Fatalf("the members payload must expose items: %s", resp.Snippet())
	}
	if len(items) != 2 {
		t.Fatalf("limit=2 must return at most two members, got %d", len(items))
	}
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("member %v is not an object", item)
		}
		if !containsString(guests, harness.StringField(entry, "guestId")) {
			t.Fatalf("member %v is not one of the fixture guests", entry)
		}
		if tenants, _ := harness.IntField(entry, "tenantCount"); tenants != 1 {
			t.Fatalf("member %v must count the one visited tenant", entry)
		}
		if harness.StringField(entry, "lastSeenAt") == "" || harness.StringField(entry, "addedAt") == "" {
			t.Fatalf("member %v must carry lastSeenAt and addedAt", entry)
		}
		if harness.StringField(entry, "displayName") == "" {
			t.Fatalf("an analyst with field.guest.display_name.read must get the guest name: %v", entry)
		}
	}
}

func TestSegmentArchiveRemovesItFromActiveList(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "segment-archive")
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "segment-archive")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "segment-archive", 4)
	segment := createSegmentViaAPI(t, s, w.tenantA.ID)

	resp, err := s.API.Delete(ctx, "/api/v1/segments/"+segment.ID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/segments/{id}: %v", err)
	}
	s.RequireStatus(resp, 204, "DELETE /api/v1/segments/{id}")
	s.RequireDeclaredStatus("DELETE", "/api/v1/segments/{id}", resp)

	status, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM segment WHERE id = $1 AND status = 'archived'`, segment.ID)
	if err != nil {
		t.Fatalf("read the archived segment: %v", err)
	}
	if status != 1 {
		t.Fatalf("the segment must be archived, not removed")
	}

	update, err := s.API.Put(ctx, "/api/v1/segments/"+segment.ID, map[string]any{
		"name": segment.Name,
		"rule": map[string]any{"visitedTenantIds": []string{w.tenantA.ID}, "excludedTenantIds": []string{}, "periodDays": 30},
	})
	if err != nil {
		t.Fatalf("PUT /api/v1/segments/{id} after archiving: %v", err)
	}
	s.RequireStatus(update, 409, "PUT /api/v1/segments/{id} on an archived segment")
	s.RequireDeclaredStatus("PUT", "/api/v1/segments/{id}", update)
}

func TestSegmentRecomputeAcceptsAndQueues(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "segment-recompute")
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "segment-recompute")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "segment-recompute", 4)
	segment := createSegmentViaAPI(t, s, w.tenantA.ID)

	resp, err := s.API.Post(ctx, "/api/v1/segments/"+segment.ID+"/recompute", nil)
	if err != nil {
		t.Fatalf("POST /api/v1/segments/{id}/recompute: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/segments/{id}/recompute")
	s.RequireDeclaredStatus("POST", "/api/v1/segments/{id}/recompute", resp)
	s.RequireSchema("POST", "/api/v1/segments/{id}/recompute", resp)

	queued, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM job_queue WHERE job_type = 'segment.recompute' AND payload LIKE '%' || $1 || '%'`, segment.ID)
	if err != nil {
		t.Fatalf("inspect the job queue: %v", err)
	}
	if queued < 1 {
		t.Fatalf("recompute must enqueue a segment.recompute job for %s", segment.ID)
	}
}

func TestSegmentRecommendationsAggregate(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "segment-reco")
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "segment-reco")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "segment-reco", 4)
	segment := createSegmentViaAPI(t, s, w.tenantA.ID)

	resp, err := s.API.Get(ctx, "/api/v1/segments/"+segment.ID+"/recommendations", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/segments/{id}/recommendations: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/segments/{id}/recommendations")
	s.RequireDeclaredStatus("GET", "/api/v1/segments/{id}/recommendations", resp)
	s.RequireSchema("GET", "/api/v1/segments/{id}/recommendations", resp)
}

func createSegmentViaAPI(t *testing.T, s *harness.Suite, tenantID string) *segmentRef {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
	defer cancel()

	name := s.Reg.Name("api-segment")
	resp, err := s.API.Post(ctx, "/api/v1/segments", map[string]any{
		"name": name,
		"rule": map[string]any{"visitedTenantIds": []string{tenantID}, "excludedTenantIds": []string{}, "periodDays": 90},
	})
	if err != nil {
		t.Fatalf("POST /api/v1/segments: %v", err)
	}
	s.RequireStatus(resp, 201, "POST /api/v1/segments")
	created := jsonObject(t, resp, "created segment")
	id := harness.StringField(created, "id")
	if id == "" {
		t.Fatalf("created segment has no id: %s", resp.Snippet())
	}
	trackSegment(s, id)
	return &segmentRef{ID: id, Name: name}
}

func trackSegment(s *harness.Suite, segmentID string) {
	s.T.Helper()
	if segmentID == "" {
		return
	}
	s.T.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cancel()
		_ = s.DB.Exec(ctx, `DELETE FROM segment_member WHERE segment_version_id IN (SELECT id FROM segment_version WHERE segment_id = $1)`, segmentID)
		_ = s.DB.Exec(ctx, `DELETE FROM segment_version WHERE segment_id = $1`, segmentID)
		_ = s.DB.Exec(ctx, `DELETE FROM segment WHERE id = $1`, segmentID)
	})
}

func containsString(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}
