package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
)

func TestTenantAccessMetricPermissions(t *testing.T) {
	cases := []struct {
		name   string
		query  string
		denied string
		status int
		code   string
	}{
		{"action denied", "", "tenant.aggregates.own", http.StatusForbidden, "forbidden"},
		{"unrequested metric denied", "", "metric.visits_count.read", http.StatusOK, ""},
		{"explicit metric denied", "?metrics=visits_count", "metric.visits_count.read", http.StatusForbidden, "metric_not_granted"},
		{"explicit permission code denied", "?metrics=metric.visits_count.read", "metric.visits_count.read", http.StatusForbidden, "metric_not_granted"},
		{"unknown metric", "?metrics=other", "", http.StatusForbidden, "metric_not_granted"},
		{"repeated requested denied metric", "?metrics=visits_count&metrics=visits_count", "metric.visits_count.read", http.StatusForbidden, "metric_not_granted"},
		{"repeated metric with unknown", "?metrics=visits_count&metrics=other", "", http.StatusForbidden, "metric_not_granted"},
		{"comma separated unknown", "?metrics=visits_count,other", "", http.StatusForbidden, "metric_not_granted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			authorizer := &permissionTestAuthorizer{denied: tc.denied}
			handler := NewTenantAccessHandler(authorizer, nil)
			tokens := auth.NewManager("test-secret", time.Hour)
			token, err := tokens.Issue(domain.Principal{UserID: "u", Role: domain.RoleTenant, TenantID: "own"})
			if err != nil {
				t.Fatal(err)
			}
			router := chi.NewRouter()
			router.Use(auth.Middleware(tokens))
			router.Get("/tenants/{id}/aggregates", handler.Aggregates)
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tenants/own/aggregates"+tc.query, nil)
			request.Header.Set("Authorization", "Bearer "+token.AccessToken)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("got status %d, want %d: %s", response.Code, tc.status, response.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tc.status == http.StatusOK {
				if len(body) != 0 {
					t.Fatalf("ungranted metric leaked: %v", body)
				}
				return
			}
			errorBody, ok := body["error"].(map[string]any)
			if !ok || errorBody["code"] != tc.code {
				t.Fatalf("got error %v, want code %q", body, tc.code)
			}
		})
	}
}
