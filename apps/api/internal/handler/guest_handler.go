package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q != "" && len([]rune(q)) < 2 {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "q: минимум 2 символа")
		return
	}
	limit, offset := pageParams(r)
	page, err := h.profiles.Search(r.Context(), q, limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, page)
}

func (h *GuestHandler) Get(w http.ResponseWriter, r *http.Request) {
	profile, err := h.profiles.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, profile)
}

func (h *GuestHandler) Movement(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	from, ok := optionalTime(w, r, "from")
	if !ok {
		return
	}
	to, ok := optionalTime(w, r, "to")
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	steps, err := h.profiles.Movement(r.Context(), chi.URLParam(r, "id"), from, to, limit)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": steps})
}

func optionalTime(w http.ResponseWriter, r *http.Request, name string) (*time.Time, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return nil, true
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", name+": ожидается RFC 3339")
		return nil, false
	}
	value = value.UTC()
	return &value, true
}

func (h *GuestHandler) Events(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	events, err := h.profiles.Events(r.Context(), chi.URLParam(r, "id"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, events, limit, offset, offset+len(events))
}
