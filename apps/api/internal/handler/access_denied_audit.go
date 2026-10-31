package handler

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
)

type deniedAuditLimiter struct {
	mu      sync.Mutex
	entries map[[32]byte]time.Time
}

func (l *deniedAuditLimiter) reserve(key [32]byte, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if previous, ok := l.entries[key]; ok && now.Sub(previous) < time.Minute {
		return false
	}
	if len(l.entries) >= 4096 {
		var oldestKey [32]byte
		var oldest time.Time
		for candidate, seen := range l.entries {
			if oldest.IsZero() || seen.Before(oldest) {
				oldestKey, oldest = candidate, seen
			}
		}
		delete(l.entries, oldestKey)
	}
	l.entries[key] = now
	return true
}

func (l *deniedAuditLimiter) release(key [32]byte, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.entries[key].Equal(now) {
		delete(l.entries, key)
	}
}

type deniedAuditWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (w *deniedAuditWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *deniedAuditWriter) WriteHeader(status int) {
	if status >= 100 && status < 200 && status != http.StatusSwitchingProtocols {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *deniedAuditWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	if w.status == http.StatusForbidden && len(w.body) < 4096 {
		remaining := 4096 - len(w.body)
		if len(data) < remaining {
			remaining = len(data)
		}
		w.body = append(w.body, data[:remaining]...)
	}
	return w.ResponseWriter.Write(data)
}

func deniedAuditCode(body []byte) string {
	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &response) != nil {
		return "other"
	}
	switch response.Error.Code {
	case "forbidden", "forbidden_role", "forbidden_permission", "forbidden_scope", "forbidden_tenant", "forbidden_field", "metric_not_granted", "privilege_escalation", "grant_not_grantable":
		return response.Error.Code
	default:
		return "other"
	}
}

func deniedAuditMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions:
		return method
	default:
		return "OTHER"
	}
}

func DeniedAuditMiddleware(repo domain.AuditRepo, ids domain.IDGen, clock domain.Clock) func(http.Handler) http.Handler {
	limiter := &deniedAuditLimiter{entries: make(map[[32]byte]time.Time)}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := auth.PrincipalFrom(r.Context())
			writer := &deniedAuditWriter{ResponseWriter: w}
			next.ServeHTTP(writer, r)
			if !ok || writer.status != http.StatusForbidden {
				return
			}
			now := clock.Now().UTC()
			code := deniedAuditCode(writer.body)
			key := sha256.Sum256([]byte(principal.UserID + "\x00" + code))
			if !limiter.reserve(key, now) {
				return
			}
			route := "unmatched"
			if routeContext := chi.RouteContext(r.Context()); routeContext != nil && routeContext.RoutePattern() != "" {
				route = routeContext.RoutePattern()
			}
			details, err := json.Marshal(map[string]string{"method": deniedAuditMethod(r.Method), "route": route, "code": code})
			if err != nil {
				limiter.release(key, now)
				return
			}
			ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 2*time.Second)
			defer cancel()
			entry := &domain.AuditLog{ID: ids.NewID(ctx), ActorUserID: principal.UserID, ActorRole: string(principal.Role),
				Action: "access.denied", EntityType: "access", EntityID: principal.UserID, Details: string(details), CreatedAt: now}
			if err := repo.Insert(ctx, entry); err != nil {
				limiter.release(key, now)
				slog.Error("access denied audit failed")
			}
		})
	}
}
