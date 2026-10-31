package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MangriMen/UCRP/internal/domain"
)

type liveUserRepo struct {
	domain.UserRepo
	user *domain.User
	err  error
}

func (r liveUserRepo) GetByID(context.Context, string) (*domain.User, error) {
	return r.user, r.err
}

func TestRequireActiveUser(t *testing.T) {
	principal := domain.Principal{UserID: "user", Role: domain.RoleTenant, TenantID: "tenant"}
	cases := []struct {
		name string
		user *domain.User
		err  error
		want int
	}{
		{"active", &domain.User{IsActive: true, Role: domain.RoleTenant, TenantID: "tenant"}, nil, http.StatusNoContent},
		{"inactive", &domain.User{Role: domain.RoleTenant, TenantID: "tenant"}, nil, http.StatusUnauthorized},
		{"changed role", &domain.User{IsActive: true, Role: domain.RoleManager}, nil, http.StatusUnauthorized},
		{"changed tenant", &domain.User{IsActive: true, Role: domain.RoleTenant, TenantID: "other"}, nil, http.StatusUnauthorized},
		{"password recovered", &domain.User{IsActive: true, Role: domain.RoleTenant, TenantID: "tenant", AuthVersion: 1}, nil, http.StatusUnauthorized},
		{"repository error", nil, errors.New("unavailable"), http.StatusUnauthorized},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
			req = req.WithContext(context.WithValue(req.Context(), principalKey, principal))
			response := httptest.NewRecorder()
			handler := RequireActiveUser(liveUserRepo{user: test.user, err: test.err})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			handler.ServeHTTP(response, req)
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
}
