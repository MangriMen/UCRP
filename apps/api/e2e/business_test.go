package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/harness"
	"github.com/MangriMen/UCRP/e2e/report"
)

func TestGuestProfileReflectsTenantCount(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "profile-tenants")

	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "profile-tenants")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "profile-a", 3)
	w.visit(t, s, guest.ID, w.tenantB.ID, "profile-b", 2)

	resp, err := s.API.Get(ctx, "/api/v1/guests/"+guest.ID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{id}: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests/{id}")
	s.RequireSchema("GET", "/api/v1/guests/{id}", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	profile, ok := body.(map[string]any)
	if !ok {
		t.Fatalf("profile is not a JSON object: %s", resp.Snippet())
	}
	count, ok := harness.IntField(profile, "tenantCount")
	if !ok {
		t.Fatalf("profile has no tenantCount field: %s", resp.Snippet())
	}
	if count != 2 {
		t.Fatalf("tenantCount: expected 2 distinct tenants with presence events, got %d", count)
	}
}

func TestGuestEventsSortedDescending(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "events-order")

	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "events-order")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "events-old", 10)
	w.visit(t, s, guest.ID, w.tenantB.ID, "events-new", 1)

	resp, err := s.API.Get(ctx, "/api/v1/guests/"+guest.ID+"/events", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{id}/events: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests/{id}/events")
	s.RequireSchema("GET", "/api/v1/guests/{id}/events", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode events: %v", err)
	}
	items, ok := harness.Items(body)
	if !ok {
		t.Fatalf("event feed has no items array: %s", resp.Snippet())
	}
	if len(items) != 2 {
		t.Fatalf("event feed: expected 2 events, got %d", len(items))
	}
	first, _ := items[0].(map[string]any)
	second, _ := items[1].(map[string]any)
	firstAt := harness.StringField(first, "occurredAt")
	secondAt := harness.StringField(second, "occurredAt")
	if firstAt == "" || secondAt == "" {
		t.Fatalf("event feed items carry no occurredAt: %s", resp.Snippet())
	}
	if firstAt < secondAt {
		t.Fatalf("event feed must be sorted by occurredAt descending: got %s before %s", firstAt, secondAt)
	}
}

func TestGuestSearchFindsByIdentifier(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "search-identifier")

	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "search-identifier")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	phone := s.Reg.Phone(11)
	if _, err := s.Reg.Identifier(ctx, guest.ID, "phone", phone, w.source.ID); err != nil {
		t.Fatalf("fixture identifier: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/guests", map[string][]string{"q": {phone}})
	if err != nil {
		t.Fatalf("GET /api/v1/guests?q=%s: %v", phone, err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests by phone")
	s.RequireSchema("GET", "/api/v1/guests", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode search result: %v", err)
	}
	items, ok := harness.Items(body)
	if !ok {
		t.Fatalf("search result has no items array: %s", resp.Snippet())
	}
	found := false
	for _, item := range items {
		obj, _ := item.(map[string]any)
		if harness.StringField(obj, "id") == guest.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("search by normalized phone %q did not return guest %s: %s", phone, guest.ID, resp.Snippet())
	}
}

func TestGuestSearchExcludesMergedProfiles(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "search-merged")

	ctx, cancel := s.Ctx()
	defer cancel()

	winner, err := s.Reg.Guest(ctx, "search-merged-winner")
	if err != nil {
		t.Fatalf("fixture winner: %v", err)
	}
	loser, err := s.Reg.Guest(ctx, "search-merged-loser")
	if err != nil {
		t.Fatalf("fixture loser: %v", err)
	}
	phone := s.Reg.Phone(12)
	if _, err := s.Reg.Identifier(ctx, loser.ID, "phone", phone, w.source.ID); err != nil {
		t.Fatalf("fixture identifier: %v", err)
	}
	if err := s.DB.Exec(ctx, `UPDATE guest SET status = 'merged', merged_into_id = $2 WHERE id = $1`, loser.ID, winner.ID); err != nil {
		t.Fatalf("mark loser merged: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/guests", map[string][]string{"q": {phone}})
	if err != nil {
		t.Fatalf("GET /api/v1/guests?q=%s: %v", phone, err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests by phone of a merged profile")
	s.RequireSchema("GET", "/api/v1/guests", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode search result: %v", err)
	}
	items, _ := harness.Items(body)
	for _, item := range items {
		obj, _ := item.(map[string]any)
		if harness.StringField(obj, "id") == loser.ID {
			report.Drifts.Record(t.Name(), "GET /api/v1/guests",
				"search returns a profile with status=merged; the contract excludes merged profiles from search results")
			t.Fatalf("search returned the merged profile %s; the contract excludes status=merged", loser.ID)
		}
	}
}

func TestSegmentPreviewCountsMatchingGuests(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "preview-count")

	ctx, cancel := s.Ctx()
	defer cancel()

	both, err := s.Reg.Guest(ctx, "preview-both")
	if err != nil {
		t.Fatalf("fixture guest both: %v", err)
	}
	onlyA, err := s.Reg.Guest(ctx, "preview-only-a")
	if err != nil {
		t.Fatalf("fixture guest only A: %v", err)
	}
	onlyC, err := s.Reg.Guest(ctx, "preview-only-c")
	if err != nil {
		t.Fatalf("fixture guest only C: %v", err)
	}
	extraA := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		g, err := s.Reg.Guest(ctx, "preview-extra-a-"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("fixture extra guest A %d: %v", i, err)
		}
		extraA = append(extraA, g.ID)
	}
	w.visit(t, s, both.ID, w.tenantA.ID, "preview-both-a", 5)
	w.visit(t, s, both.ID, w.tenantB.ID, "preview-both-b", 4)
	w.visit(t, s, onlyA.ID, w.tenantA.ID, "preview-only-a-a", 3)
	w.visit(t, s, onlyC.ID, w.tenantC.ID, "preview-only-c-c", 2)
	for i, id := range extraA {
		w.visit(t, s, id, w.tenantA.ID, "preview-extra-a-"+string(rune('a'+i)), 3)
	}

	contractBody := map[string]any{
		"rule": map[string]any{
			"visitedTenantIds":  []string{w.tenantA.ID, w.tenantB.ID},
			"excludedTenantIds": []string{w.tenantC.ID},
			"periodDays":        90,
		},
	}
	implementationBody := map[string]any{
		"visitedTenantIds":  []string{w.tenantA.ID, w.tenantB.ID},
		"excludedTenantIds": []string{w.tenantC.ID},
		"periodDays":        90,
	}
	resp, err := s.PostContractShape(ctx, "/api/v1/segments/preview", contractBody, implementationBody)
	if err != nil {
		t.Fatalf("POST /api/v1/segments/preview: %v", err)
	}
	s.RequireStatus(resp, 200, "POST /api/v1/segments/preview")
	s.RequireSchema("POST", "/api/v1/segments/preview", resp)

	decoded, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	obj, _ := decoded.(map[string]any)
	size, ok := harness.IntField(obj, "size")
	if !ok {
		t.Fatalf("preview response has no size field: %s", resp.Snippet())
	}
	if size != 5 {
		t.Fatalf("preview size: expected 5 guests that visited A or B and never C, got %d", size)
	}
}

func TestSegmentPreviewDoesNotPersist(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "preview-nowrite")

	ctx, cancel := s.Ctx()
	defer cancel()

	before, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM segment`)
	if err != nil {
		t.Fatalf("count segments before preview: %v", err)
	}

	contractBody := map[string]any{
		"rule": map[string]any{
			"visitedTenantIds":  []string{w.tenantA.ID},
			"excludedTenantIds": []string{w.tenantC.ID},
			"periodDays":        30,
		},
	}
	implementationBody := map[string]any{
		"visitedTenantIds":  []string{w.tenantA.ID},
		"excludedTenantIds": []string{w.tenantC.ID},
		"periodDays":        30,
	}
	resp, err := s.PostContractShape(ctx, "/api/v1/segments/preview", contractBody, implementationBody)
	if err != nil {
		t.Fatalf("POST /api/v1/segments/preview: %v", err)
	}
	s.RequireStatusOrDrift(resp, 200, "POST /api/v1/segments/preview")

	after, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM segment`)
	if err != nil {
		t.Fatalf("count segments after preview: %v", err)
	}
	if after != before {
		t.Fatalf("preview must not persist a segment: segment count went from %d to %d", before, after)
	}
}

func TestSegmentCreatePersistsMembership(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "create-members")

	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "create-members")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "create-members-a", 1)

	contractBody := map[string]any{
		"name": s.Reg.Name("segment-members"),
		"rule": map[string]any{
			"visitedTenantIds":  []string{w.tenantA.ID},
			"excludedTenantIds": []string{w.tenantC.ID},
			"periodDays":        90,
		},
	}
	implementationBody := map[string]any{
		"name":              s.Reg.Name("segment-members"),
		"visitedTenantIds":  []string{w.tenantA.ID},
		"excludedTenantIds": []string{w.tenantC.ID},
		"periodDays":        90,
	}
	resp, err := s.PostContractShape(ctx, "/api/v1/segments", contractBody, implementationBody)
	if err != nil {
		t.Fatalf("POST /api/v1/segments: %v", err)
	}
	s.RequireStatusOrDrift(resp, 201, "POST /api/v1/segments")
	s.RequireSchema("POST", "/api/v1/segments", resp)

	decoded, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode created segment: %v", err)
	}
	created, _ := decoded.(map[string]any)
	segmentID := harness.StringField(created, "id")
	if segmentID == "" {
		t.Fatalf("created segment has no id: %s", resp.Snippet())
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cleanupCancel()
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM segment_member WHERE segment_version_id IN (SELECT id FROM segment_version WHERE segment_id = $1)`, segmentID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM segment_version WHERE segment_id = $1`, segmentID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM segment WHERE id = $1`, segmentID)
	})

	members, err := s.DB.Count(ctx, `
		SELECT COUNT(*) FROM segment_member sm
		JOIN segment_version sv ON sv.id = sm.segment_version_id
		WHERE sv.segment_id = $1 AND sm.guest_id = $2`, segmentID, guest.ID)
	if err != nil {
		t.Fatalf("count members: %v", err)
	}
	if members != 1 {
		t.Fatalf("created segment %s must materialize the fixture guest %s, got %d matching members", segmentID, guest.ID, members)
	}

	versionStatus, err := s.DB.Count(ctx, `
		SELECT COUNT(*) FROM segment_version
		WHERE segment_id = $1 AND version_no = 1 AND status = 'ready' AND computed_at IS NOT NULL`, segmentID)
	if err != nil {
		t.Fatalf("read version status: %v", err)
	}
	if versionStatus != 1 {
		t.Fatalf("created segment %s must have version 1 in status ready with computed_at set", segmentID)
	}
}

func TestSegmentGetReportsReadyVersion(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "segment-ready")

	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "segment-ready")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "segment-ready-a", 1)
	memberIDs := []string{guest.ID}
	for i := 0; i < 4; i++ {
		extra, err := s.Reg.Guest(ctx, "segment-ready-"+string(rune('b'+i)))
		if err != nil {
			t.Fatalf("fixture extra guest %d: %v", i, err)
		}
		w.visit(t, s, extra.ID, w.tenantA.ID, "segment-ready-extra-"+string(rune('b'+i)), 1)
		memberIDs = append(memberIDs, extra.ID)
	}

	segment, err := s.Reg.Segment(ctx, "ready", `{"visitedTenantIds":[],"excludedTenantIds":[],"periodDays":90}`, memberIDs)
	if err != nil {
		t.Fatalf("fixture segment: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/segments/"+segment.ID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/segments/{id}: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/segments/{id}")
	s.RequireSchema("GET", "/api/v1/segments/{id}", resp)

	decoded, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode segment: %v", err)
	}
	obj, _ := decoded.(map[string]any)
	count, ok := harness.IntField(obj, "memberCount")
	if !ok {
		t.Fatalf("segment response has no memberCount: %s", resp.Snippet())
	}
	if count != 5 {
		t.Fatalf("memberCount: expected 5 materialized members, got %d", count)
	}
}

func TestRecommendationsExcludeVisitedTenants(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "reco-visited")

	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "reco-visited")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "reco-visited-a", 2)

	resp, err := s.API.Post(ctx, "/api/v1/guests/"+guest.ID+"/recommendations/refresh", nil)
	if err != nil {
		t.Fatalf("POST refresh recommendations: %v", err)
	}
	s.RequireStatus(resp, 200, "POST /api/v1/guests/{id}/recommendations/refresh")
	s.RequireSchema("POST", "/api/v1/guests/{id}/recommendations/refresh", resp)

	decoded, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode recommendations: %v", err)
	}
	items, ok := harness.Items(decoded)
	if !ok {
		t.Fatalf("recommendation response has no items array: %s", resp.Snippet())
	}
	if len(items) > 3 {
		t.Fatalf("recommendations: contract caps the list at 3, got %d", len(items))
	}
	for _, item := range items {
		obj, _ := item.(map[string]any)
		tenantID := harness.StringField(obj, "tenantId")
		if tenantID == w.tenantA.ID {
			t.Fatalf("recommendation list contains the visited tenant %s", tenantID)
		}
	}
}

func TestRecommendationsRefreshReplacesSet(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "reco-replace")

	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "reco-replace")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "reco-replace-a", 2)

	first, err := s.API.Post(ctx, "/api/v1/guests/"+guest.ID+"/recommendations/refresh", nil)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	s.RequireStatus(first, 200, "first refresh")

	second, err := s.API.Post(ctx, "/api/v1/guests/"+guest.ID+"/recommendations/refresh", nil)
	if err != nil {
		t.Fatalf("second refresh: %v", err)
	}
	s.RequireStatus(second, 200, "second refresh")

	rows, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM recommendation WHERE guest_id = $1`, guest.ID)
	if err != nil {
		t.Fatalf("count recommendations: %v", err)
	}
	decoded, err := second.JSON()
	if err != nil {
		t.Fatalf("decode second refresh: %v", err)
	}
	items, _ := harness.Items(decoded)
	if rows != len(items) {
		t.Fatalf("refresh must replace the set: %d rows in recommendation, %d items in the response", rows, len(items))
	}
}

func TestRecommendationsUnknownGuestReturns404(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/guests/00000000-0000-0000-0000-000000000000/recommendations", nil)
	if err != nil {
		t.Fatalf("GET recommendations of an unknown guest: %v", err)
	}
	s.RequireStatus(resp, 404, "GET recommendations of an unknown guest")
	s.RequireDeclaredStatus("GET", "/api/v1/guests/{id}/recommendations", resp)
	s.RequireSchema("GET", "/api/v1/guests/{id}/recommendations", resp)
}

func TestMergeCandidateListFiltersByResolution(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	guestA, err := s.Reg.Guest(ctx, "filter-a")
	if err != nil {
		t.Fatalf("fixture guest A: %v", err)
	}
	guestB, err := s.Reg.Guest(ctx, "filter-b")
	if err != nil {
		t.Fatalf("fixture guest B: %v", err)
	}
	candidate, err := s.Reg.MergeCandidate(ctx, guestA.ID, guestB.ID, "phone", "shared_phone_names_differ", 55)
	if err != nil {
		t.Fatalf("fixture candidate: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/merge-candidates", map[string][]string{"resolution": {"pending"}})
	if err != nil {
		t.Fatalf("GET /api/v1/merge-candidates?resolution=pending: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/merge-candidates?resolution=pending")
	s.RequireSchema("GET", "/api/v1/merge-candidates", resp)

	decoded, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode candidates: %v", err)
	}
	items, _ := harness.Items(decoded)
	found := false
	for _, item := range items {
		obj, _ := item.(map[string]any)
		if harness.StringField(obj, "id") == candidate.ID {
			found = true
			if resolution := harness.StringField(obj, "resolution"); resolution != "pending" {
				t.Fatalf("candidate %s: expected resolution pending, got %q", candidate.ID, resolution)
			}
		}
	}
	if !found {
		t.Fatalf("pending candidate %s is missing from the filtered queue", candidate.ID)
	}
}

func TestMergeCandidateConfirmMovesEventsAndIdentifiers(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "merge-confirm")

	ctx, cancel := s.Ctx()
	defer cancel()

	winner, err := s.Reg.Guest(ctx, "merge-winner")
	if err != nil {
		t.Fatalf("fixture winner: %v", err)
	}
	loser, err := s.Reg.Guest(ctx, "merge-loser")
	if err != nil {
		t.Fatalf("fixture loser: %v", err)
	}
	w.visit(t, s, loser.ID, w.tenantB.ID, "merge-loser-event", 3)
	if _, err := s.Reg.Identifier(ctx, loser.ID, "phone", "79"+s.Cfg.RunID+"00003", w.source.ID); err != nil {
		t.Fatalf("fixture identifier: %v", err)
	}
	candidate, err := s.Reg.MergeCandidate(ctx, winner.ID, loser.ID, "phone", "shared_phone_names_differ", 55)
	if err != nil {
		t.Fatalf("fixture candidate: %v", err)
	}

	resp, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/confirm", nil)
	if err != nil {
		t.Fatalf("POST confirm: %v", err)
	}
	s.RequireStatus(resp, 200, "POST /api/v1/merge-candidates/{id}/confirm")
	s.RequireSchema("POST", "/api/v1/merge-candidates/{id}/confirm", resp)

	decoded, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode confirm: %v", err)
	}
	obj, _ := decoded.(map[string]any)
	winnerID := harness.StringField(obj, "winnerGuestId")
	loserID := harness.StringField(obj, "mergedGuestId")
	if winnerID == "" || loserID == "" {
		t.Fatalf("confirm response must report winnerGuestId and mergedGuestId: %s", resp.Snippet())
	}

	loserEvents, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM event WHERE guest_id = $1`, loserID)
	if err != nil {
		t.Fatalf("count loser events: %v", err)
	}
	if loserEvents != 0 {
		t.Fatalf("after confirm the loser %s still owns %d events", loserID, loserEvents)
	}
	winnerEvents, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM event WHERE guest_id = $1`, winnerID)
	if err != nil {
		t.Fatalf("count winner events: %v", err)
	}
	if winnerEvents != 1 {
		t.Fatalf("after confirm the winner %s must own 1 event, got %d", winnerID, winnerEvents)
	}
	loserIdentifiers, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM guest_identifier WHERE guest_id = $1`, loserID)
	if err != nil {
		t.Fatalf("count loser identifiers: %v", err)
	}
	if loserIdentifiers != 0 {
		t.Fatalf("after confirm the loser %s still owns %d identifiers", loserID, loserIdentifiers)
	}

	profileResp, err := s.API.Get(ctx, "/api/v1/guests/"+loserID, nil)
	if err != nil {
		t.Fatalf("GET loser profile: %v", err)
	}
	s.RequireStatus(profileResp, 200, "GET merged loser profile")
	profileDecoded, err := profileResp.JSON()
	if err != nil {
		t.Fatalf("decode loser profile: %v", err)
	}
	profile, _ := profileDecoded.(map[string]any)
	if status := harness.StringField(profile, "status"); status != "merged" {
		t.Fatalf("loser status: expected merged, got %q", status)
	}
	if mergedInto := harness.StringField(profile, "mergedIntoId"); mergedInto != winnerID {
		t.Fatalf("loser mergedIntoId: expected %s, got %q", winnerID, mergedInto)
	}
}

func TestMergeCandidateConfirmTwiceReturns409(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	guestA, err := s.Reg.Guest(ctx, "twice-a")
	if err != nil {
		t.Fatalf("fixture guest A: %v", err)
	}
	guestB, err := s.Reg.Guest(ctx, "twice-b")
	if err != nil {
		t.Fatalf("fixture guest B: %v", err)
	}
	candidate, err := s.Reg.MergeCandidate(ctx, guestA.ID, guestB.ID, "phone", "shared_phone_names_differ", 55)
	if err != nil {
		t.Fatalf("fixture candidate: %v", err)
	}

	first, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/confirm", nil)
	if err != nil {
		t.Fatalf("first confirm: %v", err)
	}
	s.RequireStatus(first, 200, "first confirm")

	second, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/confirm", nil)
	if err != nil {
		t.Fatalf("second confirm: %v", err)
	}
	s.RequireStatus(second, 409, "second confirm of an already resolved candidate")
	s.RequireDeclaredStatus("POST", "/api/v1/merge-candidates/{id}/confirm", second)
	s.RequireSchema("POST", "/api/v1/merge-candidates/{id}/confirm", second)
}

func TestMergeCandidateRejectCreatesBlock(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	guestA, err := s.Reg.Guest(ctx, "reject-a")
	if err != nil {
		t.Fatalf("fixture guest A: %v", err)
	}
	guestB, err := s.Reg.Guest(ctx, "reject-b")
	if err != nil {
		t.Fatalf("fixture guest B: %v", err)
	}
	candidate, err := s.Reg.MergeCandidate(ctx, guestA.ID, guestB.ID, "phone", "shared_phone_names_differ", 55)
	if err != nil {
		t.Fatalf("fixture candidate: %v", err)
	}

	resp, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/reject", nil)
	if err != nil {
		t.Fatalf("POST reject: %v", err)
	}
	s.RequireStatus(resp, 200, "POST /api/v1/merge-candidates/{id}/reject")
	s.RequireSchema("POST", "/api/v1/merge-candidates/{id}/reject", resp)

	blocks, err := s.DB.Count(ctx, `
		SELECT COUNT(*) FROM merge_block
		WHERE (guest_a_id = $1 AND guest_b_id = $2) OR (guest_a_id = $2 AND guest_b_id = $1)`,
		guestA.ID, guestB.ID)
	if err != nil {
		t.Fatalf("count merge block: %v", err)
	}
	if blocks != 1 {
		t.Fatalf("reject must create exactly one merge_block for the pair, got %d", blocks)
	}
}

func TestMergeCandidateRevertRestoresOrigin(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "merge-revert")
	ctx, cancel := s.Ctx()
	defer cancel()

	winner, err := s.Reg.Guest(ctx, "revert-winner")
	if err != nil {
		t.Fatalf("fixture winner: %v", err)
	}
	loser, err := s.Reg.Guest(ctx, "revert-loser")
	if err != nil {
		t.Fatalf("fixture loser: %v", err)
	}
	w.visit(t, s, loser.ID, w.tenantB.ID, "revert-loser-event", 3)
	candidate, err := s.Reg.MergeCandidate(ctx, winner.ID, loser.ID, "phone", "shared_phone_names_differ", 55)
	if err != nil {
		t.Fatalf("fixture candidate: %v", err)
	}

	confirm, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/confirm", nil)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	s.RequireStatus(confirm, 200, "confirm before revert")

	revert, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/revert", nil)
	if err != nil {
		t.Fatalf("revert: %v", err)
	}
	s.RequireStatus(revert, 200, "POST /api/v1/merge-candidates/{id}/revert")
	s.RequireSchema("POST", "/api/v1/merge-candidates/{id}/revert", revert)

	loserEvents, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM event WHERE guest_id = $1`, loser.ID)
	if err != nil {
		t.Fatalf("count loser events after revert: %v", err)
	}
	if loserEvents != 1 {
		t.Fatalf("revert must return the event to the loser %s, got %d events", loser.ID, loserEvents)
	}
	status, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM guest WHERE id = $1 AND status = 'active' AND merged_into_id IS NULL`, loser.ID)
	if err != nil {
		t.Fatalf("read loser status: %v", err)
	}
	if status != 1 {
		t.Fatalf("revert must set the loser %s back to status=active with merged_into_id NULL", loser.ID)
	}
}

func TestSegmentPreviewRejectsEmptyWas(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "preview-empty-visited")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/segments/preview", map[string]any{
		"visitedTenantIds":  []string{},
		"excludedTenantIds": []string{w.tenantC.ID},
		"periodDays":        90,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/segments/preview with an empty visitedTenantIds: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/segments/preview with an empty visitedTenantIds")
	s.RequireDeclaredStatus("POST", "/api/v1/segments/preview", resp)
	s.RequireSchema("POST", "/api/v1/segments/preview", resp)
}

func TestSegmentPreviewAcceptsEmptyExcludedTenantIds(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "preview-empty-excluded")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/segments/preview", map[string]any{
		"visitedTenantIds": []string{w.tenantA.ID},
		"periodDays":       90,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/segments/preview without excludedTenantIds: %v", err)
	}
	s.RequireStatusOrDrift(resp, 200, "POST /api/v1/segments/preview without excludedTenantIds")
	s.RequireSchema("POST", "/api/v1/segments/preview", resp)
}

func TestMergeCandidateRevertPendingReturns409(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	guestA, err := s.Reg.Guest(ctx, "revert-pending-a")
	if err != nil {
		t.Fatalf("fixture guest A: %v", err)
	}
	guestB, err := s.Reg.Guest(ctx, "revert-pending-b")
	if err != nil {
		t.Fatalf("fixture guest B: %v", err)
	}
	candidate, err := s.Reg.MergeCandidate(ctx, guestA.ID, guestB.ID, "phone", "shared_phone_names_differ", 55)
	if err != nil {
		t.Fatalf("fixture candidate: %v", err)
	}

	resp, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/revert", nil)
	if err != nil {
		t.Fatalf("revert of a pending candidate: %v", err)
	}
	s.RequireStatus(resp, 409, "revert of a pending candidate")
	s.RequireDeclaredStatus("POST", "/api/v1/merge-candidates/{id}/revert", resp)
	s.RequireSchema("POST", "/api/v1/merge-candidates/{id}/revert", resp)
}

func TestMergeCandidateUnknownReturns404(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/merge-candidates/00000000-0000-0000-0000-000000000000", nil)
	if err != nil {
		t.Fatalf("GET unknown candidate: %v", err)
	}
	s.RequireStatus(resp, 404, "GET unknown merge candidate")
	s.RequireDeclaredStatus("GET", "/api/v1/merge-candidates/{id}", resp)
	s.RequireSchema("GET", "/api/v1/merge-candidates/{id}", resp)
}

func TestGuestEventsFilterSortAndCount(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "events-filter")

	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "events-filter")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	cheap, expensive := int64(10000), int64(50000)
	now := time.Now().UTC()
	if _, err := s.Reg.Event(ctx, guest.ID, w.tenantA.ID, w.eventType.ID, w.source.ID, "filter-cheap", now.AddDate(0, 0, -3), &cheap); err != nil {
		t.Fatalf("fixture event: %v", err)
	}
	if _, err := s.Reg.Event(ctx, guest.ID, w.tenantB.ID, w.eventType.ID, w.source.ID, "filter-expensive", now.AddDate(0, 0, -1), &expensive); err != nil {
		t.Fatalf("fixture event: %v", err)
	}

	fetch := func(query map[string][]string) (map[string]any, []any) {
		t.Helper()
		resp, err := s.API.Get(ctx, "/api/v1/guests/"+guest.ID+"/events", query)
		if err != nil {
			t.Fatalf("GET /api/v1/guests/{id}/events: %v", err)
		}
		s.RequireStatus(resp, 200, "GET /api/v1/guests/{id}/events")
		body, err := resp.JSON()
		if err != nil {
			t.Fatalf("decode events: %v", err)
		}
		page, _ := body.(map[string]any)
		items, ok := harness.Items(body)
		if !ok {
			t.Fatalf("event feed has no items array: %s", resp.Snippet())
		}
		return page, items
	}

	page, items := fetch(map[string][]string{"limit": {"1"}})
	meta, _ := page["meta"].(map[string]any)
	if total, _ := harness.IntField(meta, "total"); total != 2 || len(items) != 1 {
		t.Fatalf("expected total 2 with one item on the page, got total=%d items=%d", total, len(items))
	}

	_, items = fetch(map[string][]string{"tenant_id": {w.tenantA.ID}})
	if len(items) != 1 {
		t.Fatalf("tenant filter: expected 1 event, got %d", len(items))
	}

	_, items = fetch(map[string][]string{"amount_min": {"20000"}})
	if len(items) != 1 {
		t.Fatalf("amount_min filter: expected 1 event, got %d", len(items))
	}

	_, items = fetch(map[string][]string{"type": {w.eventType.Code}})
	if len(items) != 2 {
		t.Fatalf("type filter: expected 2 events, got %d", len(items))
	}
	first, _ := items[0].(map[string]any)
	if harness.StringField(first, "eventTypeCode") != w.eventType.Code || harness.StringField(first, "tenantName") == "" {
		t.Fatalf("event must carry eventTypeCode and tenantName: %v", first)
	}

	_, items = fetch(map[string][]string{"sort": {"amountMinor"}})
	first, _ = items[0].(map[string]any)
	if amount, _ := harness.IntField(first, "amountMinor"); amount != int(cheap) {
		t.Fatalf("sort=amountMinor must start with the cheapest event, got %v", first)
	}

	resp, err := s.API.Get(ctx, "/api/v1/guests/"+guest.ID+"/events", map[string][]string{"sort": {"bogus"}})
	if err != nil {
		t.Fatalf("GET with bad sort: %v", err)
	}
	s.RequireStatus(resp, 400, "GET /api/v1/guests/{id}/events with invalid sort")
}

func TestGuestListFilterSortAndCount(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "list-filter")

	ctx, cancel := s.Ctx()
	defer cancel()

	typesResp, err := s.API.Get(ctx, "/api/v1/event-types", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/event-types: %v", err)
	}
	s.RequireStatus(typesResp, 200, "GET /api/v1/event-types")
	typesBody, err := typesResp.JSON()
	if err != nil {
		t.Fatalf("decode event types: %v", err)
	}
	typeItems, _ := harness.Items(typesBody)
	purchaseTypeID := ""
	for _, item := range typeItems {
		entry, _ := item.(map[string]any)
		if harness.StringField(entry, "code") == "purchase" {
			purchaseTypeID = harness.StringField(entry, "id")
		}
	}
	if purchaseTypeID == "" {
		t.Fatalf("purchase event type is missing from the dictionary")
	}

	buyer, err := s.Reg.Guest(ctx, "list-buyer")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	visitor, err := s.Reg.Guest(ctx, "list-visitor")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	cheap, expensive := int64(10000), int64(50000)
	now := time.Now().UTC()
	if _, err := s.Reg.Event(ctx, buyer.ID, w.tenantA.ID, purchaseTypeID, w.source.ID, "list-cheap", now.AddDate(0, 0, -3), &cheap); err != nil {
		t.Fatalf("fixture event: %v", err)
	}
	if _, err := s.Reg.Event(ctx, buyer.ID, w.tenantB.ID, purchaseTypeID, w.source.ID, "list-expensive", now.AddDate(0, 0, -1), &expensive); err != nil {
		t.Fatalf("fixture event: %v", err)
	}
	if _, err := s.Reg.Event(ctx, visitor.ID, w.tenantA.ID, w.eventType.ID, w.source.ID, "list-visit", now.AddDate(0, 0, -40), nil); err != nil {
		t.Fatalf("fixture event: %v", err)
	}

	date := func(daysAgo int) string { return now.AddDate(0, 0, -daysAgo).Format("2006-01-02") }
	fetch := func(query map[string][]string) (int, []map[string]any) {
		t.Helper()
		resp, err := s.API.Get(ctx, "/api/v1/guests", query)
		if err != nil {
			t.Fatalf("GET /api/v1/guests: %v", err)
		}
		s.RequireStatus(resp, 200, "GET /api/v1/guests")
		body, err := resp.JSON()
		if err != nil {
			t.Fatalf("decode guests: %v", err)
		}
		page, _ := body.(map[string]any)
		meta, _ := page["meta"].(map[string]any)
		total, _ := harness.IntField(meta, "total")
		raw, ok := harness.Items(body)
		if !ok {
			t.Fatalf("guest list has no items array: %s", resp.Snippet())
		}
		items := make([]map[string]any, 0, len(raw))
		for _, item := range raw {
			entry, _ := item.(map[string]any)
			items = append(items, entry)
		}
		return total, items
	}
	requireOnly := func(label string, query map[string][]string, want string) {
		t.Helper()
		total, items := fetch(query)
		if total != 1 || len(items) != 1 || harness.StringField(items[0], "id") != want {
			t.Fatalf("%s: expected only guest %s, got total=%d items=%v", label, want, total, items)
		}
	}
	atTenantA := func(extra string, value string) map[string][]string {
		query := map[string][]string{"tenant_id": {w.tenantA.ID}}
		if extra != "" {
			query[extra] = []string{value}
		}
		return query
	}

	total, items := fetch(atTenantA("page_size", "1"))
	if total != 2 || len(items) != 1 {
		t.Fatalf("tenant filter: expected total 2 with one item on the page, got total=%d items=%d", total, len(items))
	}

	total, _ = fetch(map[string][]string{"source_id": {w.source.ID}})
	if total != 2 {
		t.Fatalf("source filter: expected 2 guests, got %d", total)
	}

	requireOnly("has_purchases", atTenantA("has_purchases", "true"), buyer.ID)
	requireOnly("has_purchases=false", atTenantA("has_purchases", "false"), visitor.ID)
	requireOnly("period_from", atTenantA("period_from", date(4)), buyer.ID)
	requireOnly("first_seen_from", atTenantA("first_seen_from", date(7)), buyer.ID)
	requireOnly("last_seen_to", atTenantA("last_seen_to", date(30)), visitor.ID)

	_, items = fetch(atTenantA("has_purchases", "true"))
	buyerRow := items[0]
	if count, _ := harness.IntField(buyerRow, "eventCount"); count != 2 {
		t.Fatalf("eventCount must cover all guest events, got %v", buyerRow)
	}
	if count, _ := harness.IntField(buyerRow, "purchaseCount"); count != 2 {
		t.Fatalf("purchaseCount: expected 2, got %v", buyerRow)
	}
	if amount, _ := harness.IntField(buyerRow, "totalAmountMinor"); amount != int(cheap+expensive) {
		t.Fatalf("totalAmountMinor: expected %d, got %v", cheap+expensive, buyerRow)
	}

	for _, sort := range []struct {
		value string
		first string
	}{
		{"-eventCount", buyer.ID},
		{"eventCount", visitor.ID},
		{"-totalAmountMinor", buyer.ID},
		{"totalAmountMinor", visitor.ID},
		{"-lastSeenAt", buyer.ID},
		{"lastSeenAt", visitor.ID},
		{"-firstSeenAt", buyer.ID},
		{"firstSeenAt", visitor.ID},
	} {
		_, items = fetch(atTenantA("sort", sort.value))
		if len(items) != 2 || harness.StringField(items[0], "id") != sort.first {
			t.Fatalf("sort=%s: expected guest %s first, got %v", sort.value, sort.first, items)
		}
	}

	for _, bad := range []map[string][]string{
		{"sort": {"bogus"}},
		{"status": {"bogus"}},
		{"has_purchases": {"maybe"}},
		{"period_from": {"yesterday"}},
	} {
		resp, err := s.API.Get(ctx, "/api/v1/guests", bad)
		if err != nil {
			t.Fatalf("GET /api/v1/guests with %v: %v", bad, err)
		}
		s.RequireStatus(resp, 400, "GET /api/v1/guests with invalid parameter")
	}
}
