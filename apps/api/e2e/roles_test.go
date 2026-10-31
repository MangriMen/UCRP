package e2e

import (
	"testing"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/harness"
)

type roleCase struct {
	name   string
	role   string
	method string
	path   string
	body   any
	want   int
}

func TestRoleMatrix(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	marketer, err := s.Reg.User(ctx, "marketer", "marketer", "")
	if err != nil {
		t.Fatalf("fixture marketer: %v", err)
	}
	manager, err := s.Reg.User(ctx, "manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}
	analyst, err := s.Reg.User(ctx, "analyst", "analyst", "")
	if err != nil {
		t.Fatalf("fixture analyst: %v", err)
	}

	clients := map[string]func() *client.Client{
		"marketer": func() *client.Client { return s.As(marketer.Email, marketer.Password) },
		"manager":  func() *client.Client { return s.As(manager.Email, manager.Password) },
		"analyst":  func() *client.Client { return s.As(analyst.Email, analyst.Password) },
	}

	cases := []roleCase{
		{"marketer reads guests", "marketer", "GET", "/api/v1/guests?q=e2e", nil, 200},
		{"marketer reads segments", "marketer", "GET", "/api/v1/segments", nil, 200},
		{"marketer reads merge queue", "marketer", "GET", "/api/v1/merge-candidates", nil, 200},
		{"marketer reads exports", "marketer", "GET", "/api/v1/exports", nil, 200},
		{"marketer cannot read imports", "marketer", "GET", "/api/v1/imports", nil, 403},
		{"marketer cannot create imports", "marketer", "POST", "/api/v1/imports", nil, 403},
		{"marketer cannot confirm merges", "marketer", "POST", "/api/v1/merge-candidates/00000000-0000-0000-0000-000000000000/confirm", nil, 403},
		{"marketer cannot create data sources", "marketer", "POST", "/api/v1/data-sources", map[string]any{}, 403},

		{"manager reads imports", "manager", "GET", "/api/v1/imports", nil, 200},
		{"manager reads data sources", "manager", "GET", "/api/v1/data-sources", nil, 200},
		{"manager cannot create imports", "manager", "POST", "/api/v1/imports", nil, 403},
		{"manager cannot read exports", "manager", "GET", "/api/v1/exports", nil, 403},
		{"manager cannot read guests", "manager", "GET", "/api/v1/guests?q=e2e", nil, 403},
		{"manager cannot confirm merges", "manager", "POST", "/api/v1/merge-candidates/00000000-0000-0000-0000-000000000000/confirm", nil, 403},

		{"analyst reads guests", "analyst", "GET", "/api/v1/guests?q=e2e", nil, 200},
		{"analyst reads imports", "analyst", "GET", "/api/v1/imports", nil, 200},
		{"analyst reads data sources", "analyst", "GET", "/api/v1/data-sources", nil, 200},
		{"analyst reads data source types", "analyst", "GET", "/api/v1/data-source-types", nil, 200},
		{"analyst cannot read exports", "analyst", "GET", "/api/v1/exports", nil, 403},
		{"analyst cannot create exports", "analyst", "POST", "/api/v1/exports", map[string]any{}, 403},
		{"analyst cannot preview exports", "analyst", "POST", "/api/v1/exports/preview", map[string]any{}, 403},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := clients[tc.role]()
			var resp *client.Response
			var err error
			if tc.method == "GET" {
				resp, err = c.Get(ctx, tc.path, nil)
			} else {
				resp, err = c.Post(ctx, tc.path, tc.body)
			}
			if err != nil {
				t.Fatalf("%s %s: %v", tc.method, tc.path, err)
			}
			if resp.Status != tc.want {
				t.Fatalf("%s %s as %s: expected HTTP %d, got %d (body=%s)",
					tc.method, tc.path, tc.role, tc.want, resp.Status, resp.Snippet())
			}
		})
	}
}

func TestRoleTenantCannotReachPrivateAPI(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "role-tenant")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	tenant, err := s.Reg.Tenant(ctx, "role-tenant", category.ID, 50)
	if err != nil {
		t.Fatalf("fixture tenant: %v", err)
	}
	tenantUser, err := s.Reg.User(ctx, "tenant", "tenant", tenant.ID)
	if err != nil {
		t.Fatalf("fixture tenant user: %v", err)
	}

	c := s.As(tenantUser.Email, tenantUser.Password)

	me, err := c.Me(ctx)
	if err != nil {
		t.Fatalf("GET /auth/me as tenant: %v", err)
	}
	s.RequireStatus(me, 200, "GET /api/v1/auth/me as tenant")

	for _, path := range []string{
		"/api/v1/guests?q=e2e",
		"/api/v1/segments",
		"/api/v1/merge-candidates",
		"/api/v1/imports",
		"/api/v1/exports",
		"/api/v1/data-sources",
	} {
		resp, err := c.Get(ctx, path, nil)
		if err != nil {
			t.Fatalf("GET %s as tenant: %v", path, err)
		}
		if resp.Status != 403 {
			t.Fatalf("GET %s as tenant: expected HTTP 403, got %d (body=%s)", path, resp.Status, resp.Snippet())
		}
	}
}
