package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type AccessGrantHandler struct {
	access *service.AccessGrantService
}

func NewAccessGrantHandler(access *service.AccessGrantService) *AccessGrantHandler {
	return &AccessGrantHandler{access: access}
}

func (h *AccessGrantHandler) Permissions(w http.ResponseWriter, r *http.Request) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	items, err := h.access.Permissions(r.Context(), principal)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, items, len(items), 0, len(items))
}

func (h *AccessGrantHandler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	limit, offset := pageParams(r)
	query := r.URL.Query()
	filter := domain.AccessGrantFilter{SubjectType: query.Get("subject_type"), SubjectID: query.Get("subject_id"), Status: query.Get("status"), Limit: limit, Offset: offset}
	items, err := h.access.List(r.Context(), principal, filter)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, items, limit, offset, offset+len(items))
}

func (h *AccessGrantHandler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	item, err := h.access.Get(r.Context(), principal, chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *AccessGrantHandler) Create(w http.ResponseWriter, r *http.Request) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	var request service.GrantRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
		return
	}
	item, err := h.access.Create(r.Context(), principal, request, httpx.RequestID(r.Context()))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *AccessGrantHandler) Update(w http.ResponseWriter, r *http.Request) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	var request service.GrantPatch
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
		return
	}
	item, err := h.access.Update(r.Context(), principal, chi.URLParam(r, "id"), request, httpx.RequestID(r.Context()))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *AccessGrantHandler) Activate(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, "activate")
}
func (h *AccessGrantHandler) Suspend(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, "suspend")
}
func (h *AccessGrantHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, "revoke")
}

func (h *AccessGrantHandler) transition(w http.ResponseWriter, r *http.Request, action string) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	var request struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
		return
	}
	item, err := h.access.Transition(r.Context(), principal, chi.URLParam(r, "id"), action, request.Reason, httpx.RequestID(r.Context()))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func accessPrincipal(w http.ResponseWriter, r *http.Request) (domain.Principal, bool) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
	}
	return principal, ok
}
