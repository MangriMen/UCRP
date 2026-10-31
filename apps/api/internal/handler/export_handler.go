package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
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
	limit, offset := pageParams(withDefaultPageSize(r))
	query := r.URL.Query()
	filter := domain.ExportFilter{
		SegmentID:  query.Get("segment_id"),
		CampaignID: query.Get("campaign_id"),
		Channel:    query.Get("channel"),
		Status:     query.Get("status"),
	}
	items, total, err := h.exports.List(r.Context(), filter, limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, items, limit, offset, total)
}

func (h *ExportHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	file, err := h.exports.Download(r.Context(), id)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="export-%s.%s"`, id, file.Extension))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(file.Data)
}
