package handler

import (
	"encoding/json"
	"testing"

	"github.com/MangriMen/UCRP/internal/service"
)

func TestSegmentRecommendationSuppression(t *testing.T) {
	items := []service.SegmentRecommendationDTO{
		{TenantID: "small", Guests: 4, EligibleCount: 8, SegmentCount: 10, IncompleteCount: 6, Score: 0.9},
		{TenantID: "large", Guests: 5, EligibleCount: 10, SegmentCount: 20, IncompleteCount: 6, Score: 0.8},
		{TenantID: "difference", Guests: 6, EligibleCount: 10, SegmentCount: 20, IncompleteCount: 6, Score: 0.7},
		{TenantID: "zeros", Guests: 10, EligibleCount: 10, SegmentCount: 10, IncompleteCount: 0, Score: 0.6},
	}
	body, err := json.Marshal(segmentRecommendationResponse(items))
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Items []struct {
			Guests     *int     `json:"guests"`
			Score      *float64 `json:"score"`
			Suppressed bool     `json:"suppressed"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 4 || !response.Items[0].Suppressed || response.Items[0].Guests != nil || response.Items[0].Score != nil {
		t.Fatalf("small group not suppressed: %s", body)
	}
	if response.Items[1].Suppressed || response.Items[1].Guests == nil || *response.Items[1].Guests != 5 || response.Items[1].Score == nil {
		t.Fatalf("eligible group hidden: %s", body)
	}
	if !response.Items[2].Suppressed || response.Items[2].Guests != nil || response.Items[2].Score != nil {
		t.Fatalf("small complementary group visible: %s", body)
	}
	if response.Items[3].Suppressed || response.Items[3].Guests == nil {
		t.Fatalf("zero counts suppressed: %s", body)
	}
}
