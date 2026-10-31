package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type IdentityHandler struct {
	identity   *service.IdentityService
	profiles   *service.ProfileService
	authorizer accessAuthorizer
}

func NewIdentityHandler(identity *service.IdentityService, profiles *service.ProfileService, authorizer accessAuthorizer) *IdentityHandler {
	return &IdentityHandler{identity: identity, profiles: profiles, authorizer: authorizer}
}

func (h *IdentityHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.identity.List(r.Context(), r.URL.Query().Get("resolution"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, items, limit, offset, offset+len(items))
}

type mergeCandidateDetail struct {
	service.MergeCandidateDTO
	GuestA     *service.GuestProfile        `json:"guestA"`
	GuestB     *service.GuestProfile        `json:"guestB"`
	SharedKeys []service.GuestIdentifierDTO `json:"sharedKeys"`
}

func sharedIdentifierKeys(a, b *service.GuestProfile) []service.GuestIdentifierDTO {
	type key struct{ idType, value string }
	seen := make(map[key]bool, len(a.Identifiers))
	for _, id := range a.Identifiers {
		seen[key{id.IDType, id.ValueNormalized}] = true
	}
	out := make([]service.GuestIdentifierDTO, 0, len(b.Identifiers))
	for _, id := range b.Identifiers {
		if seen[key{id.IDType, id.ValueNormalized}] {
			out = append(out, service.GuestIdentifierDTO{
				ID: id.ID, GuestID: id.GuestID, IDType: id.IDType, ValueNormalized: id.ValueNormalized,
			})
		}
	}
	return out
}

func (h *IdentityHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.identity.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	showAmount, err := canReadGuestAmount(r, h.authorizer)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	guestA, err := h.profiles.Get(r.Context(), item.GuestAID, showAmount)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	guestB, err := h.profiles.Get(r.Context(), item.GuestBID, showAmount)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, mergeCandidateDetail{
		MergeCandidateDTO: *item,
		GuestA:            guestA,
		GuestB:            guestB,
		SharedKeys:        sharedIdentifierKeys(guestA, guestB),
	})
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

func (h *IdentityHandler) Merge(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	var req struct {
		TargetGuestID string `json:"targetGuestId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TargetGuestID == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Нужен target_guest_id")
		return
	}
	item, err := h.identity.Merge(r.Context(), chi.URLParam(r, "id"), req.TargetGuestID, p.UserID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *IdentityHandler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.identity.ListBlocks(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, items, limit, offset, offset+len(items))
}

func (h *IdentityHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	if err := h.identity.DeleteBlock(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
