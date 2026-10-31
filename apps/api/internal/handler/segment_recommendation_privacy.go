package handler

import "github.com/MangriMen/UCRP/internal/service"

type segmentRecommendationView struct {
	TenantID        string   `json:"tenant_id"`
	TenantName      string   `json:"tenant_name"`
	CategoryID      string   `json:"category_id,omitempty"`
	Score           *float64 `json:"score"`
	Guests          *int     `json:"guests"`
	EligibleCount   *int     `json:"eligible_count"`
	SegmentCount    *int     `json:"segment_count"`
	AverageScore    *float64 `json:"average_score"`
	IncompleteCount *int     `json:"incomplete_count"`
	Suppressed      bool     `json:"suppressed"`
}

func segmentRecommendationResponse(items []service.SegmentRecommendationDTO) map[string]any {
	views := make([]segmentRecommendationView, 0, len(items))
	for _, item := range items {
		view := segmentRecommendationView{TenantID: item.TenantID, TenantName: item.TenantName, CategoryID: item.CategoryID}
		if item.Guests < 5 || item.EligibleCount < 5 || item.SegmentCount < 5 || item.IncompleteCount < 5 ||
			item.SegmentCount-item.EligibleCount < 5 || item.EligibleCount-item.Guests < 5 {
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
	return map[string]any{"items": views}
}
