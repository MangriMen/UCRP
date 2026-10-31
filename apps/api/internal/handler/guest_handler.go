package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
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

const defaultEventPageSize = 50

var guestEventSortFields = map[string]bool{
	"occurredAt":   true,
	"-occurredAt":  true,
	"amountMinor":  true,
	"-amountMinor": true,
}

func parseEventFilter(r *http.Request) (domain.EventFilter, string) {
	query := r.URL.Query()
	filter := domain.EventFilter{
		TypeCode: strings.TrimSpace(query.Get("type")),
		TenantID: strings.TrimSpace(query.Get("tenant_id")),
		Sort:     query.Get("sort"),
	}
	if filter.Sort != "" && !guestEventSortFields[filter.Sort] {
		return filter, "sort: допустимы occurredAt, amountMinor с префиксом - для убывания"
	}
	if raw := query.Get("from"); raw != "" {
		from, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return filter, "from: ожидается дата YYYY-MM-DD"
		}
		filter.From = &from
	}
	if raw := query.Get("to"); raw != "" {
		to, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return filter, "to: ожидается дата YYYY-MM-DD"
		}
		end := to.AddDate(0, 0, 1)
		filter.To = &end
	}
	if raw := query.Get("amount_min"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 {
			return filter, "amount_min: ожидается неотрицательное целое"
		}
		filter.AmountMin = &value
	}
	if raw := query.Get("amount_max"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 {
			return filter, "amount_max: ожидается неотрицательное целое"
		}
		filter.AmountMax = &value
	}
	if filter.AmountMin != nil && filter.AmountMax != nil && *filter.AmountMin > *filter.AmountMax {
		return filter, "amount_min не может быть больше amount_max"
	}
	return filter, ""
}

func withDefaultPageSize(r *http.Request, size int) *http.Request {
	query := r.URL.Query()
	if query.Get("limit") != "" || query.Get("page_size") != "" {
		return r
	}
	query.Set("page_size", strconv.Itoa(size))
	clone := r.Clone(r.Context())
	clone.URL.RawQuery = query.Encode()
	return clone
}

func (h *GuestHandler) Events(w http.ResponseWriter, r *http.Request) {
	filter, problem := parseEventFilter(r)
	if problem != "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", problem)
		return
	}
	limit, offset := pageParams(withDefaultPageSize(r, defaultEventPageSize))
	events, total, err := h.profiles.Events(r.Context(), chi.URLParam(r, "id"), filter, limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, events, limit, offset, total)
}
