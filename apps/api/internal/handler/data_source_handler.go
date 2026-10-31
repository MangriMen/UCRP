package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
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
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
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
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
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
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
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
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
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

const dataSourcePageSize = 50

func (h *DataSourceHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, problem := parseDataSourceFilter(r)
	if problem != "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", problem)
		return
	}
	limit, offset := pageParams(withDefaultPageSize(r, dataSourcePageSize))
	items, total, err := h.sources.Search(r.Context(), filter, limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, items, limit, offset, total)
}

func parseDataSourceFilter(r *http.Request) (domain.DataSourceFilter, string) {
	query := r.URL.Query()
	filter := domain.DataSourceFilter{
		Query:        strings.TrimSpace(query.Get("q")),
		TenantID:     strings.TrimSpace(query.Get("tenant_id")),
		TypeID:       strings.TrimSpace(query.Get("type_id")),
		ImportStatus: query.Get("import_status"),
		Sort:         query.Get("sort"),
	}
	if filter.Sort != "" && !dataSourceSortFields[filter.Sort] {
		return filter, "sort: допустимы name, tenantName, typeName, lastImportAt, createdAt с префиксом - для убывания"
	}
	if filter.ImportStatus != "" && !dataSourceImportStatuses[filter.ImportStatus] {
		return filter, "import_status: допустимы queued, running, completed, completed_with_errors, failed, none"
	}
	if raw := query.Get("is_active"); raw != "" {
		active, err := strconv.ParseBool(raw)
		if err != nil {
			return filter, "is_active: ожидается true или false"
		}
		filter.IsActive = &active
	}
	return filter, ""
}

var dataSourceSortFields = map[string]bool{
	"name": true, "tenantName": true, "typeName": true, "lastImportAt": true, "createdAt": true,
	"-name": true, "-tenantName": true, "-typeName": true, "-lastImportAt": true, "-createdAt": true,
}

var dataSourceImportStatuses = map[string]bool{
	"queued": true, "running": true, "completed": true, "completed_with_errors": true, "failed": true, "none": true,
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
	httpx.WriteJSON(w, http.StatusOK, items)
}
