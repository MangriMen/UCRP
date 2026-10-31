package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type ExportHandler struct {
	exports *service.ExportService
}

func NewExportHandler(exports *service.ExportService) *ExportHandler {
	return &ExportHandler{exports: exports}
}

func (h *ExportHandler) Preview(w http.ResponseWriter, r *http.Request) {
	var req service.ExportPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
		return
	}
	res, err := h.exports.Preview(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *ExportHandler) Create(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	var req service.ExportCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
		return
	}
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = key
	}
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	exp, err := h.exports.Create(r.Context(), req, p.UserID)
	if err != nil {
		var ce *service.ExportConflictError
		if errors.As(err, &ce) {
			httpx.WriteError(w, r, http.StatusConflict, ce.Code, ce.Message)
			return
		}
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, exp)
}

func (h *ExportHandler) Get(w http.ResponseWriter, r *http.Request) {
	exp, err := h.exports.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, exp)
}

func (h *ExportHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.exports.List(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ExportHandler) Download(w http.ResponseWriter, r *http.Request) {
	data, err := h.exports.CSV(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
