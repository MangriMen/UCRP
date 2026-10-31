package handler

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type ImportHandler struct {
	imports *service.ImportService
	sources *service.DataSourceService
}

func NewImportHandler(imports *service.ImportService, sources *service.DataSourceService) *ImportHandler {
	return &ImportHandler{imports: imports, sources: sources}
}

func (h *ImportHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Ожидается multipart/form-data")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Не передан файл")
		return
	}
	defer func() { _ = file.Close() }()
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	res, err := h.imports.Create(r.Context(), service.CreateImportInput{
		DataSourceID: r.FormValue("data_source_id"),
		FileName:     header.Filename,
		File:         file,
		ActorUserID:  p.UserID,
	})
	if err != nil {
		writeImportError(w, r, err)
		return
	}
	httpx.WriteJSON(w, res.HTTPStatus, map[string]any{
		"batch_id":  res.BatchID,
		"status":    res.Status,
		"duplicate": res.Duplicate,
	})
}

func (h *ImportHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.imports.List(r.Context(), r.URL.Query().Get("data_source_id"), r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ImportHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.imports.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *ImportHandler) RejectedRows(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	if r.URL.Query().Get("format") == "csv" {
		h.rejectedCSV(w, r, chi.URLParam(r, "id"))
		return
	}
	items, err := h.imports.RejectedRows(r.Context(), chi.URLParam(r, "id"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	page := 1
	if limit > 0 {
		page = offset/limit + 1
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"items":     items,
		"page":      page,
		"page_size": limit,
		"total":     offset + len(items),
	})
}

func (h *ImportHandler) rejectedCSV(w http.ResponseWriter, r *http.Request, batchID string) {
	items, err := h.imports.RejectedRows(r.Context(), batchID, 500, 0)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="rejected-rows.csv"`)
	cw := csv.NewWriter(w)
	defer cw.Flush()
	_ = cw.Write([]string{"row_number", "error_code", "error_message", "raw_line"})
	for _, it := range items {
		_ = cw.Write([]string{strconv.Itoa(it.RowNumber), it.ErrorCode, it.ErrorMessage, it.RawLine})
	}
}

func (h *ImportHandler) Sync(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	res, err := h.imports.Sync(r.Context(), chi.URLParam(r, "id"), p.UserID)
	if err != nil {
		writeImportError(w, r, err)
		return
	}
	httpx.WriteJSON(w, res.HTTPStatus, map[string]any{
		"batch_id": res.BatchID,
		"status":   res.Status,
	})
}

func writeImportError(w http.ResponseWriter, r *http.Request, err error) {
	var ie *service.ImportError
	if errors.As(err, &ie) {
		httpx.WriteError(w, r, ie.Status, ie.Code, ie.Message)
		return
	}
	httpx.WriteDomainError(w, r, err)
}
