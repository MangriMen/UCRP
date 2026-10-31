package auth

import (
	"net/http"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
)

func RequireActiveUser(users domain.UserRepo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFrom(r.Context())
			if !ok {
				httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
				return
			}
			user, err := users.GetByID(r.Context(), principal.UserID)
			if err != nil || !user.IsActive || user.Role != principal.Role || user.TenantID != principal.TenantID {
				httpx.WriteError(w, r, http.StatusUnauthorized, "invalid_credentials", "Пользователь недоступен или его права изменились")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
