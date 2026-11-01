package e2e

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestUserCreateByManager(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	manager := actor(t, s, "manager")
	email := s.Reg.Name("created-user") + "@e2e.local"
	resp, err := manager.Post(ctx, "/api/v1/users", map[string]any{
		"email":       email,
		"password":    "e2e-password-1",
		"displayName": "E2E created user",
		"role":        "analyst",
	})
	if err != nil {
		t.Fatalf("POST /api/v1/users: %v", err)
	}
	s.RequireStatus(resp, 201, "POST /api/v1/users as manager")
	s.RequireDeclaredStatus("POST", "/api/v1/users", resp)
	s.RequireSchema("POST", "/api/v1/users", resp)

	created := jsonObject(t, resp, "created user")
	if got := harness.StringField(created, "email"); got != email {
		t.Fatalf("created user email: expected %q, got %q", email, got)
	}
	if got := harness.StringField(created, "role"); got != "analyst" {
		t.Fatalf("created user role: expected analyst, got %q", got)
	}
	trackUser(s, harness.StringField(created, "id"))
}

func TestUserCreateTenantRequiresTenantID(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	manager := actor(t, s, "manager")
	resp, err := manager.Post(ctx, "/api/v1/users", map[string]any{
		"email":       s.Reg.Name("tenant-no-id") + "@e2e.local",
		"password":    "e2e-password-1",
		"displayName": "E2E tenant user",
		"role":        "tenant",
	})
	if err != nil {
		t.Fatalf("POST /api/v1/users: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/users with the tenant role but no tenant_id")
	s.RequireDeclaredStatus("POST", "/api/v1/users", resp)
	s.RequireSchema("POST", "/api/v1/users", resp)
}

func TestUserCreateDuplicateEmail(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	manager := actor(t, s, "manager")
	payload := map[string]any{
		"email":       s.Reg.Name("duplicate-user") + "@e2e.local",
		"password":    "e2e-password-1",
		"displayName": "E2E duplicate",
		"role":        "analyst",
	}

	first, err := manager.Post(ctx, "/api/v1/users", payload)
	if err != nil {
		t.Fatalf("POST /api/v1/users: %v", err)
	}
	s.RequireStatus(first, 201, "POST /api/v1/users first time")
	trackUser(s, harness.StringField(jsonObject(t, first, "created user"), "id"))

	second, err := manager.Post(ctx, "/api/v1/users", payload)
	if err != nil {
		t.Fatalf("POST /api/v1/users second time: %v", err)
	}
	s.RequireStatusOrDrift(second, 409, "POST /api/v1/users with a duplicate email")
	s.RequireDeclaredStatus("POST", "/api/v1/users", second)
}

func TestUserDeleteLastManagerReturns409(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	manager := actor(t, s, "manager")
	managers, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM app_user WHERE role = 'manager' AND is_active`)
	if err != nil {
		t.Fatalf("count active managers: %v", err)
	}
	if managers != 1 {
		t.Skipf("the fixture manager is not the only active manager (%d found); the rule is not isolated", managers)
	}

	self, err := manager.Me(ctx)
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me: %v", err)
	}
	s.RequireStatus(self, 200, "GET /api/v1/auth/me as manager")
	managerID := harness.StringField(jsonObject(t, self, "me"), "id")
	if managerID == "" {
		t.Skip("the profile payload does not expose the user id under the expected key")
	}

	resp, err := manager.Delete(ctx, "/api/v1/users/"+managerID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/users/{id}: %v", err)
	}
	s.RequireStatus(resp, 409, "DELETE /api/v1/users/{id} for the last active manager")
	s.RequireDeclaredStatus("DELETE", "/api/v1/users/{id}", resp)
}

func TestUserListAndGetContract(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	manager := actor(t, s, "manager")
	list, err := manager.Get(ctx, "/api/v1/users", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/users: %v", err)
	}
	s.RequireStatus(list, 200, "GET /api/v1/users as manager")
	s.RequireDeclaredStatus("GET", "/api/v1/users", list)
	s.RequireSchema("GET", "/api/v1/users", list)

	items, ok := harness.Items(mustJSON(t, list, "user list"))
	if !ok || len(items) == 0 {
		t.Fatalf("the user list must contain at least the fixture manager: %s", list.Snippet())
	}
	listed, _ := items[0].(map[string]any)
	id := harness.StringField(listed, "id")
	if id == "" {
		t.Fatalf("a listed user has no id: %s", list.Snippet())
	}

	get, err := manager.Get(ctx, "/api/v1/users/"+id, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/users/{id}: %v", err)
	}
	s.RequireStatus(get, 200, "GET /api/v1/users/{id} as manager")
	s.RequireDeclaredStatus("GET", "/api/v1/users/{id}", get)
	s.RequireSchema("GET", "/api/v1/users/{id}", get)
}

func TestUsersForbiddenForOtherRoles(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	for _, role := range []string{"analyst", "marketer"} {
		c := actor(t, s, role)
		list, err := c.Get(ctx, "/api/v1/users", nil)
		if err != nil {
			t.Fatalf("GET /api/v1/users as %s: %v", role, err)
		}
		s.RequireStatus(list, 403, "GET /api/v1/users as "+role)

		create, err := c.Post(ctx, "/api/v1/users", map[string]any{
			"email":       s.Reg.Name("forbidden-"+role) + "@e2e.local",
			"password":    "e2e-password-1",
			"displayName": "E2E forbidden " + role,
			"role":        "analyst",
		})
		if err != nil {
			t.Fatalf("POST /api/v1/users as %s: %v", role, err)
		}
		s.RequireStatus(create, 403, "POST /api/v1/users as "+role)
	}
}

func TestUserListFilters(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	token := s.Reg.NewID()[:8]
	analyst, err := s.Reg.User(ctx, token+"-analyst", "analyst", "")
	if err != nil {
		t.Fatalf("fixture analyst: %v", err)
	}
	dormant, err := s.Reg.User(ctx, token+"-dormant", "analyst", "")
	if err != nil {
		t.Fatalf("fixture dormant analyst: %v", err)
	}
	marketer, err := s.Reg.User(ctx, token+"-marketer", "marketer", "")
	if err != nil {
		t.Fatalf("fixture marketer: %v", err)
	}
	if err := s.DB.Exec(ctx, `UPDATE app_user SET is_active = FALSE WHERE id = $1`, dormant.ID); err != nil {
		t.Fatalf("deactivate fixture user: %v", err)
	}
	s.As(analyst.Email, analyst.Password)

	manager := actor(t, s, "manager")
	list := func(what string, query url.Values) (map[string]any, []any) {
		t.Helper()
		resp, err := manager.Get(ctx, "/api/v1/users", query)
		if err != nil {
			t.Fatalf("GET /api/v1/users (%s): %v", what, err)
		}
		s.RequireStatus(resp, 200, "GET /api/v1/users "+what)
		s.RequireSchema("GET", "/api/v1/users", resp)
		body := jsonObject(t, resp, what)
		items, ok := harness.Items(body)
		if !ok {
			t.Fatalf("%s: response has no items", what)
		}
		return body, items
	}
	total := func(body map[string]any) int {
		t.Helper()
		meta, _ := body["meta"].(map[string]any)
		value, ok := harness.IntField(meta, "total")
		if !ok {
			t.Fatalf("meta.total is missing: %v", body["meta"])
		}
		return value
	}

	body, items := list("by q", url.Values{"q": {strings.ToUpper(token)}})
	if len(items) != 3 || total(body) != 3 {
		t.Fatalf("q: expected 3 items and total 3, got %d items and total %d", len(items), total(body))
	}

	body, items = list("q and page_size", url.Values{"q": {token}, "page_size": {"1"}})
	if len(items) != 1 || total(body) != 3 {
		t.Fatalf("q with page_size=1: expected 1 item and total 3, got %d items and total %d", len(items), total(body))
	}

	body, items = list("q and role", url.Values{"q": {token}, "role": {"analyst"}})
	if len(items) != 2 || total(body) != 2 {
		t.Fatalf("role: expected 2 items and total 2, got %d items and total %d", len(items), total(body))
	}

	body, items = list("q and is_active=false", url.Values{"q": {token}, "is_active": {"false"}})
	if len(items) != 1 || total(body) != 1 || harness.StringField(items[0].(map[string]any), "id") != dormant.ID {
		t.Fatalf("is_active=false: expected only the deactivated user, got %v", items)
	}

	body, items = list("q and is_active=true", url.Values{"q": {token}, "is_active": {"true"}})
	if len(items) != 2 || total(body) != 2 {
		t.Fatalf("is_active=true: expected 2 items and total 2, got %d items and total %d", len(items), total(body))
	}

	_, items = list("by exact email", url.Values{"q": {analyst.Email}})
	if len(items) != 1 {
		t.Fatalf("q=email: expected 1 item, got %d", len(items))
	}
	if value, ok := items[0].(map[string]any)["lastLoginAt"].(string); !ok || value == "" {
		t.Fatalf("lastLoginAt must be set after a login, got %v", items[0].(map[string]any)["lastLoginAt"])
	}

	_, items = list("never logged in", url.Values{"q": {marketer.Email}})
	if len(items) != 1 || items[0].(map[string]any)["lastLoginAt"] != nil {
		t.Fatalf("lastLoginAt must be null for a user who never logged in, got %v", items)
	}

	resp, err := manager.Get(ctx, "/api/v1/users", url.Values{"is_active": {"maybe"}})
	if err != nil {
		t.Fatalf("GET /api/v1/users with invalid is_active: %v", err)
	}
	s.RequireStatus(resp, 400, "GET /api/v1/users with is_active=maybe")
	s.RequireDeclaredStatus("GET", "/api/v1/users", resp)
	s.RequireSchema("GET", "/api/v1/users", resp)
	resp, err = manager.Get(ctx, "/api/v1/users", url.Values{"role": {"bogus"}})
	if err != nil {
		t.Fatalf("GET /api/v1/users with an unknown role: %v", err)
	}
	s.RequireStatus(resp, 400, "GET /api/v1/users with role=bogus")
}

func trackUser(s *harness.Suite, userID string) {
	s.T.Helper()
	if userID == "" {
		return
	}
	s.T.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
		defer cancel()
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM app_user WHERE id = $1`, userID)
	})
}
