package e2e

import (
	"context"
	"net/url"
	"testing"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestTenantCreateAndPatch(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "tenant-create")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	mall, err := s.Reg.Mall(ctx, "tenant-create")
	if err != nil {
		t.Fatalf("fixture mall: %v", err)
	}
	code := s.Reg.Code("api-tenant")
	resp, err := s.API.Post(ctx, "/api/v1/tenants", map[string]any{
		"code":       code,
		"name":       "E2E API tenant",
		"mallId":     mall.ID,
		"categoryId": category.ID,
		"priority":   70,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/tenants: %v", err)
	}
	s.RequireStatus(resp, 201, "POST /api/v1/tenants")
	s.RequireDeclaredStatus("POST", "/api/v1/tenants", resp)
	s.RequireSchema("POST", "/api/v1/tenants", resp)

	created := jsonObject(t, resp, "created tenant")
	tenantID := harness.StringField(created, "id")
	if tenantID == "" {
		t.Fatalf("created tenant has no id: %s", resp.Snippet())
	}
	if got := harness.StringField(created, "code"); got != code {
		t.Fatalf("created tenant code: expected %q, got %q", code, got)
	}
	trackTenant(s, tenantID)

	patch, err := s.API.Patch(ctx, "/api/v1/tenants/"+tenantID, map[string]any{"priority": 15})
	if err != nil {
		t.Fatalf("PATCH /api/v1/tenants/{id}: %v", err)
	}
	s.RequireStatus(patch, 200, "PATCH /api/v1/tenants/{id}")
	s.RequireDeclaredStatus("PATCH", "/api/v1/tenants/{id}", patch)
	if got, _ := harness.IntField(jsonObject(t, patch, "patched tenant"), "priority"); got != 15 {
		t.Fatalf("patched priority: expected 15, got %d", got)
	}
}

func TestTenantCreateDuplicateCode(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "tenant-duplicate")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	mall, err := s.Reg.Mall(ctx, "tenant-duplicate")
	if err != nil {
		t.Fatalf("fixture mall: %v", err)
	}
	payload := map[string]any{
		"code":       s.Reg.Code("dup-tenant"),
		"name":       "E2E duplicate tenant",
		"mallId":     mall.ID,
		"categoryId": category.ID,
	}

	first, err := s.API.Post(ctx, "/api/v1/tenants", payload)
	if err != nil {
		t.Fatalf("POST /api/v1/tenants: %v", err)
	}
	s.RequireStatus(first, 201, "POST /api/v1/tenants first time")
	trackTenant(s, harness.StringField(jsonObject(t, first, "created tenant"), "id"))

	second, err := s.API.Post(ctx, "/api/v1/tenants", payload)
	if err != nil {
		t.Fatalf("POST /api/v1/tenants second time: %v", err)
	}
	s.RequireStatus(second, 409, "POST /api/v1/tenants with a duplicate code")
	s.RequireDeclaredStatus("POST", "/api/v1/tenants", second)
}

func TestTenantDeleteWithEventsReturns409(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	w := buildWorld(t, s, "tenant-delete")
	guest, err := s.Reg.Guest(ctx, "tenant-delete")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	w.visit(t, s, guest.ID, w.tenantA.ID, "tenant-delete", 3)

	resp, err := s.API.Delete(ctx, "/api/v1/tenants/"+w.tenantA.ID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/tenants/{id}: %v", err)
	}
	s.RequireStatus(resp, 409, "DELETE /api/v1/tenants/{id} for a tenant that owns events")
	s.RequireDeclaredStatus("DELETE", "/api/v1/tenants/{id}", resp)
}

func TestTenantDeleteWithoutEvents(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "tenant-clean")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	tenant, err := s.Reg.Tenant(ctx, "tenant-clean", category.ID, 40)
	if err != nil {
		t.Fatalf("fixture tenant: %v", err)
	}

	resp, err := s.API.Delete(ctx, "/api/v1/tenants/"+tenant.ID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/tenants/{id}: %v", err)
	}
	s.RequireStatus(resp, 204, "DELETE /api/v1/tenants/{id} for a tenant without events")
	s.RequireDeclaredStatus("DELETE", "/api/v1/tenants/{id}", resp)

	get, err := s.API.Get(ctx, "/api/v1/tenants/"+tenant.ID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/tenants/{id} after deletion: %v", err)
	}
	s.RequireStatus(get, 404, "GET /api/v1/tenants/{id} after the soft delete")
}

func TestTenantDeleteMarketerForbidden(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "tenant-role")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	tenant, err := s.Reg.Tenant(ctx, "tenant-role", category.ID, 40)
	if err != nil {
		t.Fatalf("fixture tenant: %v", err)
	}

	marketer := actor(t, s, "marketer")
	resp, err := marketer.Delete(ctx, "/api/v1/tenants/"+tenant.ID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/tenants/{id} as marketer: %v", err)
	}
	s.RequireStatus(resp, 403, "DELETE /api/v1/tenants/{id} as marketer")

	patch, err := marketer.Patch(ctx, "/api/v1/tenants/"+tenant.ID, map[string]any{"priority": 10})
	if err != nil {
		t.Fatalf("PATCH /api/v1/tenants/{id} as marketer: %v", err)
	}
	s.RequireStatus(patch, 403, "PATCH /api/v1/tenants/{id} as marketer")
}

func TestCategoryRelationsReplace(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	primary, err := s.Reg.Category(ctx, "relations-primary")
	if err != nil {
		t.Fatalf("fixture primary category: %v", err)
	}
	related, err := s.Reg.Category(ctx, "relations-related")
	if err != nil {
		t.Fatalf("fixture related category: %v", err)
	}
	clearCategoryRelations(s, primary.ID)

	replace, err := s.API.Put(ctx, "/api/v1/tenant-categories/"+primary.ID+"/relations", map[string]any{
		"relations": []any{
			map[string]any{"categoryId": primary.ID, "relatedCategoryId": related.ID, "relationType": "adjacent", "weightPct": 25},
		},
	})
	if err != nil {
		t.Fatalf("PUT /api/v1/tenant-categories/{id}/relations: %v", err)
	}
	s.RequireStatus(replace, 200, "PUT /api/v1/tenant-categories/{id}/relations")
	s.RequireDeclaredStatus("PUT", "/api/v1/tenant-categories/{id}/relations", replace)

	replaceBody := mustJSON(t, replace, "replaced relations")
	replaceObj, ok := replaceBody.(map[string]any)
	if !ok {
		t.Fatalf("replaced relations is not an object: %s", replace.Snippet())
	}
	if cat := harness.StringField(replaceObj, "categoryId"); cat != primary.ID {
		t.Fatalf("categoryId: expected %q, got %q", primary.ID, cat)
	}
	items, ok := replaceObj["relations"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("the relation set must contain exactly one entry: %s", replace.Snippet())
	}
	entry, _ := items[0].(map[string]any)
	if got := harness.StringField(entry, "relationType"); got != "adjacent" {
		t.Fatalf("relationType: expected adjacent, got %q", got)
	}
	if got, _ := harness.IntField(entry, "weightPct"); got != 25 {
		t.Fatalf("weightPct: expected 25, got %d", got)
	}

	list, err := s.API.Get(ctx, "/api/v1/tenant-categories/"+primary.ID+"/relations", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/tenant-categories/{id}/relations: %v", err)
	}
	s.RequireStatus(list, 200, "GET /api/v1/tenant-categories/{id}/relations")
	s.RequireDeclaredStatus("GET", "/api/v1/tenant-categories/{id}/relations", list)
	s.RequireSchema("GET", "/api/v1/tenant-categories/{id}/relations", list)
}

func TestCategoryRelationsRejectUnknownType(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	primary, err := s.Reg.Category(ctx, "relations-bad")
	if err != nil {
		t.Fatalf("fixture primary category: %v", err)
	}
	related, err := s.Reg.Category(ctx, "relations-bad-related")
	if err != nil {
		t.Fatalf("fixture related category: %v", err)
	}

	resp, err := s.API.Put(ctx, "/api/v1/tenant-categories/"+primary.ID+"/relations", map[string]any{
		"relations": []any{
			map[string]any{"categoryId": primary.ID, "relatedCategoryId": related.ID, "relationType": "bogus", "weightPct": 10},
		},
	})
	if err != nil {
		t.Fatalf("PUT /api/v1/tenant-categories/{id}/relations: %v", err)
	}
	s.RequireStatus(resp, 400, "PUT relations with an unknown relation_type")
	s.RequireDeclaredStatus("PUT", "/api/v1/tenant-categories/{id}/relations", resp)
	s.RequireSchema("PUT", "/api/v1/tenant-categories/{id}/relations", resp)
}

func relationsPath(categoryID string) string {
	return "/api/v1/tenant-categories/" + categoryID + "/relations"
}

func clearCategoryRelations(s *harness.Suite, categoryIDs ...string) {
	s.T.Cleanup(func() {
		for _, id := range categoryIDs {
			if err := s.DB.Exec(context.Background(), `
				DELETE FROM tenant_category_relation WHERE category_id = $1 OR related_category_id = $1`, id); err != nil {
				s.T.Logf("cleanup relations of %s: %v", id, err)
			}
		}
	})
}

func relationEntries(t *testing.T, resp *client.Response, what string) []map[string]any {
	t.Helper()
	items, ok := jsonObject(t, resp, what)["relations"].([]any)
	if !ok {
		t.Fatalf("%s: relations is not an array: %s", what, resp.Snippet())
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry, _ := item.(map[string]any)
		out = append(out, entry)
	}
	return out
}

func TestCategoryRelationsAreSymmetric(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	cinema, err := s.Reg.Category(ctx, "relations-sym-cinema")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	food, err := s.Reg.Category(ctx, "relations-sym-food")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	clearCategoryRelations(s, cinema.ID, food.ID)

	put, err := s.API.Put(ctx, relationsPath(cinema.ID), map[string]any{
		"relations": []any{
			map[string]any{"categoryId": cinema.ID, "relatedCategoryId": food.ID, "relationType": "adjacent", "weightPct": 30},
		},
	})
	if err != nil {
		t.Fatalf("PUT relations: %v", err)
	}
	s.RequireStatus(put, 200, "PUT relations of the cinema category")
	s.RequireSchema("PUT", "/api/v1/tenant-categories/{id}/relations", put)

	reverse, err := s.API.Get(ctx, relationsPath(food.ID), nil)
	if err != nil {
		t.Fatalf("GET relations: %v", err)
	}
	s.RequireStatus(reverse, 200, "GET relations of the related category")
	s.RequireSchema("GET", "/api/v1/tenant-categories/{id}/relations", reverse)
	entries := relationEntries(t, reverse, "reverse relations")
	if len(entries) != 1 {
		t.Fatalf("the related category must see the relation once: %s", reverse.Snippet())
	}
	entry := entries[0]
	if got := harness.StringField(entry, "categoryId"); got != food.ID {
		t.Fatalf("categoryId: expected the path category %q, got %q", food.ID, got)
	}
	if got := harness.StringField(entry, "relatedCategoryId"); got != cinema.ID {
		t.Fatalf("relatedCategoryId: expected %q, got %q", cinema.ID, got)
	}
	if got := harness.StringField(entry, "relatedCategoryName"); got != cinema.Name {
		t.Fatalf("relatedCategoryName: expected %q, got %q", cinema.Name, got)
	}
	if got, _ := harness.IntField(entry, "weightPct"); got != 30 {
		t.Fatalf("weightPct: expected 30, got %d", got)
	}

	cleared, err := s.API.Put(ctx, relationsPath(food.ID), map[string]any{"relations": []any{}})
	if err != nil {
		t.Fatalf("PUT empty relations: %v", err)
	}
	s.RequireStatus(cleared, 200, "PUT an empty relation set on the related category")

	forward, err := s.API.Get(ctx, relationsPath(cinema.ID), nil)
	if err != nil {
		t.Fatalf("GET relations: %v", err)
	}
	s.RequireStatus(forward, 200, "GET relations after clearing the reverse side")
	if entries := relationEntries(t, forward, "forward relations"); len(entries) != 0 {
		t.Fatalf("clearing one side must remove the pair: %s", forward.Snippet())
	}
}

func TestCategoryRelationsRejectInvalidSet(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	primary, err := s.Reg.Category(ctx, "relations-invalid")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	related, err := s.Reg.Category(ctx, "relations-invalid-related")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	inactive, err := s.Reg.Category(ctx, "relations-invalid-inactive")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	if err := s.DB.Exec(ctx, `UPDATE tenant_category SET is_active = FALSE WHERE id = $1`, inactive.ID); err != nil {
		t.Fatalf("deactivate category: %v", err)
	}
	clearCategoryRelations(s, primary.ID)

	relation := func(relatedID, relationType string, weight any) map[string]any {
		return map[string]any{"categoryId": primary.ID, "relatedCategoryId": relatedID, "relationType": relationType, "weightPct": weight}
	}
	cases := []struct {
		name      string
		relations []any
	}{
		{"self relation", []any{relation(primary.ID, "adjacent", 50)}},
		{"duplicate pair", []any{relation(related.ID, "adjacent", 50), relation(related.ID, "adjacent", 60)}},
		{"pair with both types", []any{relation(related.ID, "adjacent", 50), relation(related.ID, "competitor", nil)}},
		{"unknown related category", []any{relation(s.Reg.NewID(), "adjacent", 50)}},
		{"inactive related category", []any{relation(inactive.ID, "adjacent", 50)}},
		{"weight above 100", []any{relation(related.ID, "adjacent", 101)}},
	}
	for _, tc := range cases {
		resp, err := s.API.Put(ctx, relationsPath(primary.ID), map[string]any{"relations": tc.relations})
		if err != nil {
			t.Fatalf("PUT relations (%s): %v", tc.name, err)
		}
		s.RequireStatus(resp, 400, "PUT relations with "+tc.name)
		s.RequireDeclaredStatus("PUT", "/api/v1/tenant-categories/{id}/relations", resp)
		s.RequireSchema("PUT", "/api/v1/tenant-categories/{id}/relations", resp)
	}
}

func TestCategoryRelationsUnknownCategoryReturns404(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	missing := s.Reg.NewID()
	get, err := s.API.Get(ctx, relationsPath(missing), nil)
	if err != nil {
		t.Fatalf("GET relations: %v", err)
	}
	s.RequireStatus(get, 404, "GET relations of an unknown category")
	s.RequireDeclaredStatus("GET", "/api/v1/tenant-categories/{id}/relations", get)

	put, err := s.API.Put(ctx, relationsPath(missing), map[string]any{"relations": []any{}})
	if err != nil {
		t.Fatalf("PUT relations: %v", err)
	}
	s.RequireStatus(put, 404, "PUT relations of an unknown category")
	s.RequireDeclaredStatus("PUT", "/api/v1/tenant-categories/{id}/relations", put)
}

func TestCategoryRelationsMarketerReadsButCannotReplace(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "relations-role")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	marketer := actor(t, s, "marketer")

	get, err := marketer.Get(ctx, relationsPath(category.ID), nil)
	if err != nil {
		t.Fatalf("GET relations as marketer: %v", err)
	}
	s.RequireStatus(get, 200, "GET relations as marketer")

	put, err := marketer.Put(ctx, relationsPath(category.ID), map[string]any{"relations": []any{}})
	if err != nil {
		t.Fatalf("PUT relations as marketer: %v", err)
	}
	s.RequireStatus(put, 403, "PUT relations as marketer")
	s.RequireDeclaredStatus("PUT", "/api/v1/tenant-categories/{id}/relations", put)
}

func TestCategoryDeleteWithRelationsReturns409(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	primary, err := s.Reg.Category(ctx, "relations-delete")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	related, err := s.Reg.Category(ctx, "relations-delete-related")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	clearCategoryRelations(s, primary.ID)

	put, err := s.API.Put(ctx, relationsPath(primary.ID), map[string]any{
		"relations": []any{
			map[string]any{"categoryId": primary.ID, "relatedCategoryId": related.ID, "relationType": "competitor"},
		},
	})
	if err != nil {
		t.Fatalf("PUT relations: %v", err)
	}
	s.RequireStatus(put, 200, "PUT relations before deletion")

	resp, err := s.API.Delete(ctx, "/api/v1/tenant-categories/"+related.ID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/tenant-categories/{id}: %v", err)
	}
	s.RequireStatus(resp, 409, "DELETE a category used by a relation")
	s.RequireDeclaredStatus("DELETE", "/api/v1/tenant-categories/{id}", resp)
	if got := resp.ErrorCode(); got != "category_in_use" {
		t.Fatalf("error code: expected category_in_use, got %q", got)
	}
}

func TestProjectCrud(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "project-crud")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	tenant, err := s.Reg.Tenant(ctx, "project-crud", category.ID, 50)
	if err != nil {
		t.Fatalf("fixture tenant: %v", err)
	}
	code := s.Reg.Code("project")

	create, err := s.API.Post(ctx, "/api/v1/projects", map[string]any{
		"tenantId": tenant.ID,
		"code":     code,
		"name":     "E2E project",
	})
	if err != nil {
		t.Fatalf("POST /api/v1/projects: %v", err)
	}
	s.RequireStatus(create, 201, "POST /api/v1/projects")
	s.RequireDeclaredStatus("POST", "/api/v1/projects", create)

	created := jsonObject(t, create, "created project")
	projectID := harness.StringField(created, "id")
	if projectID == "" {
		t.Fatalf("created project has no id: %s", create.Snippet())
	}
	trackProject(s, projectID)

	list, err := s.API.Get(ctx, "/api/v1/projects", map[string][]string{"tenant_id": {tenant.ID}})
	if err != nil {
		t.Fatalf("GET /api/v1/projects: %v", err)
	}
	s.RequireStatus(list, 200, "GET /api/v1/projects")
	s.RequireDeclaredStatus("GET", "/api/v1/projects", list)
	listed, _ := harness.Items(mustJSON(t, list, "project list"))
	if !containsID(listed, projectID) {
		t.Fatalf("the created project %s is absent from the filtered list", projectID)
	}

	update, err := s.API.Patch(ctx, "/api/v1/projects/"+projectID, map[string]any{"name": "E2E project renamed"})
	if err != nil {
		t.Fatalf("PATCH /api/v1/projects/{id}: %v", err)
	}
	s.RequireStatus(update, 200, "PATCH /api/v1/projects/{id}")
	s.RequireDeclaredStatus("PATCH", "/api/v1/projects/{id}", update)

	del, err := s.API.Delete(ctx, "/api/v1/projects/"+projectID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/projects/{id}: %v", err)
	}
	s.RequireStatus(del, 204, "DELETE /api/v1/projects/{id}")
	s.RequireDeclaredStatus("DELETE", "/api/v1/projects/{id}", del)
}

func TestEventTypeCountsAsPresence(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	code := s.Reg.Code("api-evt")
	create, err := s.API.Post(ctx, "/api/v1/event-types", map[string]any{
		"code":             code,
		"name":             "E2E top up",
		"hasAmount":        true,
		"countsAsPresence": false,
	})
	if err != nil {
		t.Fatalf("POST /api/v1/event-types: %v", err)
	}
	s.RequireStatus(create, 201, "POST /api/v1/event-types")
	s.RequireDeclaredStatus("POST", "/api/v1/event-types", create)
	s.RequireSchema("POST", "/api/v1/event-types", create)

	created := jsonObject(t, create, "created event type")
	eventTypeID := harness.StringField(created, "id")
	if eventTypeID == "" {
		t.Fatalf("created event type has no id: %s", create.Snippet())
	}
	if presence, ok := created["countsAsPresence"].(bool); !ok || presence {
		t.Fatalf("countsAsPresence must stay false for a top-up type, got %v", created["countsAsPresence"])
	}
	trackEventType(s, eventTypeID)

	list, err := s.API.Get(ctx, "/api/v1/event-types", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/event-types: %v", err)
	}
	s.RequireStatus(list, 200, "GET /api/v1/event-types")
	s.RequireDeclaredStatus("GET", "/api/v1/event-types", list)
	items, _ := harness.Items(mustJSON(t, list, "event type list"))
	if !containsID(items, eventTypeID) {
		t.Fatalf("the created event type %s is absent from the catalogue", eventTypeID)
	}
}

func TestEventTypeRequiresCodeAndName(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/event-types", map[string]any{
		"code": "",
		"name": "",
	})
	if err != nil {
		t.Fatalf("POST /api/v1/event-types: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/event-types without code and name")
	s.RequireDeclaredStatus("POST", "/api/v1/event-types", resp)
	s.RequireSchema("POST", "/api/v1/event-types", resp)
}

func trackTenant(s *harness.Suite, tenantID string) {
	s.T.Helper()
	if tenantID == "" {
		return
	}
	s.T.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cancel()
		_ = s.DB.Exec(ctx, `DELETE FROM tenant WHERE id = $1`, tenantID)
	})
}

func trackProject(s *harness.Suite, projectID string) {
	s.T.Helper()
	if projectID == "" {
		return
	}
	s.T.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cancel()
		_ = s.DB.Exec(ctx, `DELETE FROM project WHERE id = $1`, projectID)
	})
}

func trackEventType(s *harness.Suite, eventTypeID string) {
	s.T.Helper()
	if eventTypeID == "" {
		return
	}
	s.T.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cancel()
		_ = s.DB.Exec(ctx, `DELETE FROM event_type WHERE id = $1`, eventTypeID)
	})
}

func findCategory(t *testing.T, resp *client.Response, id string) map[string]any {
	t.Helper()
	items, ok := mustJSON(t, resp, "categories").([]any)
	if !ok {
		t.Fatalf("categories: expected an array, got %s", resp.Snippet())
	}
	for _, item := range items {
		if category, _ := item.(map[string]any); harness.StringField(category, "id") == id {
			return category
		}
	}
	return nil
}

func TestCategoryDescriptionAndInactiveListing(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "inactive-listing")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	path := "/api/v1/tenant-categories/" + category.ID

	patch, err := s.API.Patch(ctx, path, map[string]any{"description": "  Кафе и рестораны  ", "isActive": false})
	if err != nil {
		t.Fatalf("PATCH category: %v", err)
	}
	s.RequireStatus(patch, 200, "PATCH category description and activity")
	s.RequireSchema("PATCH", path, patch)
	patched := jsonObject(t, patch, "patched category")
	if patched["description"] != "Кафе и рестораны" || patched["isActive"] != false {
		t.Fatalf("patched category: %s", patch.Snippet())
	}

	active, err := s.API.Get(ctx, "/api/v1/tenant-categories", nil)
	if err != nil {
		t.Fatalf("GET categories: %v", err)
	}
	s.RequireStatus(active, 200, "GET categories")
	s.RequireSchema("GET", "/api/v1/tenant-categories", active)
	if findCategory(t, active, category.ID) != nil {
		t.Fatalf("inactive category must not be listed by default")
	}

	all, err := s.API.Get(ctx, "/api/v1/tenant-categories", url.Values{"include_inactive": {"true"}})
	if err != nil {
		t.Fatalf("GET categories with inactive: %v", err)
	}
	s.RequireStatus(all, 200, "GET categories with inactive")
	s.RequireSchema("GET", "/api/v1/tenant-categories", all)
	listed := findCategory(t, all, category.ID)
	if listed == nil || listed["isActive"] != false || listed["description"] != "Кафе и рестораны" {
		t.Fatalf("inactive category with description must be listed: %s", all.Snippet())
	}

	reactivate, err := s.API.Patch(ctx, path, map[string]any{"isActive": true})
	if err != nil {
		t.Fatalf("PATCH category activity: %v", err)
	}
	s.RequireStatus(reactivate, 200, "PATCH category activity")
	if reactivated := jsonObject(t, reactivate, "reactivated category"); reactivated["description"] != "Кафе и рестораны" || reactivated["isActive"] != true {
		t.Fatalf("reactivation must keep the description: %s", reactivate.Snippet())
	}

	invalid, err := s.API.Get(ctx, "/api/v1/tenant-categories", url.Values{"include_inactive": {"maybe"}})
	if err != nil {
		t.Fatalf("GET categories with a bad flag: %v", err)
	}
	s.RequireStatus(invalid, 400, "GET categories with a bad include_inactive")
}
