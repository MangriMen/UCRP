package handler

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
)

type deniedAuditTestRepo struct {
	domain.AuditRepo
	entries  []domain.AuditLog
	attempts int
	fail     bool
}

func TestDeniedAuditLimiterIsBounded(t *testing.T) {
	limiter := &deniedAuditLimiter{entries: make(map[[32]byte]time.Time)}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5000; i++ {
		limiter.reserve(sha256.Sum256([]byte(fmt.Sprint(i))), now.Add(time.Duration(i)*time.Millisecond))
	}
	if len(limiter.entries) != 4096 {
		t.Fatalf("limiter retained %d keys, want 4096", len(limiter.entries))
	}
}

func (r *deniedAuditTestRepo) Insert(_ context.Context, entry *domain.AuditLog) error {
	r.attempts++
	if r.fail {
		return errors.New("database unavailable")
	}
	r.entries = append(r.entries, *entry)
	return nil
}

type deniedAuditTestClock struct{ now time.Time }

func (c *deniedAuditTestClock) Now() time.Time { return c.now }

type deniedAuditTestIDs struct{ next int }

func (ids *deniedAuditTestIDs) NewID(context.Context) string {
	ids.next++
	return fmt.Sprint(ids.next)
}

func TestDeniedAuditMiddleware(t *testing.T) {
	tokens := auth.NewManager("secret", time.Hour)
	token, err := tokens.Issue(domain.Principal{UserID: "user-1", Role: domain.RoleManager})
	if err != nil {
		t.Fatal(err)
	}
	repo := &deniedAuditTestRepo{}
	clock := &deniedAuditTestClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	router := chi.NewRouter()
	router.Use(auth.Middleware(tokens))
	router.Use(DeniedAuditMiddleware(repo, &deniedAuditTestIDs{}, clock))
	router.Get("/guests/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("allow") == "yes" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		httpx.WriteError(w, r, http.StatusForbidden, r.URL.Query().Get("code"), "private-phone-and-email")
	})
	request := func(query string, want int) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/guests/private-contact?"+query+"&email=secret@example.test", nil)
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != want {
			t.Fatalf("status %d, want %d", response.Code, want)
		}
	}
	request("code=forbidden_permission", http.StatusForbidden)
	request("code=forbidden_permission", http.StatusForbidden)
	request("code=forbidden_scope", http.StatusForbidden)
	request("allow=yes", http.StatusNoContent)
	if len(repo.entries) != 2 {
		t.Fatalf("got %d audit entries, want two distinct denial codes", len(repo.entries))
	}
	entry := repo.entries[0]
	if entry.ActorUserID != "user-1" || entry.ActorRole != "manager" || entry.Action != "access.denied" || !strings.Contains(entry.Details, "/guests/{id}") {
		t.Fatalf("unexpected audit entry: %+v", entry)
	}
	if !strings.Contains(entry.Details, "forbidden_permission") {
		t.Fatal("audit entry does not contain the sanitized denial code")
	}
	for _, secret := range []string{"private-contact", "secret@example.test", "private-phone-and-email"} {
		if strings.Contains(fmt.Sprintf("%+v", entry), secret) {
			t.Fatalf("audit entry contains sensitive request content: %s", secret)
		}
	}
	clock.now = clock.now.Add(time.Minute)
	request("code=forbidden_permission", http.StatusForbidden)
	if len(repo.entries) != 3 {
		t.Fatal("denial was not recorded after rate limit expired")
	}
	repo.fail = true
	clock.now = clock.now.Add(time.Minute)
	request("code=forbidden_permission", http.StatusForbidden)
	repo.fail = false
	request("code=forbidden_permission", http.StatusForbidden)
	if len(repo.entries) != 4 || repo.attempts != 5 {
		t.Fatalf("failed audit did not preserve denial and allow retry: entries=%d attempts=%d", len(repo.entries), repo.attempts)
	}
}
