package domain

import (
	"context"
	"time"
)

type Recommendation struct {
	ID               string
	GuestID          string
	TenantID         string
	TenantName       string
	CategoryName     string
	Floor            *int
	Rank             int
	Score            float64
	Explanation      string
	AlgorithmVersion string
	ComputedAt       time.Time
	Factors          []RecommendationFactor
}

type RecommendationFactor struct {
	RecommendationID string
	FactorCode       string
	RawValue         float64
	Weight           float64
	Contribution     float64
}

type SegmentTenantScore struct {
	TenantID        string
	TenantName      string
	CategoryName    string
	GuestCount      int
	EligibleCount   int
	SegmentCount    int
	AverageScore    float64
	IncompleteCount int
}

type RecommendationEval struct {
	ID                       string
	GuestID                  string
	Probabilities            map[string]float64
	Ranking                  []string
	Confidence               *float64
	RecommendationComputedAt time.Time
	EvaluatedAt              time.Time
}

type RecommendationRepo interface {
	ListByGuest(ctx context.Context, guestID string) ([]Recommendation, error)
}

type RecommendationEvalRepo interface {
	Insert(ctx context.Context, ev *RecommendationEval) error
}

type Recommender interface {
	Refresh(ctx context.Context, guestID string) ([]Recommendation, error)
}
