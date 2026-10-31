package e2e

import (
	"context"
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestE2EIdentityMergeLifecycle(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "e2e-identity")
	ctx, cancel := s.Ctx()
	defer cancel()

	winner, err := s.Reg.Guest(ctx, "e2e-identity-winner")
	if err != nil {
		t.Fatalf("fixture winner: %v", err)
	}
	loser, err := s.Reg.Guest(ctx, "e2e-identity-loser")
	if err != nil {
		t.Fatalf("fixture loser: %v", err)
	}
	sharedPhone := s.Reg.Phone(14)
	if _, err := s.Reg.Identifier(ctx, winner.ID, "phone", sharedPhone, w.source.ID); err != nil {
		t.Fatalf("fixture winner identifier: %v", err)
	}
	if _, err := s.Reg.Identifier(ctx, loser.ID, "email", s.Reg.Name("loser")+"@e2e.local", w.source.ID); err != nil {
		t.Fatalf("fixture loser identifier: %v", err)
	}
	w.visit(t, s, loser.ID, w.tenantB.ID, "e2e-identity-loser-event", 4)
	w.visit(t, s, winner.ID, w.tenantA.ID, "e2e-identity-winner-event", 2)

	candidate, err := s.Reg.MergeCandidate(ctx, winner.ID, loser.ID, "phone", "shared_phone_names_differ", 55)
	if err != nil {
		t.Fatalf("fixture candidate: %v", err)
	}

	queue, err := s.API.Get(ctx, "/api/v1/merge-candidates", map[string][]string{"resolution": {"pending"}})
	if err != nil {
		t.Fatalf("step 1 GET pending queue: %v", err)
	}
	s.RequireStatus(queue, 200, "step 1: pending merge queue")
	s.RequireSchema("GET", "/api/v1/merge-candidates", queue)
	queueBody, err := queue.JSON()
	if err != nil {
		t.Fatalf("step 1 decode queue: %v", err)
	}
	queueItems, _ := harness.Items(queueBody)
	if !containsID(queueItems, candidate.ID) {
		t.Fatalf("step 1: candidate %s is absent from the pending queue", candidate.ID)
	}

	detail, err := s.API.Get(ctx, "/api/v1/merge-candidates/"+candidate.ID, nil)
	if err != nil {
		t.Fatalf("step 2 GET candidate detail: %v", err)
	}
	s.RequireStatus(detail, 200, "step 2: candidate detail")
	s.RequireSchema("GET", "/api/v1/merge-candidates/{id}", detail)

	confirm, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/confirm", nil)
	if err != nil {
		t.Fatalf("step 3 confirm: %v", err)
	}
	s.RequireStatus(confirm, 200, "step 3: confirm merge")
	s.RequireSchema("POST", "/api/v1/merge-candidates/{id}/confirm", confirm)

	confirmBody, err := confirm.JSON()
	if err != nil {
		t.Fatalf("step 3 decode confirm: %v", err)
	}
	confirmObj, _ := confirmBody.(map[string]any)
	winnerID := harness.StringField(confirmObj, "winnerGuestId")
	loserID := harness.StringField(confirmObj, "mergedGuestId")
	if winnerID == "" || loserID == "" {
		t.Fatalf("step 3: the confirm response must report winnerGuestId and mergedGuestId: %s", confirm.Snippet())
	}
	if winnerID == loserID {
		t.Fatalf("step 3: winner and merged guest must differ, both are %s", winnerID)
	}

	loserProfile, err := s.API.Get(ctx, "/api/v1/guests/"+loserID, nil)
	if err != nil {
		t.Fatalf("step 4 GET loser profile: %v", err)
	}
	s.RequireStatus(loserProfile, 200, "step 4: loser profile after merge")
	loserBody, err := loserProfile.JSON()
	if err != nil {
		t.Fatalf("step 4 decode loser profile: %v", err)
	}
	loserObj, _ := loserBody.(map[string]any)
	if status := harness.StringField(loserObj, "status"); status != "merged" {
		t.Fatalf("step 4: loser status must be merged, got %q", status)
	}
	if mergedInto := harness.StringField(loserObj, "mergedIntoId"); mergedInto != winnerID {
		t.Fatalf("step 4: loser mergedIntoId must be %s, got %q", winnerID, mergedInto)
	}

	winnerEvents, err := s.API.Get(ctx, "/api/v1/guests/"+winnerID+"/events", nil)
	if err != nil {
		t.Fatalf("step 5 GET winner events: %v", err)
	}
	s.RequireStatus(winnerEvents, 200, "step 5: winner event feed")
	winnerBody, err := winnerEvents.JSON()
	if err != nil {
		t.Fatalf("step 5 decode winner events: %v", err)
	}
	winnerItems, _ := harness.Items(winnerBody)
	if len(winnerItems) != 2 {
		t.Fatalf("step 5: winner must own both events after the merge, got %d", len(winnerItems))
	}

	winnerProfile, err := s.API.Get(ctx, "/api/v1/guests/"+winnerID, nil)
	if err != nil {
		t.Fatalf("step 6 GET winner profile: %v", err)
	}
	s.RequireStatus(winnerProfile, 200, "step 6: winner profile")
	winnerProfileBody, err := winnerProfile.JSON()
	if err != nil {
		t.Fatalf("step 6 decode winner profile: %v", err)
	}
	winnerObj, _ := winnerProfileBody.(map[string]any)
	if count, ok := harness.IntField(winnerObj, "tenantCount"); !ok || count != 2 {
		t.Fatalf("step 6: winner tenantCount must be 2 after absorbing the loser, got %v", count)
	}

	repeat, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/confirm", nil)
	if err != nil {
		t.Fatalf("step 7 repeat confirm: %v", err)
	}
	s.RequireStatus(repeat, 409, "step 7: repeated confirm must be rejected as a state conflict")

	revert, err := s.API.Post(ctx, "/api/v1/merge-candidates/"+candidate.ID+"/revert", nil)
	if err != nil {
		t.Fatalf("step 8 revert: %v", err)
	}
	s.RequireStatus(revert, 200, "step 8: revert merge")
	s.RequireSchema("POST", "/api/v1/merge-candidates/{id}/revert", revert)

	loserEvents, err := s.API.Get(ctx, "/api/v1/guests/"+loserID+"/events", nil)
	if err != nil {
		t.Fatalf("step 9 GET loser events after revert: %v", err)
	}
	s.RequireStatus(loserEvents, 200, "step 9: loser event feed after revert")
	loserEventsBody, err := loserEvents.JSON()
	if err != nil {
		t.Fatalf("step 9 decode loser events: %v", err)
	}
	loserItems, _ := harness.Items(loserEventsBody)
	if len(loserItems) != 1 {
		t.Fatalf("step 9: revert must return exactly one event to the loser, got %d", len(loserItems))
	}

	blocks, err := s.DB.Count(ctx, `
		SELECT COUNT(*) FROM merge_block
		WHERE (guest_a_id = $1 AND guest_b_id = $2) OR (guest_a_id = $2 AND guest_b_id = $1)`,
		winnerID, loserID)
	if err != nil {
		t.Fatalf("step 10 count merge block: %v", err)
	}
	if blocks != 1 {
		t.Fatalf("step 10: revert must leave exactly one merge_block for the pair, got %d", blocks)
	}
}

func containsID(items []any, id string) bool {
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if harness.StringField(obj, "id") == id {
			return true
		}
	}
	return false
}

func TestE2ESegmentToRecommendation(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	w := buildWorld(t, s, "e2e-segment-reco")
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "e2e-segment-reco")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "e2e-segment-reco-a", 3)

	preview, err := s.PostContractShape(ctx, "/api/v1/segments/preview",
		map[string]any{"rule": map[string]any{"visitedTenantIds": []string{w.tenantA.ID}, "excludedTenantIds": []string{w.tenantC.ID}, "periodDays": 90}},
		map[string]any{"visitedTenantIds": []string{w.tenantA.ID}, "excludedTenantIds": []string{w.tenantC.ID}, "periodDays": 90})
	if err != nil {
		t.Fatalf("step 1 preview: %v", err)
	}
	s.RequireStatusOrDrift(preview, 200, "step 1: segment preview")
	previewBody, err := preview.JSON()
	if err != nil {
		t.Fatalf("step 1 decode preview: %v", err)
	}
	previewObj, _ := previewBody.(map[string]any)
	size, ok := harness.IntField(previewObj, "size")
	if !ok || size < 1 {
		t.Fatalf("step 1: preview size must include the fixture guest, got %v", size)
	}

	create, err := s.PostContractShape(ctx, "/api/v1/segments",
		map[string]any{"name": s.Reg.Name("e2e-segment"), "rule": map[string]any{"visitedTenantIds": []string{w.tenantA.ID}, "excludedTenantIds": []string{w.tenantC.ID}, "periodDays": 90}},
		map[string]any{"name": s.Reg.Name("e2e-segment"), "visitedTenantIds": []string{w.tenantA.ID}, "excludedTenantIds": []string{w.tenantC.ID}, "periodDays": 90})
	if err != nil {
		t.Fatalf("step 2 create segment: %v", err)
	}
	s.RequireStatusOrDrift(create, 201, "step 2: create segment")
	s.RequireSchema("POST", "/api/v1/segments", create)
	createBody, err := create.JSON()
	if err != nil {
		t.Fatalf("step 2 decode created segment: %v", err)
	}
	createdObj, _ := createBody.(map[string]any)
	segmentID := harness.StringField(createdObj, "id")
	if segmentID == "" {
		t.Fatalf("step 2: created segment has no id: %s", create.Snippet())
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cleanupCancel()
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM segment_member WHERE segment_version_id IN (SELECT id FROM segment_version WHERE segment_id = $1)`, segmentID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM segment_version WHERE segment_id = $1`, segmentID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM segment WHERE id = $1`, segmentID)
	})

	get, err := s.API.Get(ctx, "/api/v1/segments/"+segmentID, nil)
	if err != nil {
		t.Fatalf("step 3 GET segment: %v", err)
	}
	s.RequireStatus(get, 200, "step 3: read the created segment")
	s.RequireSchema("GET", "/api/v1/segments/{id}", get)
	getBody, err := get.JSON()
	if err != nil {
		t.Fatalf("step 3 decode segment: %v", err)
	}
	getObj, _ := getBody.(map[string]any)
	if count, ok := harness.IntField(getObj, "memberCount"); !ok || count < 1 {
		t.Fatalf("step 3: created segment must materialize at least one member, got %v", count)
	}

	refresh, err := s.API.Post(ctx, "/api/v1/guests/"+guest.ID+"/recommendations/refresh", nil)
	if err != nil {
		t.Fatalf("step 4 refresh recommendations: %v", err)
	}
	s.RequireStatus(refresh, 200, "step 4: refresh recommendations")
	s.RequireSchema("POST", "/api/v1/guests/{id}/recommendations/refresh", refresh)
	refreshBody, err := refresh.JSON()
	if err != nil {
		t.Fatalf("step 4 decode recommendations: %v", err)
	}
	items, _ := harness.Items(refreshBody)
	for _, item := range items {
		obj, _ := item.(map[string]any)
		if harness.StringField(obj, "tenantId") == w.tenantA.ID {
			t.Fatalf("step 4: recommendations must not include the visited tenant %s", w.tenantA.ID)
		}
	}

	list, err := s.API.Get(ctx, "/api/v1/guests/"+guest.ID+"/recommendations", nil)
	if err != nil {
		t.Fatalf("step 5 GET recommendations: %v", err)
	}
	s.RequireStatus(list, 200, "step 5: read recommendations")
	s.RequireSchema("GET", "/api/v1/guests/{id}/recommendations", list)
	listBody, err := list.JSON()
	if err != nil {
		t.Fatalf("step 5 decode recommendations: %v", err)
	}
	listItems, _ := harness.Items(listBody)
	if len(listItems) != len(items) {
		t.Fatalf("step 5: reading recommendations must return the refreshed set (%d), got %d", len(items), len(listItems))
	}
}
