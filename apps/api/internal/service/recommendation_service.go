package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type RecommendationService struct {
	guests domain.GuestRepo
	repo   domain.RecommendationRepo
	evals  domain.RecommendationEvalRepo
	rec    domain.Recommender
	q      domain.Queue
}

func NewRecommendationService(guests domain.GuestRepo, repo domain.RecommendationRepo, evals domain.RecommendationEvalRepo, rec domain.Recommender, q domain.Queue) *RecommendationService {
	return &RecommendationService{guests: guests, repo: repo, evals: evals, rec: rec, q: q}
}

type RecommendationDTO struct {
	ID               string                    `json:"id"`
	GuestID          string                    `json:"guestId"`
	Rank             int                       `json:"rank"`
	TenantID         string                    `json:"tenantId"`
	TenantName       string                    `json:"tenantName,omitempty"`
	CategoryName     *string                   `json:"categoryName"`
	Floor            *int                      `json:"floor"`
	Score            float64                   `json:"score"`
	Explanation      string                    `json:"explanation"`
	AlgorithmVersion string                    `json:"algorithmVersion"`
	Factors          []RecommendationFactorDTO `json:"factors"`
	ComputedAt       string                    `json:"computedAt"`
}

type RecommendationFactorDTO struct {
	FactorCode   string  `json:"factorCode"`
	RawValue     float64 `json:"rawValue"`
	Weight       float64 `json:"weight"`
	Contribution float64 `json:"contribution"`
}

type SegmentRecommendationDTO struct {
	TenantID        string  `json:"tenantId"`
	TenantName      string  `json:"tenantName"`
	CategoryID      string  `json:"categoryId,omitempty"`
	Score           float64 `json:"score"`
	Guests          int     `json:"guests"`
	EligibleCount   int     `json:"eligibleCount"`
	SegmentCount    int     `json:"segmentCount"`
	AverageScore    float64 `json:"averageScore"`
	IncompleteCount int     `json:"incompleteCount"`
}

type SegmentAggregator interface {
	AggregateSegment(ctx context.Context, segmentID string, limit int) ([]domain.SegmentTenantScore, error)
}

func (s *RecommendationService) ForSegment(ctx context.Context, segmentID string, limit int) ([]SegmentRecommendationDTO, error) {
	agg, ok := s.rec.(SegmentAggregator)
	if !ok {
		return nil, fmt.Errorf("recommendations: segment aggregation is not available")
	}
	items, err := agg.AggregateSegment(ctx, segmentID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]SegmentRecommendationDTO, 0, len(items))
	for _, it := range items {
		out = append(out, SegmentRecommendationDTO{
			TenantID:        it.TenantID,
			TenantName:      it.TenantName,
			CategoryID:      it.CategoryName,
			Score:           it.AverageScore,
			Guests:          it.GuestCount,
			EligibleCount:   it.EligibleCount,
			SegmentCount:    it.SegmentCount,
			AverageScore:    it.AverageScore,
			IncompleteCount: it.IncompleteCount,
		})
	}
	return out, nil
}

func (s *RecommendationService) ForGuest(ctx context.Context, guestID string) ([]RecommendationDTO, error) {
	if _, err := s.guests.GetByID(ctx, guestID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListByGuest(ctx, guestID)
	if err != nil {
		return nil, err
	}
	ev, err := s.evals.LatestForGuest(ctx, guestID)
	if err != nil {
		return nil, err
	}
	if ev != nil && len(items) > 0 && ev.RecommendationComputedAt.Equal(items[0].ComputedAt) {
		items = applyEvalOrder(items, ev.Ranking)
	}
	return toRecoDTOs(items), nil
}

func applyEvalOrder(items []domain.Recommendation, ranking []string) []domain.Recommendation {
	position := map[string]int{}
	for i, tenantID := range ranking {
		position[tenantID] = i
	}
	if len(position) == 0 {
		return items
	}
	ordered := make([]domain.Recommendation, 0, len(items))
	var rest []domain.Recommendation
	for _, item := range items {
		if _, ok := position[item.TenantID]; ok {
			ordered = append(ordered, item)
		} else {
			rest = append(rest, item)
		}
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		return position[ordered[i].TenantID] > position[ordered[j].TenantID]
	})
	ordered = append(ordered, rest...)
	for i := range ordered {
		ordered[i].Rank = i + 1
	}
	return ordered
}

func (s *RecommendationService) Refresh(ctx context.Context, guestID string) ([]RecommendationDTO, error) {
	if _, err := s.guests.GetByID(ctx, guestID); err != nil {
		return nil, err
	}
	if _, err := s.rec.Refresh(ctx, guestID); err != nil {
		return nil, fmt.Errorf("refresh: %w", err)
	}
	if err := enqueueRecommendEvaluate(ctx, s.q, guestID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListByGuest(ctx, guestID)
	if err != nil {
		return nil, err
	}
	return toRecoDTOs(items), nil
}

func enqueueRecommendEvaluate(ctx context.Context, q domain.Queue, guestID string) error {
	payload, err := json.Marshal(map[string]string{"guest_id": guestID})
	if err != nil {
		return err
	}
	return q.Enqueue(ctx, domain.Job{
		JobType:  "recommend.evaluate",
		Payload:  string(payload),
		DedupKey: "eval:" + guestID,
	})
}

func toRecoDTOs(items []domain.Recommendation) []RecommendationDTO {
	out := make([]RecommendationDTO, 0, len(items))
	for _, it := range items {
		factors := make([]RecommendationFactorDTO, 0, len(it.Factors))
		for _, f := range it.Factors {
			factors = append(factors, RecommendationFactorDTO{
				FactorCode:   f.FactorCode,
				RawValue:     f.RawValue,
				Weight:       f.Weight,
				Contribution: f.Contribution,
			})
		}
		var categoryName *string
		if it.CategoryName != "" {
			categoryName = &it.CategoryName
		}
		out = append(out, RecommendationDTO{
			ID:               it.ID,
			GuestID:          it.GuestID,
			Rank:             it.Rank,
			TenantID:         it.TenantID,
			TenantName:       it.TenantName,
			CategoryName:     categoryName,
			Floor:            it.Floor,
			Score:            it.Score,
			Explanation:      it.Explanation,
			AlgorithmVersion: it.AlgorithmVersion,
			Factors:          factors,
			ComputedAt:       it.ComputedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}
