package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type DataSourceHandler struct {
	sources *service.DataSourceService
}

func NewDataSourceHandler(sources *service.DataSourceService) *DataSourceHandler {
	return &DataSourceHandler{sources: sources}
}

func (h *DataSourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateDataSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	ds, err := h.sources.Create(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, ds)
}

func (h *DataSourceHandler) CreateType(w http.ResponseWriter, r *http.Request) {
	var req service.DataSourceTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	item, err := h.sources.CreateType(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *DataSourceHandler) UpdateType(w http.ResponseWriter, r *http.Request) {
	var req service.DataSourceTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	item, err := h.sources.UpdateType(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *DataSourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req service.DataSourceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	ds, err := h.sources.Update(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ds)
}

func (h *DataSourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.sources.Deactivate(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DataSourceHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.sources.List(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *DataSourceHandler) Get(w http.ResponseWriter, r *http.Request) {
	ds, err := h.sources.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ds)
}

func (h *DataSourceHandler) ListTypes(w http.ResponseWriter, r *http.Request) {
	items, err := h.sources.ListTypes(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
