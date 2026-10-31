package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
)

type permissionTestAuthorizer struct {
	denied        string
	changeVersion bool
	versionCalls  int
}

func (a *permissionTestAuthorizer) Authorize(_ context.Context, _ domain.Principal, permission string, _ access.Resource) (access.Decision, error) {
	return access.Decision{Allowed: permission != a.denied, KThreshold: 5}, nil
}

func (a *permissionTestAuthorizer) PolicyVersion(context.Context) (int64, error) {
	a.versionCalls++
	if a.changeVersion && a.versionCalls > 1 {
		return 2, nil
	}
	return 1, nil
}

func TestPermissionMiddleware(t *testing.T) {
	cases := []struct {
		name          string
		method        string
		path          string
		role          domain.Role
		denied        string
		changeVersion bool
		want          int
	}{
		{"profile permitted", http.MethodGet, "/api/v1/guests/guest-1", domain.RoleManager, "", false, http.StatusNoContent},
		{"profile field denied", http.MethodGet, "/api/v1/guests/guest-1", domain.RoleManager, "field.guest.phone.read", false, http.StatusForbidden},
		{"search identifier denied", http.MethodGet, "/api/v1/guests?q=contact", domain.RoleManager, "guest.identifier.read", false, http.StatusForbidden},
		{"search phone denied", http.MethodGet, "/api/v1/guests?q=contact", domain.RoleManager, "field.guest.phone.read", false, http.StatusForbidden},
		{"search email denied", http.MethodGet, "/api/v1/guests?q=contact", domain.RoleManager, "field.guest.email.read", false, http.StatusForbidden},
		{"search card denied", http.MethodGet, "/api/v1/guests?q=contact", domain.RoleManager, "field.guest.card.read", false, http.StatusForbidden},
		{"events field denied", http.MethodGet, "/api/v1/guests/guest-1/events", domain.RoleManager, "field.event.amount.read", false, http.StatusForbidden},
		{"tenant profile delegated", http.MethodGet, "/api/v1/guests/guest-1", domain.RoleTenant, "guest.profile.read", false, http.StatusNoContent},
		{"tenant events delegated", http.MethodGet, "/api/v1/guests/guest-1/events", domain.RoleTenant, "guest.event.read", false, http.StatusNoContent},
		{"policy changed", http.MethodGet, "/api/v1/guests/guest-1", domain.RoleManager, "", true, http.StatusForbidden},
		{"unknown route", http.MethodGet, "/api/v1/unknown", domain.RoleManager, "", false, http.StatusForbidden},
		{"user creation permitted", http.MethodPost, "/api/v1/users", domain.RoleManager, "", false, http.StatusNoContent},
		{"user update permitted", http.MethodPatch, "/api/v1/users/user-1", domain.RoleManager, "", false, http.StatusNoContent},
		{"user update denied", http.MethodPatch, "/api/v1/users/user-1", domain.RoleManager, "user.write", false, http.StatusForbidden},
		{"cross-tenant permitted", http.MethodGet, "/api/v1/dashboard/cross-tenant", domain.RoleManager, "", false, http.StatusNoContent},
		{"cross-tenant metric denied", http.MethodGet, "/api/v1/dashboard/cross-tenant", domain.RoleManager, "metric.cross_tenant_distribution.read", false, http.StatusForbidden},
		{"dashboard metrics permitted", http.MethodGet, "/api/v1/dashboard/metrics", domain.RoleManager, "", false, http.StatusNoContent},
		{"dashboard metric denied", http.MethodGet, "/api/v1/dashboard/metrics", domain.RoleManager, "metric.match_rate.read", false, http.StatusForbidden},
		{"data quality permitted", http.MethodGet, "/api/v1/data-quality/sources", domain.RoleManager, "", false, http.StatusNoContent},
		{"data quality metric denied", http.MethodGet, "/api/v1/data-quality/sources", domain.RoleManager, "metric.data_quality.read", false, http.StatusForbidden},
		{"segment count allowed", http.MethodGet, "/api/v1/segments/segment-1", domain.RoleManager, "", false, http.StatusNoContent},
		{"segment metric denied", http.MethodGet, "/api/v1/segments/segment-1", domain.RoleManager, "metric.segment_size.read", false, http.StatusForbidden},
		{"segment recommendations permitted", http.MethodGet, "/api/v1/segments/segment-1/recommendations", domain.RoleMarketer, "", false, http.StatusNoContent},
		{"segment recommendations metric denied", http.MethodGet, "/api/v1/segments/segment-1/recommendations", domain.RoleMarketer, "metric.segment_size.read", false, http.StatusForbidden},
		{"exports permitted", http.MethodGet, "/api/v1/exports", domain.RoleManager, "", false, http.StatusNoContent},
		{"export preview permitted", http.MethodPost, "/api/v1/exports/preview", domain.RoleMarketer, "", false, http.StatusNoContent},
		{"export create permitted", http.MethodPost, "/api/v1/exports", domain.RoleMarketer, "", false, http.StatusNoContent},
		{"export details permitted", http.MethodGet, "/api/v1/exports/export-1", domain.RoleManager, "", false, http.StatusNoContent},
		{"export metric denied", http.MethodGet, "/api/v1/exports/export-1", domain.RoleManager, "metric.export_counts.read", false, http.StatusForbidden},
		{"own aggregate delegated", http.MethodGet, "/api/v1/tenants/t/aggregates", domain.RoleTenant, "", false, http.StatusNoContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := auth.NewManager("test-secret", time.Hour)
			token, err := tokens.Issue(domain.Principal{UserID: "u", Role: tc.role, TenantID: "t"})
			if err != nil {
				t.Fatal(err)
			}
			authorizer := &permissionTestAuthorizer{denied: tc.denied, changeVersion: tc.changeVersion}
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
			h := auth.Middleware(tokens)(PermissionMiddleware(authorizer)(next))
			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+token.AccessToken)
			response := httptest.NewRecorder()
			h.ServeHTTP(response, req)
			if response.Code != tc.want {
				t.Fatalf("got HTTP %d, want %d", response.Code, tc.want)
			}
		})
	}
}
