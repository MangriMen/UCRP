package service

import (
	"testing"

	"github.com/MangriMen/UCRP/internal/domain"
)

func TestApplyEvalOrderPutsBestFirst(t *testing.T) {
	items := []domain.Recommendation{
		{ID: "r1", TenantID: "t-a", Rank: 1, Score: 0.5},
		{ID: "r2", TenantID: "t-b", Rank: 2, Score: 0.3},
		{ID: "r3", TenantID: "t-c", Rank: 3, Score: 0.1},
	}
	ranking := []string{"t-b", "t-a", "t-c"}
	got := applyEvalOrder(items, ranking)
	want := []string{"t-c", "t-a", "t-b"}
	for i, id := range want {
		if got[i].TenantID != id {
			t.Fatalf("order[%d] = %s, want %s", i, got[i].TenantID, id)
		}
	}
	for i, item := range got {
		if item.Rank != i+1 {
			t.Errorf("rank[%d] = %d", i, item.Rank)
		}
	}
}

func TestApplyEvalOrderKeepsUnknownItemsLast(t *testing.T) {
	items := []domain.Recommendation{
		{ID: "r1", TenantID: "t-a", Rank: 1},
		{ID: "r2", TenantID: "t-x", Rank: 2},
	}
	got := applyEvalOrder(items, []string{"t-a"})
	if got[0].TenantID != "t-a" || got[1].TenantID != "t-x" {
		t.Errorf("order = %s, %s", got[0].TenantID, got[1].TenantID)
	}
}

func TestApplyEvalOrderEmptyRankingKeepsBaseOrder(t *testing.T) {
	items := []domain.Recommendation{
		{ID: "r1", TenantID: "t-a", Rank: 1},
		{ID: "r2", TenantID: "t-b", Rank: 2},
	}
	got := applyEvalOrder(items, nil)
	if got[0].TenantID != "t-a" || got[0].Rank != 1 {
		t.Errorf("items changed: %+v", got[0])
	}
}
