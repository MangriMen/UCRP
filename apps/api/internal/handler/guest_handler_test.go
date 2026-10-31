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

type amountTestAuthorizer struct {
	decision access.Decision
}

func (a amountTestAuthorizer) Authorize(context.Context, domain.Principal, string, access.Resource) (access.Decision, error) {
	return a.decision, nil
}

func requestAsManager(t *testing.T, target string, handle http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	tokens := auth.NewManager("test-secret", time.Hour)
	token, err := tokens.Issue(domain.Principal{UserID: "u", Role: domain.RoleManager})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	response := httptest.NewRecorder()
	auth.Middleware(tokens)(handle).ServeHTTP(response, req)
	return response
}

func TestCanReadGuestField(t *testing.T) {
	cases := []struct {
		name     string
		decision access.Decision
		want     bool
	}{
		{"allowed", access.Decision{Allowed: true}, true},
		{"denied", access.Decision{Allowed: false}, false},
		{"allowed with masked fields", access.Decision{Allowed: true, MaskFields: []string{"amountMinor"}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got bool
			requestAsManager(t, "/api/v1/guests", func(_ http.ResponseWriter, r *http.Request) {
				var err error
				got, err = canReadGuestField(r, amountTestAuthorizer{decision: tc.decision}, guestAmountPermission)
				if err != nil {
					t.Fatal(err)
				}
			})
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}

	t.Run("no principal", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/guests", nil)
		got, err := canReadGuestField(req, amountTestAuthorizer{decision: access.Decision{Allowed: true}}, guestIdentifiersPermission)
		if err != nil || got {
			t.Fatalf("got %v, %v, want false, nil", got, err)
		}
	})
}

func TestGuestSearchRejectsAmountSortWithoutPermission(t *testing.T) {
	cases := []struct {
		name     string
		decision access.Decision
		sort     string
		want     int
	}{
		{"sort ascending denied", access.Decision{Allowed: false}, "totalAmountMinor", http.StatusForbidden},
		{"sort descending denied", access.Decision{Allowed: false}, "-totalAmountMinor", http.StatusForbidden},
		{"masked denied", access.Decision{Allowed: true, MaskFields: []string{"amountMinor"}}, "-totalAmountMinor", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewGuestHandler(nil, amountTestAuthorizer{decision: tc.decision})
			response := requestAsManager(t, "/api/v1/guests?sort="+tc.sort, h.Search)
			if response.Code != tc.want {
				t.Fatalf("got HTTP %d, want %d", response.Code, tc.want)
			}
		})
	}
}
