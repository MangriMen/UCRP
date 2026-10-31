package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type RecommendationHandler struct {
	recommendations *service.RecommendationService
}

func NewRecommendationHandler(recommendations *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recommendations: recommendations}
}

func (h *RecommendationHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.recommendations.ForGuest(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *RecommendationHandler) SegmentRecommendations(w http.ResponseWriter, r *http.Request) {
	items, err := h.recommendations.ForSegment(r.Context(), chi.URLParam(r, "id"), 10)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, segmentRecommendationResponse(items))
}

func (h *RecommendationHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	items, err := h.recommendations.Refresh(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
