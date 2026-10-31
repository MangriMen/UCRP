package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/service"
)

type tenantReferenceTestPolicy struct {
	allowed       bool
	changeVersion bool
	versions      int
}

func (p *tenantReferenceTestPolicy) Authorize(_ context.Context, principal domain.Principal, permission string, resource access.Resource) (access.Decision, error) {
	return access.Decision{Allowed: p.allowed && permission == "reference.read" && resource.TenantID == principal.TenantID}, nil
}

func (p *tenantReferenceTestPolicy) PolicyVersion(context.Context) (int64, error) {
	p.versions++
	if p.changeVersion && p.versions > 1 {
		return 2, nil
	}
	return 1, nil
}

type tenantReferenceTestReader struct {
	requested []string
}

func (r *tenantReferenceTestReader) GetTenant(_ context.Context, id string) (*service.TenantDTO, error) {
	r.requested = append(r.requested, id)
	return &service.TenantDTO{ID: id, Name: "Own tenant", Priority: 99}, nil
}

func TestTenantReferenceIsolation(t *testing.T) {
	for _, tc := range []struct {
		name          string
		path          string
		allowed       bool
		changeVersion bool
		want          int
		wantRequests  int
	}{
		{"own list", "/api/v1/tenants", true, false, http.StatusOK, 1},
		{"own list offset", "/api/v1/tenants?offset=1", true, false, http.StatusOK, 1},
		{"own details", "/api/v1/tenants/own", true, false, http.StatusOK, 1},
		{"foreign details", "/api/v1/tenants/other", true, false, http.StatusForbidden, 0},
		{"no grant list", "/api/v1/tenants", false, false, http.StatusForbidden, 0},
		{"no grant details", "/api/v1/tenants/own", false, false, http.StatusForbidden, 0},
		{"changed policy", "/api/v1/tenants/own", true, true, http.StatusConflict, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := &tenantReferenceTestPolicy{allowed: tc.allowed, changeVersion: tc.changeVersion}
			reader := &tenantReferenceTestReader{}
			handler := NewTenantReferenceHandler(policy, reader, nil, nil, nil)
			tokens := auth.NewManager("secret", time.Hour)
			token, err := tokens.Issue(domain.Principal{UserID: "u", Role: domain.RoleTenant, TenantID: "own"})
			if err != nil {
				t.Fatal(err)
			}
			router := chi.NewRouter()
			router.Use(auth.Middleware(tokens))
			router.Use(PermissionMiddleware(policy))
			router.Get("/api/v1/tenants", handler.List)
			router.Get("/api/v1/tenants/{id}", handler.Get)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+token.AccessToken)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != tc.want || len(reader.requested) != tc.wantRequests {
				t.Fatalf("got status %d, reads %v; want status %d and %d reads", response.Code, reader.requested, tc.want, tc.wantRequests)
			}
			if response.Code != http.StatusOK {
				return
			}
			var body map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if items, isList := body["items"].([]any); isList {
				if tc.path == "/api/v1/tenants?offset=1" {
					if len(items) != 0 {
						t.Fatalf("offset returned items: %v", items)
					}
					return
				}
				if len(items) != 1 {
					t.Fatalf("unexpected tenant list: %v", items)
				}
				body, _ = items[0].(map[string]any)
			}
			if body["id"] != "own" {
				t.Fatalf("unexpected tenant: %v", body)
			}
			if _, found := body["priority"]; found {
				t.Fatalf("priority leaked: %v", body)
			}
		})
	}
}
