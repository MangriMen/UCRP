package service

import (
	"context"
	"fmt"

	"github.com/MangriMen/UCRP/internal/domain"
)

type RecommendationService struct {
	guests domain.GuestRepo
	repo   domain.RecommendationRepo
	rec    domain.Recommender
}

func NewRecommendationService(guests domain.GuestRepo, repo domain.RecommendationRepo, rec domain.Recommender) *RecommendationService {
	return &RecommendationService{guests: guests, repo: repo, rec: rec}
}

type RecommendationDTO struct {
	Rank             int     `json:"rank"`
	TenantID         string  `json:"tenantId"`
	Score            float64 `json:"score"`
	Explanation      string  `json:"explanation"`
	AlgorithmVersion string  `json:"algorithmVersion"`
	ComputedAt       string  `json:"computedAt"`
}

type SegmentRecommendationDTO struct {
	TenantID string  `json:"tenantId"`
	Score    float64 `json:"score"`
	Guests   int     `json:"guests"`
}

func (s *RecommendationService) ForSegment(ctx context.Context, segmentID string, limit int) ([]SegmentRecommendationDTO, error) {
	items, err := s.repo.TopTenantsForSegment(ctx, segmentID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]SegmentRecommendationDTO, 0, len(items))
	for _, it := range items {
		out = append(out, SegmentRecommendationDTO{
			TenantID: it.TenantID,
			Score:    it.Score,
			Guests:   it.Guests,
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
	return toRecoDTOs(items), nil
}

func (s *RecommendationService) Refresh(ctx context.Context, guestID string) ([]RecommendationDTO, error) {
	if _, err := s.guests.GetByID(ctx, guestID); err != nil {
		return nil, err
	}
	items, err := s.rec.Refresh(ctx, guestID)
	if err != nil {
		return nil, fmt.Errorf("refresh: %w", err)
	}
	return toRecoDTOs(items), nil
}

func toRecoDTOs(items []domain.Recommendation) []RecommendationDTO {
	out := make([]RecommendationDTO, 0, len(items))
	for _, it := range items {
		out = append(out, RecommendationDTO{
			Rank:             it.Rank,
			TenantID:         it.TenantID,
			Score:            it.Score,
			Explanation:      it.Explanation,
			AlgorithmVersion: it.AlgorithmVersion,
			ComputedAt:       it.ComputedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return out
}
