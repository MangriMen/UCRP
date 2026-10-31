package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh")
	if err != nil || cookie.Value == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет refresh токена")
		return
	}
	res, err := h.auth.Refresh(r.Context(), cookie.Value)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    cookie.Value,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 3600,
	})
	httpx.WriteJSON(w, http.StatusOK, loginResponse{
		AccessToken: res.AccessToken,
		ExpiresIn:   res.ExpiresIn,
		User:        res.User,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false,
		MaxAge:   7 * 24 * 3600,
	})
}

type UserHandler struct {
	auth *service.AuthService
}

func NewUserHandler(authSvc *service.AuthService) *UserHandler {
	return &UserHandler{auth: authSvc}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.auth.ListUsers(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.auth.GetUser(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	item, err := h.auth.CreateUserService(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req service.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	item, err := h.auth.UpdateUser(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	if err := h.auth.DeactivateUser(r.Context(), chi.URLParam(r, "id"), p.Role); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type AuditHandler struct {
	audit *service.AuditService
}

func NewAuditHandler(audit *service.AuditService) *AuditHandler {
	return &AuditHandler{audit: audit}
}

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := pageParams(r)
	items, err := h.audit.List(r.Context(), q.Get("entity_type"), q.Get("entity_id"), q.Get("actor"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

type CampaignHandler struct {
	campaigns *service.CampaignService
}

func NewCampaignHandler(campaigns *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{campaigns: campaigns}
}

func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.campaigns.List(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *CampaignHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.campaigns.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	var req service.CampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	item, err := h.campaigns.Create(r.Context(), req, p.UserID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req service.CampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	item, err := h.campaigns.Update(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.campaigns.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type AnalyticsHandler struct {
	analytics *service.AnalyticsService
}

func NewAnalyticsHandler(analytics *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) Penetration(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var from, to time.Time
	if v := q.Get("from"); v != "" {
		from, _ = time.Parse(time.RFC3339, v)
	}
	if v := q.Get("to"); v != "" {
		to, _ = time.Parse(time.RFC3339, v)
	}
	res, err := h.analytics.Penetration(r.Context(), from, to)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *AnalyticsHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	res, err := h.analytics.Metrics(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *AnalyticsHandler) TenantAggregates(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	res, err := h.analytics.TenantAggregates(r.Context(), chi.URLParam(r, "id"), p.TenantID, p.Role)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *AnalyticsHandler) DataQuality(w http.ResponseWriter, r *http.Request) {
	items, err := h.analytics.DataQuality(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

var _ = domain.ErrNotFound
