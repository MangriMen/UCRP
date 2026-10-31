package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type SegmentHandler struct {
	segments *service.SegmentService
}

func NewSegmentHandler(segments *service.SegmentService) *SegmentHandler {
	return &SegmentHandler{segments: segments}
}

func (h *SegmentHandler) Preview(w http.ResponseWriter, r *http.Request) {
	var req service.SegmentPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	res, err := h.segments.Preview(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *SegmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.SegmentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	seg, err := h.segments.Create(r.Context(), req, p.UserID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, seg)
}

func (h *SegmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	seg, err := h.segments.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, seg)
}

func (h *SegmentHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.segments.List(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
