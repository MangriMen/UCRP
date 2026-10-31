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

type SegmentTenantScore struct {
	TenantID string
	Score    float64
	Guests   int
}

type RecommendationRepo interface {
	ListByGuest(ctx context.Context, guestID string) ([]Recommendation, error)
	TopTenantsForSegment(ctx context.Context, segmentID string, limit int) ([]SegmentTenantScore, error)
}

type Recommender interface {
	Refresh(ctx context.Context, guestID string) ([]Recommendation, error)
}
