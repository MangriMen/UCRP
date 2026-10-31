package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

const (
	defaultSegmentRecommendationLimit = 10
	maxSegmentRecommendationLimit     = 50
)

type RecommendationHandler struct {
	recommendations *service.RecommendationService
}

func NewRecommendationHandler(recommendations *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recommendations: recommendations}
}

func (h *RecommendationHandler) List(w http.ResponseWriter, r *http.Request) {
	guestID := chi.URLParam(r, "id")
	items, err := h.recommendations.ForGuest(r.Context(), guestID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, recommendationListResponse(guestID, items))
}

func (h *RecommendationHandler) SegmentRecommendations(w http.ResponseWriter, r *http.Request) {
	segmentID := chi.URLParam(r, "id")
	items, err := h.recommendations.ForSegment(r.Context(), segmentID, segmentRecommendationLimit(r))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	response := segmentRecommendationResponse(items)
	response["segmentId"] = segmentID
	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *RecommendationHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	guestID := chi.URLParam(r, "id")
	items, err := h.recommendations.Refresh(r.Context(), guestID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, recommendationListResponse(guestID, items))
}

func recommendationListResponse(guestID string, items []service.RecommendationDTO) map[string]any {
	var computedAt *string
	if len(items) > 0 {
		computedAt = &items[0].ComputedAt
	}
	return map[string]any{"guestId": guestID, "computedAt": computedAt, "items": nonNil(items)}
}

func segmentRecommendationLimit(r *http.Request) int {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		return defaultSegmentRecommendationLimit
	}
	return min(limit, maxSegmentRecommendationLimit)
}
