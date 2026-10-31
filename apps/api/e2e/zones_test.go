package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/harness"
)

type zoneWorld struct {
	source       string
	mallA        string
	mallB        string
	tenantA2     string
	tenantA3     string
	tenantB2     string
	buyerA2      string
	visitorA2    string
	buyerB2      string
	buyerA3      string
	buyerA2AndB2 string
}

func buildZoneWorld(t *testing.T, s *harness.Suite, suffix string) *zoneWorld {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	category, err := s.Reg.Category(ctx, suffix)
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	mallA, err := s.Reg.Mall(ctx, suffix+"-a")
	if err != nil {
		t.Fatalf("fixture mall A: %v", err)
	}
	mallB, err := s.Reg.Mall(ctx, suffix+"-b")
	if err != nil {
		t.Fatalf("fixture mall B: %v", err)
	}
	tenantA2, err := s.Reg.TenantInZone(ctx, suffix+"-a2", category.ID, mallA.ID, 2)
	if err != nil {
		t.Fatalf("fixture tenant A2: %v", err)
	}
	tenantA3, err := s.Reg.TenantInZone(ctx, suffix+"-a3", category.ID, mallA.ID, 3)
	if err != nil {
		t.Fatalf("fixture tenant A3: %v", err)
	}
	tenantB2, err := s.Reg.TenantInZone(ctx, suffix+"-b2", category.ID, mallB.ID, 2)
	if err != nil {
		t.Fatalf("fixture tenant B2: %v", err)
	}
	source, err := s.Reg.DataSource(ctx, suffix, tenantA2.ID)
	if err != nil {
		t.Fatalf("fixture data source: %v", err)
	}

	guest := func(name string) string {
		g, err := s.Reg.Guest(ctx, suffix+"-"+name)
		if err != nil {
			t.Fatalf("fixture guest %s: %v", name, err)
		}
		return g.ID
	}
	event := func(code, guestID, tenantID, key string, hoursAgo int) {
		at := time.Now().UTC().Add(-time.Duration(hoursAgo) * time.Hour)
		if err := s.Reg.PresenceEvent(ctx, code, guestID, tenantID, source.ID, suffix+"-"+key, at); err != nil {
			t.Fatalf("fixture event %s: %v", key, err)
		}
	}

	w := &zoneWorld{
		source:       source.ID,
		mallA:        mallA.ID,
		mallB:        mallB.ID,
		tenantA2:     tenantA2.ID,
		tenantA3:     tenantA3.ID,
		tenantB2:     tenantB2.ID,
		buyerA2:      guest("buyer-a2"),
		visitorA2:    guest("visitor-a2"),
		buyerB2:      guest("buyer-b2"),
		buyerA3:      guest("buyer-a3"),
		buyerA2AndB2: guest("buyer-both"),
	}
	event("purchase", w.buyerA2, w.tenantA2, "buyer-a2", 5)
	event("visit", w.visitorA2, w.tenantA2, "visitor-a2", 5)
	event("purchase", w.buyerB2, w.tenantB2, "buyer-b2", 5)
	event("purchase", w.buyerA3, w.tenantA3, "buyer-a3", 5)
	event("purchase", w.buyerA2AndB2, w.tenantA2, "both-a2", 6)
	event("purchase", w.buyerA2AndB2, w.tenantB2, "both-b2", 4)
	return w
}

func previewMembers(t *testing.T, s *harness.Suite, rule map[string]any) []string {
	t.Helper()
	ctx, cancel := s.Ctx()
	defer cancel()
	resp, err := s.API.Post(ctx, "/api/v1/segments/preview", map[string]any{"rule": rule, "limit": 100})
	if err != nil {
		t.Fatalf("POST /api/v1/segments/preview: %v", err)
	}
	s.RequireStatus(resp, 200, "POST /api/v1/segments/preview")
	body := jsonObject(t, resp, "preview")
	examples, _ := body["examples"].([]any)
	out := make([]string, 0, len(examples))
	for _, item := range examples {
		out = append(out, harness.StringField(item.(map[string]any), "guestId"))
	}
	return out
}

func TestSegmentPurchasedZones_CountsOnlyPurchasesInTheZone(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildZoneWorld(t, s, "zone-purchase")

	got := previewMembers(t, s, map[string]any{
		"periodDays":     30,
		"purchasedZones": []map[string]any{{"mallId": w.mallA, "floor": 2}},
	})

	if !containsString(got, w.buyerA2) || !containsString(got, w.buyerA2AndB2) {
		t.Fatalf("buyers on mall A floor 2 must match, got %v", got)
	}
	for name, id := range map[string]string{"visitor": w.visitorA2, "other mall": w.buyerB2, "other floor": w.buyerA3} {
		if containsString(got, id) {
			t.Fatalf("%s must not match the zone, got %v", name, got)
		}
	}
}

func TestSegmentPurchasedZones_MatchesAnyOfSeveralZones(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildZoneWorld(t, s, "zone-any")

	got := previewMembers(t, s, map[string]any{
		"periodDays": 30,
		"purchasedZones": []map[string]any{
			{"mallId": w.mallA, "floor": 3},
			{"mallId": w.mallB, "floor": 2},
		},
	})

	for _, id := range []string{w.buyerA3, w.buyerB2, w.buyerA2AndB2} {
		if !containsString(got, id) {
			t.Fatalf("guest %s must match one of the zones, got %v", id, got)
		}
	}
	if containsString(got, w.buyerA2) {
		t.Fatalf("a buyer only on mall A floor 2 must not match, got %v", got)
	}
}

func TestSegmentPurchasedZones_CombinesWithExcludedTenants(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildZoneWorld(t, s, "zone-exclude")

	got := previewMembers(t, s, map[string]any{
		"periodDays":        30,
		"purchasedZones":    []map[string]any{{"mallId": w.mallA, "floor": 2}},
		"excludedTenantIds": []string{w.tenantB2},
	})

	if !containsString(got, w.buyerA2) {
		t.Fatalf("the buyer who never visited the excluded tenant must match, got %v", got)
	}
	if containsString(got, w.buyerA2AndB2) {
		t.Fatalf("the buyer at the excluded tenant must not match, got %v", got)
	}
}

func TestSegmentRuleGroups_CombineZonesWithAndOr(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildZoneWorld(t, s, "zone-group")
	zoneA2 := map[string]any{"periodDays": 30, "purchasedZones": []map[string]any{{"mallId": w.mallA, "floor": 2}}}
	zoneB2 := map[string]any{"periodDays": 30, "purchasedZones": []map[string]any{{"mallId": w.mallB, "floor": 2}}}

	both := previewMembers(t, s, map[string]any{"operator": "and", "conditions": []any{zoneA2, zoneB2}})
	if len(both) != 1 || both[0] != w.buyerA2AndB2 {
		t.Fatalf("and of two zones must keep only the guest who bought in both, got %v", both)
	}

	either := previewMembers(t, s, map[string]any{"operator": "or", "conditions": []any{zoneA2, zoneB2}})
	for _, id := range []string{w.buyerA2, w.buyerB2, w.buyerA2AndB2} {
		if !containsString(either, id) {
			t.Fatalf("or of two zones must include %s, got %v", id, either)
		}
	}
	if containsString(either, w.buyerA3) {
		t.Fatalf("a buyer outside both zones must not match, got %v", either)
	}
}

func TestSegmentRuleGroups_RejectsEmptyGroup(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/segments/preview", map[string]any{
		"rule": map[string]any{"operator": "and", "conditions": []any{}},
	})
	if err != nil {
		t.Fatalf("POST /api/v1/segments/preview: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/segments/preview with empty group")
}

func TestGuestMovement_ListsPresenceEventsWithMallAndFloor(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildZoneWorld(t, s, "movement")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/guests/"+w.buyerA2AndB2+"/movement", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{id}/movement: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests/{id}/movement")
	s.RequireDeclaredStatus("GET", "/api/v1/guests/{id}/movement", resp)
	s.RequireSchema("GET", "/api/v1/guests/{id}/movement", resp)

	items, _ := jsonObject(t, resp, "movement")["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected two steps, got %d: %s", len(items), resp.Snippet())
	}
	first, second := items[0].(map[string]any), items[1].(map[string]any)
	if harness.StringField(first, "mallId") != w.mallA || harness.StringField(second, "mallId") != w.mallB {
		t.Fatalf("steps must follow the visit order mall A then mall B: %s", resp.Snippet())
	}
	if floor, _ := harness.IntField(first, "floor"); floor != 2 {
		t.Fatalf("first step floor: expected 2, got %d", floor)
	}
}

func TestGuestMovement_ReturnsNotFoundForUnknownGuest(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/guests/00000000-0000-0000-0000-00000000dead/movement", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{id}/movement: %v", err)
	}
	s.RequireStatus(resp, 404, "GET /api/v1/guests/{id}/movement for an unknown guest")
}

func TestMallCrud_RejectsDeleteWhileTenantsExist(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	created, err := s.API.Post(ctx, "/api/v1/malls", map[string]any{"code": s.Reg.Code("api-mall"), "name": "E2E API mall"})
	if err != nil {
		t.Fatalf("POST /api/v1/malls: %v", err)
	}
	s.RequireStatus(created, 201, "POST /api/v1/malls")
	s.RequireDeclaredStatus("POST", "/api/v1/malls", created)
	s.RequireSchema("POST", "/api/v1/malls", created)
	mallID := harness.StringField(jsonObject(t, created, "created mall"), "id")
	s.T.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cancel()
		_ = s.DB.Exec(cleanup, `DELETE FROM tenant WHERE mall_id = $1`, mallID)
		_ = s.DB.Exec(cleanup, `DELETE FROM mall WHERE id = $1`, mallID)
	})

	category, err := s.Reg.Category(ctx, "mall-crud")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	tenant, err := s.API.Post(ctx, "/api/v1/tenants", map[string]any{
		"code":       s.Reg.Code("mall-tenant"),
		"name":       "E2E mall tenant",
		"mallId":     mallID,
		"categoryId": category.ID,
		"floor":      1,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/tenants: %v", err)
	}
	s.RequireStatus(tenant, 201, "POST /api/v1/tenants with mallId")
	if got := harness.StringField(jsonObject(t, tenant, "created tenant"), "mallId"); got != mallID {
		t.Fatalf("tenant mallId: expected %q, got %q", mallID, got)
	}

	list, err := s.API.Get(ctx, "/api/v1/malls", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/malls: %v", err)
	}
	s.RequireStatus(list, 200, "GET /api/v1/malls")
	s.RequireSchema("GET", "/api/v1/malls", list)

	blocked, err := s.API.Delete(ctx, "/api/v1/malls/"+mallID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/malls/{id}: %v", err)
	}
	s.RequireStatus(blocked, 409, "DELETE /api/v1/malls/{id} with tenants")
}

func TestTenantCreate_RejectsUnknownMall(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "tenant-no-mall")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	resp, err := s.API.Post(ctx, "/api/v1/tenants", map[string]any{
		"code":       s.Reg.Code("no-mall"),
		"name":       "E2E tenant without mall",
		"mallId":     "00000000-0000-0000-0000-00000000dead",
		"categoryId": category.ID,
		"floor":      1,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/tenants: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/tenants with an unknown mall")
}
