package e2e

import (
	"context"
	"testing"

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
	code := s.Reg.Code("api-tenant")
	resp, err := s.API.Post(ctx, "/api/v1/tenants", map[string]any{
		"code":        code,
		"name":        "E2E API tenant",
		"category_id": category.ID,
		"priority":    70,
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
	payload := map[string]any{
		"code":        s.Reg.Code("dup-tenant"),
		"name":        "E2E duplicate tenant",
		"category_id": category.ID,
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

	replace, err := s.API.Put(ctx, "/api/v1/tenant-categories/"+primary.ID+"/relations", map[string]any{
		"items": []any{
			map[string]any{"related_category_id": related.ID, "relation_type": "adjacent", "weight_pct": 25},
		},
	})
	if err != nil {
		t.Fatalf("PUT /api/v1/tenant-categories/{id}/relations: %v", err)
	}
	s.RequireStatus(replace, 200, "PUT /api/v1/tenant-categories/{id}/relations")
	s.RequireDeclaredStatus("PUT", "/api/v1/tenant-categories/{id}/relations", replace)

	items, ok := harness.Items(mustJSON(t, replace, "replaced relations"))
	if !ok || len(items) != 1 {
		t.Fatalf("the relation set must contain exactly one entry: %s", replace.Snippet())
	}
	entry, _ := items[0].(map[string]any)
	if got := harness.StringField(entry, "relation_type"); got != "adjacent" {
		t.Fatalf("relation_type: expected adjacent, got %q", got)
	}
	if got, _ := harness.IntField(entry, "weight_pct"); got != 25 {
		t.Fatalf("weight_pct: expected 25, got %d", got)
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
		"items": []any{
			map[string]any{"related_category_id": related.ID, "relation_type": "bogus", "weight_pct": 10},
		},
	})
	if err != nil {
		t.Fatalf("PUT /api/v1/tenant-categories/{id}/relations: %v", err)
	}
	s.RequireStatus(resp, 400, "PUT relations with an unknown relation_type")
	s.RequireDeclaredStatus("PUT", "/api/v1/tenant-categories/{id}/relations", resp)
	s.RequireSchema("PUT", "/api/v1/tenant-categories/{id}/relations", resp)
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
		"tenant_id": tenant.ID,
		"code":      code,
		"name":      "E2E project",
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
		"code":               code,
		"name":               "E2E top up",
		"has_amount":         true,
		"counts_as_presence": false,
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
	if presence, ok := created["counts_as_presence"].(bool); !ok || presence {
		t.Fatalf("counts_as_presence must stay false for a top-up type, got %v", created["counts_as_presence"])
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
