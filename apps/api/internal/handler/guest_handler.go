package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type GuestHandler struct {
	profiles *service.ProfileService
}

func NewGuestHandler(profiles *service.ProfileService) *GuestHandler {
	return &GuestHandler{profiles: profiles}
}

func (h *GuestHandler) Search(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.profiles.Search(r.Context(), r.URL.Query().Get("q"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *GuestHandler) Get(w http.ResponseWriter, r *http.Request) {
	profile, err := h.profiles.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, profile)
}

func (h *GuestHandler) Events(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	events, err := h.profiles.Events(r.Context(), chi.URLParam(r, "id"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": events})
}

func pageParams(r *http.Request) (int, int) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return limit, offset
}
