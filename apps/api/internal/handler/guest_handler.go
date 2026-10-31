package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type GuestHandler struct {
	profiles   *service.ProfileService
	authorizer accessAuthorizer
}

func NewGuestHandler(profiles *service.ProfileService, authorizer accessAuthorizer) *GuestHandler {
	return &GuestHandler{profiles: profiles, authorizer: authorizer}
}

const guestAmountPermission = "field.guest.total_amount.read"

var guestListSortFields = map[string]bool{
	"createdAt":         true,
	"-createdAt":        true,
	"eventCount":        true,
	"-eventCount":       true,
	"totalAmountMinor":  true,
	"-totalAmountMinor": true,
	"firstSeenAt":       true,
	"-firstSeenAt":      true,
	"lastSeenAt":        true,
	"-lastSeenAt":       true,
}

func parseDateParam(query url.Values, name string, inclusiveEnd bool) (*time.Time, string) {
	raw := query.Get(name)
	if raw == "" {
		return nil, ""
	}
	date, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, name + ": ожидается дата YYYY-MM-DD"
	}
	if inclusiveEnd {
		date = date.AddDate(0, 0, 1)
	}
	return &date, ""
}

func parseGuestFilter(r *http.Request) (domain.GuestFilter, string) {
	query := r.URL.Query()
	filter := domain.GuestFilter{
		Query:     strings.TrimSpace(query.Get("q")),
		SourceID:  strings.TrimSpace(query.Get("source_id")),
		TenantID:  strings.TrimSpace(query.Get("tenant_id")),
		SegmentID: strings.TrimSpace(query.Get("segment_id")),
		Status:    domain.GuestStatus(query.Get("status")),
		Sort:      query.Get("sort"),
	}
	if filter.Query != "" && len([]rune(filter.Query)) < 2 {
		return filter, "q: минимум 2 символа"
	}
	switch filter.Status {
	case "", domain.GuestActive, domain.GuestMerged, domain.GuestDeleted:
	default:
		return filter, "status: допустимы active, merged, deleted"
	}
	if filter.Sort != "" && !guestListSortFields[filter.Sort] {
		return filter, "sort: допустимы createdAt, eventCount, totalAmountMinor, firstSeenAt, lastSeenAt с префиксом - для убывания"
	}
	if raw := query.Get("has_purchases"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return filter, "has_purchases: ожидается true или false"
		}
		filter.HasPurchases = &value
	}
	dates := []struct {
		name         string
		target       **time.Time
		inclusiveEnd bool
	}{
		{"period_from", &filter.PeriodFrom, false},
		{"period_to", &filter.PeriodTo, true},
		{"first_seen_from", &filter.FirstSeenFrom, false},
		{"last_seen_to", &filter.LastSeenTo, true},
	}
	for _, d := range dates {
		value, problem := parseDateParam(query, d.name, d.inclusiveEnd)
		if problem != "" {
			return filter, problem
		}
		*d.target = value
	}
	if filter.PeriodFrom != nil && filter.PeriodTo != nil && !filter.PeriodFrom.Before(*filter.PeriodTo) {
		return filter, "period_from не может быть позже period_to"
	}
	return filter, ""
}

func canReadGuestAmount(r *http.Request, authorizer accessAuthorizer) (bool, error) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok || authorizer == nil {
		return false, nil
	}
	decision, err := authorizer.Authorize(r.Context(), principal, guestAmountPermission, access.Resource{})
	if err != nil {
		return false, err
	}
	return decision.Allowed && len(decision.MaskFields) == 0, nil
}

func (h *GuestHandler) Search(w http.ResponseWriter, r *http.Request) {
	h.search(w, r, "")
}

func (h *GuestHandler) search(w http.ResponseWriter, r *http.Request, tenantID string) {
	filter, problem := parseGuestFilter(r)
	if problem != "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", problem)
		return
	}
	if tenantID != "" {
		filter.TenantID = tenantID
	}
	showAmount, err := canReadGuestAmount(r, h.authorizer)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	if !showAmount && strings.TrimPrefix(filter.Sort, "-") == "totalAmountMinor" {
		httpx.WriteError(w, r, http.StatusForbidden, "forbidden_permission", "Нет доступа к сумме покупок")
		return
	}
	limit, offset := pageParams(r)
	page, err := h.profiles.Search(r.Context(), filter, limit, offset, showAmount)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, page)
}

func (h *GuestHandler) Get(w http.ResponseWriter, r *http.Request) {
	showAmount, err := canReadGuestAmount(r, h.authorizer)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	profile, err := h.profiles.Get(r.Context(), chi.URLParam(r, "id"), showAmount)
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
