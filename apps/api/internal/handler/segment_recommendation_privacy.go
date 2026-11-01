package handler

import "github.com/MangriMen/UCRP/internal/service"

type segmentRecommendationView struct {
	TenantID        string   `json:"tenantId"`
	TenantName      string   `json:"tenantName"`
	CategoryID      string   `json:"categoryId,omitempty"`
	Score           *float64 `json:"score"`
	Guests          *int     `json:"guests"`
	EligibleCount   *int     `json:"eligibleCount"`
	SegmentCount    *int     `json:"segmentCount"`
	AverageScore    *float64 `json:"averageScore"`
	IncompleteCount *int     `json:"incompleteCount"`
	Suppressed      bool     `json:"suppressed"`
}

func isSmallCell(count int) bool {
	return count > 0 && count < minSuppressionK
}

func revealsSmallGroup(item service.SegmentRecommendationDTO) bool {
	counts := []int{
		item.Guests,
		item.EligibleCount,
		item.SegmentCount,
		item.IncompleteCount,
		item.SegmentCount - item.EligibleCount,
		item.EligibleCount - item.Guests,
	}
	for _, count := range counts {
		if isSmallCell(count) {
			return true
		}
	}
	return false
}

func segmentRecommendationResponse(segmentID string, items []service.SegmentRecommendationDTO) map[string]any {
	views := make([]segmentRecommendationView, 0, len(items))
	for _, item := range items {
		view := segmentRecommendationView{TenantID: item.TenantID, TenantName: item.TenantName, CategoryID: item.CategoryID}
		if revealsSmallGroup(item) {
			view.Suppressed = true
		} else {
			view.Score = &item.Score
			view.Guests = &item.Guests
			view.EligibleCount = &item.EligibleCount
			view.SegmentCount = &item.SegmentCount
			view.AverageScore = &item.AverageScore
			view.IncompleteCount = &item.IncompleteCount
		}
		views = append(views, view)
	}
	return map[string]any{"segmentId": segmentID, "items": views}
}
