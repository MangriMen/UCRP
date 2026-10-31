package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserUpdateRejectsCredentialChanges(t *testing.T) {
	for _, body := range []string{`{"role":"analyst"}`, `{"tenantId":"other"}`, `{"password":"secret"}`, `{"isActive":true}`} {
		response := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/users/id", strings.NewReader(body))
		NewUserHandler(nil).Update(response, req)
		if response.Code != http.StatusForbidden {
			t.Fatalf("request %s: got %d, want 403", body, response.Code)
		}
	}
}
