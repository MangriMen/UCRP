package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

func (h *AccessGrantHandler) Presets(w http.ResponseWriter, r *http.Request) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	items, err := h.access.Presets(r.Context(), principal)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, items, len(items), 0, len(items))
}

func (h *AccessGrantHandler) EffectiveSubject(w http.ResponseWriter, r *http.Request) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	query := r.URL.Query()
	item, err := h.access.SubjectEffective(r.Context(), principal, query.Get("subject_type"), query.Get("subject_id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *AccessGrantHandler) Simulate(w http.ResponseWriter, r *http.Request) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	var request service.GrantRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
		return
	}
	item, err := h.access.Simulate(r.Context(), principal, request, httpx.RequestID(r.Context()))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *AccessGrantHandler) Audit(w http.ResponseWriter, r *http.Request) {
	principal, ok := accessPrincipal(w, r)
	if !ok {
		return
	}
	query := r.URL.Query()
	var from, to *time.Time
	if query.Get("from") != "" {
		value, err := time.Parse(time.RFC3339, query.Get("from"))
		if err != nil {
			httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректная дата from")
			return
		}
		from = &value
	}
	if query.Get("to") != "" {
		value, err := time.Parse(time.RFC3339, query.Get("to"))
		if err != nil {
			httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректная дата to")
			return
		}
		to = &value
	}
	limit, offset := accessPageParams(r)
	items, total, err := h.access.Audit(r.Context(), principal, query.Get("grant_id"), query.Get("actor"), from, to, limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, items, limit, offset, total)
}
