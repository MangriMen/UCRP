package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type TenantGuestHandler struct {
	service *service.TenantGuestService
	legacy  *GuestHandler
}

func NewTenantGuestHandler(service *service.TenantGuestService, legacy *GuestHandler) *TenantGuestHandler {
	return &TenantGuestHandler{service: service, legacy: legacy}
}

func (h *TenantGuestHandler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	if principal.Role != domain.RoleTenant {
		h.legacy.Get(w, r)
		return
	}
	profile, err := h.service.Profile(r.Context(), principal, chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, profile)
}

func (h *TenantGuestHandler) Events(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	if principal.Role != domain.RoleTenant {
		h.legacy.Events(w, r)
		return
	}
	limit, offset := pageParams(r)
	events, total, err := h.service.Events(r.Context(), principal, chi.URLParam(r, "id"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, events, limit, offset, total)
}
