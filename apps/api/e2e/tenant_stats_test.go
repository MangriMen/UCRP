package e2e

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func tenantListItems(t *testing.T, body any) map[string]map[string]any {
	t.Helper()
	items, ok := harness.Items(body)
	if !ok {
		t.Fatalf("tenant list has no items: %v", body)
	}
	byID := make(map[string]map[string]any, len(items))
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("tenant list item is %T", raw)
		}
		byID[harness.StringField(item, "id")] = item
	}
	return byID
}

func TestTenantList_FiltersSortsAndReportsExactTotal(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "tenant-list")
	ctx, cancel := s.Ctx()
	defer cancel()

	query := url.Values{"q": {"E2E tenant tenant-list"}, "sort": {"-priority"}, "page_size": {"2"}}
	resp, err := s.API.Get(ctx, "/api/v1/tenants", query)
	if err != nil {
		t.Fatalf("GET /api/v1/tenants: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/tenants")
	s.RequireDeclaredStatus("GET", "/api/v1/tenants", resp)
	s.RequireSchema("GET", "/api/v1/tenants", resp)

	body := jsonObject(t, resp, "tenant list")
	items, _ := harness.Items(body)
	if len(items) != 2 {
		t.Fatalf("page_size=2 must return two items, got %d", len(items))
	}
	if first := harness.StringField(items[0].(map[string]any), "id"); first != w.tenantA.ID {
		t.Fatalf("-priority must put the highest priority first: got %s want %s", first, w.tenantA.ID)
	}
	meta, _ := body["meta"].(map[string]any)
	if total, _ := harness.IntField(meta, "total"); total != 3 {
		t.Fatalf("meta.total must count all matches, got %d", total)
	}

	inactive, err := s.API.Get(ctx, "/api/v1/tenants", url.Values{"q": {"E2E tenant tenant-list"}, "is_active": {"false"}})
	if err != nil {
		t.Fatalf("GET /api/v1/tenants is_active=false: %v", err)
	}
	s.RequireStatus(inactive, 200, "GET /api/v1/tenants is_active=false")
	if got := len(tenantListItems(t, mustJSON(t, inactive, "inactive tenants"))); got != 0 {
		t.Fatalf("no tenant is inactive, got %d", got)
	}

	bad, err := s.API.Get(ctx, "/api/v1/tenants", url.Values{"sort": {"bogus"}})
	if err != nil {
		t.Fatalf("GET /api/v1/tenants sort=bogus: %v", err)
	}
	s.RequireStatus(bad, 400, "GET /api/v1/tenants with an unknown sort")
}

func TestTenantList_FiltersByMall(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "tenant-mall")
	ctx, cancel := s.Ctx()
	defer cancel()

	mall, err := s.Reg.Mall(ctx, "tenant-mall")
	if err != nil {
		t.Fatalf("fixture mall: %v", err)
	}
	inMall, err := s.Reg.TenantInZone(ctx, "tenant-mall-zone", w.category.ID, mall.ID, 1)
	if err != nil {
		t.Fatalf("fixture tenant in mall: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/tenants", url.Values{"q": {"E2E tenant tenant-mall"}, "mall_id": {mall.ID}})
	if err != nil {
		t.Fatalf("GET /api/v1/tenants mall_id: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/tenants with mall_id")
	s.RequireSchema("GET", "/api/v1/tenants", resp)

	byID := tenantListItems(t, mustJSON(t, resp, "tenants of a mall"))
	if len(byID) != 1 || byID[inMall.ID] == nil {
		t.Fatalf("mall_id must keep only the tenant of that mall: %s", resp.Snippet())
	}
	if got := harness.StringField(byID[inMall.ID], "mallId"); got != mall.ID {
		t.Fatalf("mallId = %s, want %s", got, mall.ID)
	}
}

func TestTenantList_ReportsPeriodStatisticsAndSuppressesSmallGroups(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "tenant-stats")
	ctx, cancel := s.Ctx()
	defer cancel()

	for i := 0; i < 5; i++ {
		guest, err := s.Reg.Guest(ctx, "tenant-stats-"+strconv.Itoa(i))
		if err != nil {
			t.Fatalf("fixture guest: %v", err)
		}
		w.visit(t, s, guest.ID, w.tenantA.ID, "tenant-stats-a-"+strconv.Itoa(i), 2)
	}
	lone, err := s.Reg.Guest(ctx, "tenant-stats-lone")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, lone.ID, w.tenantB.ID, "tenant-stats-b", 2)

	from := time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02")
	to := time.Now().UTC().Format("2006-01-02")
	resp, err := s.API.Get(ctx, "/api/v1/tenants", url.Values{
		"q": {"E2E tenant tenant-stats"}, "period_from": {from}, "period_to": {to}, "sort": {"-visitsCount"},
	})
	if err != nil {
		t.Fatalf("GET /api/v1/tenants: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/tenants with a period")
	s.RequireSchema("GET", "/api/v1/tenants", resp)

	body := jsonObject(t, resp, "tenant stats")
	items, _ := harness.Items(body)
	if got := harness.StringField(items[0].(map[string]any), "id"); got != w.tenantA.ID {
		t.Fatalf("-visitsCount must start with the tenant that has visits, got %s", got)
	}
	byID := tenantListItems(t, body)
	a, b := byID[w.tenantA.ID], byID[w.tenantB.ID]
	if v, _ := harness.IntField(a, "visitsCount"); v != 5 {
		t.Fatalf("tenant A visitsCount = %v, want 5", a["visitsCount"])
	}
	if v, _ := harness.IntField(a, "newGuestsCount"); v != 5 {
		t.Fatalf("tenant A newGuestsCount = %v, want 5", a["newGuestsCount"])
	}
	if b["visitsCount"] != nil || b["guestsCount"] != nil || b["suppressed"] != true {
		t.Fatalf("one guest is below k = 5 and must be suppressed: %v", b)
	}
	if n, _ := harness.IntField(a, "sourcesCount"); n != 1 {
		t.Fatalf("tenant A has one active source, got %v", a["sourcesCount"])
	}
}

func TestTenantAggregates_ReturnsDailySeriesForPeriod(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "tenant-series")
	ctx, cancel := s.Ctx()
	defer cancel()

	for i := 0; i < 5; i++ {
		guest, err := s.Reg.Guest(ctx, "tenant-series-"+strconv.Itoa(i))
		if err != nil {
			t.Fatalf("fixture guest: %v", err)
		}
		w.visit(t, s, guest.ID, w.tenantA.ID, "tenant-series-"+strconv.Itoa(i), 3)
	}

	now := time.Now().UTC()
	query := url.Values{"from": {now.AddDate(0, 0, -6).Format("2006-01-02")}, "to": {now.Format("2006-01-02")}}
	resp, err := s.API.Get(ctx, "/api/v1/tenants/"+w.tenantA.ID+"/aggregates", query)
	if err != nil {
		t.Fatalf("GET /api/v1/tenants/{id}/aggregates: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/tenants/{id}/aggregates with a period")
	s.RequireDeclaredStatus("GET", "/api/v1/tenants/{id}/aggregates", resp)
	s.RequireSchema("GET", "/api/v1/tenants/{id}/aggregates", resp)

	aggregate := jsonObject(t, resp, "tenant aggregate")
	if v, _ := harness.IntField(aggregate, "visitsCount"); v != 5 {
		t.Fatalf("visitsCount = %v, want 5", aggregate["visitsCount"])
	}
	series, _ := aggregate["series"].([]any)
	if len(series) != 7 {
		t.Fatalf("a seven day period must return seven points, got %d", len(series))
	}
	busyDay := now.AddDate(0, 0, -3).Format("2006-01-02")
	for _, raw := range series {
		point := raw.(map[string]any)
		day := harness.StringField(point, "date")
		visits, hasVisits := harness.IntField(point, "visitsCount")
		if day == busyDay && (!hasVisits || visits != 5) {
			t.Fatalf("%s must show five visits: %v", day, point)
		}
		if day != busyDay && point["visitsCount"] != nil {
			t.Fatalf("%s has no guests and must be suppressed: %v", day, point)
		}
	}

	bad, err := s.API.Get(ctx, "/api/v1/tenants/"+w.tenantA.ID+"/aggregates", url.Values{"from": {"2026-01-01"}})
	if err != nil {
		t.Fatalf("GET aggregates with only from: %v", err)
	}
	s.RequireStatus(bad, 400, "GET aggregates with only one period bound")
}

func TestTenantGuests_ListsOnlyGuestsOfTheTenant(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "tenant-guests")
	ctx, cancel := s.Ctx()
	defer cancel()

	here, err := s.Reg.Guest(ctx, "tenant-guests-here")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	elsewhere, err := s.Reg.Guest(ctx, "tenant-guests-elsewhere")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, here.ID, w.tenantA.ID, "tenant-guests-a", 1)
	w.visit(t, s, elsewhere.ID, w.tenantB.ID, "tenant-guests-b", 1)

	resp, err := s.API.Get(ctx, "/api/v1/tenants/"+w.tenantA.ID+"/guests", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/tenants/{id}/guests: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/tenants/{id}/guests")
	s.RequireDeclaredStatus("GET", "/api/v1/tenants/{id}/guests", resp)
	s.RequireSchema("GET", "/api/v1/tenants/{id}/guests", resp)

	body := jsonObject(t, resp, "tenant guests")
	items, _ := harness.Items(body)
	if len(items) != 1 || harness.StringField(items[0].(map[string]any), "id") != here.ID {
		t.Fatalf("only the guest of tenant A must be listed: %s", resp.Snippet())
	}
	meta, _ := body["meta"].(map[string]any)
	if total, _ := harness.IntField(meta, "total"); total != 1 {
		t.Fatalf("meta.total = %d, want 1", total)
	}

	missing, err := s.API.Get(ctx, "/api/v1/tenants/"+s.Reg.NewID()+"/guests", nil)
	if err != nil {
		t.Fatalf("GET guests of a missing tenant: %v", err)
	}
	s.RequireStatus(missing, 404, "GET /api/v1/tenants/{id}/guests for an unknown tenant")
}

func TestTenantDeactivate_DisablesItsDataSources(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "tenant-off")
	ctx, cancel := s.Ctx()
	defer cancel()

	patch, err := s.API.Patch(ctx, "/api/v1/tenants/"+w.tenantA.ID, map[string]any{"isActive": false})
	if err != nil {
		t.Fatalf("PATCH /api/v1/tenants/{id}: %v", err)
	}
	s.RequireStatus(patch, 200, "PATCH /api/v1/tenants/{id} isActive=false")

	source, err := s.API.Get(ctx, "/api/v1/data-sources/"+w.source.ID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/data-sources/{id}: %v", err)
	}
	s.RequireStatus(source, 200, "GET /api/v1/data-sources/{id}")
	if active, _ := jsonObject(t, source, "data source")["isActive"].(bool); active {
		t.Fatalf("a source of a deactivated tenant must be disabled: %s", source.Snippet())
	}
}

func TestTenantCreate_RejectsInvalidReferences(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "tenant-invalid")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	mall, err := s.Reg.Mall(ctx, "tenant-invalid")
	if err != nil {
		t.Fatalf("fixture mall: %v", err)
	}
	base := func() map[string]any {
		return map[string]any{
			"code": s.Reg.Code("invalid-tenant"), "name": "E2E invalid tenant",
			"mallId": mall.ID, "categoryId": category.ID,
		}
	}
	for name, mutate := range map[string]func(map[string]any){
		"unknown category":     func(p map[string]any) { p["categoryId"] = s.Reg.NewID() },
		"priority above range": func(p map[string]any) { p["priority"] = 101 },
		"unknown kind":         func(p map[string]any) { p["kind"] = "kiosk" },
	} {
		payload := base()
		mutate(payload)
		resp, err := s.API.Post(ctx, "/api/v1/tenants", payload)
		if err != nil {
			t.Fatalf("POST /api/v1/tenants (%s): %v", name, err)
		}
		s.RequireStatus(resp, 400, "POST /api/v1/tenants with "+name)
	}
}

func purchaseEventTypeID(t *testing.T, s *harness.Suite) string {
	t.Helper()
	ctx, cancel := s.Ctx()
	defer cancel()
	var id string
	if err := s.DB.DB.QueryRowContext(ctx, `SELECT id FROM event_type WHERE code = 'purchase'`).Scan(&id); err != nil {
		t.Fatalf("purchase event type: %v", err)
	}
	return id
}

func seedTenantPurchases(t *testing.T, s *harness.Suite, w *world, suffix string) {
	t.Helper()
	ctx, cancel := s.Ctx()
	defer cancel()
	purchaseType := purchaseEventTypeID(t, s)
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	for i := 0; i < 5; i++ {
		regular, err := s.Reg.Guest(ctx, suffix+"-regular-"+strconv.Itoa(i))
		if err != nil {
			t.Fatalf("fixture guest: %v", err)
		}
		w.visit(t, s, regular.ID, w.tenantA.ID, suffix+"-old-"+strconv.Itoa(i), 30)
		w.visit(t, s, regular.ID, w.tenantA.ID, suffix+"-again-"+strconv.Itoa(i), 2)

		buyer, err := s.Reg.Guest(ctx, suffix+"-buyer-"+strconv.Itoa(i))
		if err != nil {
			t.Fatalf("fixture guest: %v", err)
		}
		amount := int64(1000 * (i + 1))
		if _, err := s.Reg.Event(ctx, buyer.ID, w.tenantA.ID, purchaseType, w.source.ID, suffix+"-buy-"+strconv.Itoa(i), yesterday, &amount); err != nil {
			t.Fatalf("fixture purchase: %v", err)
		}
	}
	lone, err := s.Reg.Guest(ctx, suffix+"-lone")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	amount := int64(7000)
	if _, err := s.Reg.Event(ctx, lone.ID, w.tenantB.ID, purchaseType, w.source.ID, suffix+"-lone", yesterday, &amount); err != nil {
		t.Fatalf("fixture purchase: %v", err)
	}
}

func lastWeek() url.Values {
	now := time.Now().UTC()
	return url.Values{"from": {now.AddDate(0, 0, -6).Format("2006-01-02")}, "to": {now.Format("2006-01-02")}}
}

func TestTenantAggregates_ReportsRepeatVisitsRevenueAndFirstPurchases(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "tenant-money")
	seedTenantPurchases(t, s, w, "tenant-money")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/tenants/"+w.tenantA.ID+"/aggregates", lastWeek())
	if err != nil {
		t.Fatalf("GET /api/v1/tenants/{id}/aggregates: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/tenants/{id}/aggregates with purchase metrics")
	s.RequireSchema("GET", "/api/v1/tenants/{id}/aggregates", resp)

	aggregate := jsonObject(t, resp, "tenant aggregate")
	for field, want := range map[string]int{
		"visitsCount": 10, "guestsCount": 10, "repeatVisitsCount": 5,
		"revenueTotalMinor": 15000, "avgCheckMinor": 3000, "firstPurchasesCount": 5,
	} {
		if got, ok := harness.IntField(aggregate, field); !ok || got != want {
			t.Fatalf("%s = %v, want %d: %s", field, aggregate[field], want, resp.Snippet())
		}
	}
	if aggregate["suppressed"] != false {
		t.Fatalf("every value has at least five guests behind it: %s", resp.Snippet())
	}
	purchaseDay := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	repeatDay := time.Now().UTC().AddDate(0, 0, -2).Format("2006-01-02")
	series, _ := aggregate["series"].([]any)
	for _, raw := range series {
		point, _ := raw.(map[string]any)
		day := harness.StringField(point, "date")
		revenue, hasRevenue := harness.IntField(point, "revenueTotalMinor")
		if day == purchaseDay && (!hasRevenue || revenue != 15000) {
			t.Fatalf("%s must show the revenue of five buyers: %v", day, point)
		}
		repeats, hasRepeats := harness.IntField(point, "repeatVisitsCount")
		if day == repeatDay && (!hasRepeats || repeats != 5) {
			t.Fatalf("%s must show five repeat visits: %v", day, point)
		}
	}

	lone, err := s.API.Get(ctx, "/api/v1/tenants/"+w.tenantB.ID+"/aggregates", lastWeek())
	if err != nil {
		t.Fatalf("GET aggregates of a tenant with one buyer: %v", err)
	}
	s.RequireStatus(lone, 200, "GET /api/v1/tenants/{id}/aggregates with one buyer")
	small := jsonObject(t, lone, "small tenant aggregate")
	for _, field := range []string{"revenueTotalMinor", "avgCheckMinor", "firstPurchasesCount", "visitsCount"} {
		if small[field] != nil {
			t.Fatalf("one buyer is below k = 5, %s must be null: %s", field, lone.Snippet())
		}
	}
	if small["suppressed"] != true {
		t.Fatalf("a suppressed aggregate must say so: %s", lone.Snippet())
	}
}
