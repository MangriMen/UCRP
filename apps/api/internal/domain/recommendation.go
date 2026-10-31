package domain

import (
	"context"
	"time"
)

type Recommendation struct {
	ID               string
	GuestID          string
	TenantID         string
	Rank             int
	Score            float64
	Explanation      string
	AlgorithmVersion string
	ComputedAt       time.Time
}

type RecommendationRepo interface {
	ListByGuest(ctx context.Context, guestID string) ([]Recommendation, error)
	ReplaceForGuest(ctx context.Context, guestID string, items []Recommendation) error
}

type Recommender interface {
	Refresh(ctx context.Context, guestID string) ([]Recommendation, error)
}
