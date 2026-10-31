package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type IdentityHandler struct {
	identity *service.IdentityService
}

func NewIdentityHandler(identity *service.IdentityService) *IdentityHandler {
	return &IdentityHandler{identity: identity}
}

func (h *IdentityHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.identity.List(r.Context(), r.URL.Query().Get("resolution"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *IdentityHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.identity.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *IdentityHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	h.decide(w, r, h.identity.Confirm)
}

func (h *IdentityHandler) Reject(w http.ResponseWriter, r *http.Request) {
	h.decide(w, r, h.identity.Reject)
}

func (h *IdentityHandler) Revert(w http.ResponseWriter, r *http.Request) {
	h.decide(w, r, h.identity.Revert)
}

func (h *IdentityHandler) decide(
	w http.ResponseWriter,
	r *http.Request,
	fn func(context.Context, string, string) (*service.MergeCandidateDTO, error),
) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	item, err := fn(r.Context(), chi.URLParam(r, "id"), p.UserID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
