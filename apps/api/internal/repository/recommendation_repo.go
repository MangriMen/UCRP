package repository

import (
	"context"
	"database/sql"

	"github.com/MangriMen/UCRP/internal/domain"
)

type RecommendationSQL struct {
	db *sql.DB
}

func NewRecommendationSQL(db *sql.DB) *RecommendationSQL {
	return &RecommendationSQL{db: db}
}

func (r *RecommendationSQL) ListByGuest(ctx context.Context, guestID string) ([]domain.Recommendation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, guest_id, tenant_id, rank, score, explanation, algorithm_version, computed_at
		FROM recommendation
		WHERE guest_id = $1
		ORDER BY rank`, guestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.Recommendation, 0)
	for rows.Next() {
		var rec domain.Recommendation
		if err := rows.Scan(&rec.ID, &rec.GuestID, &rec.TenantID, &rec.Rank, &rec.Score, &rec.Explanation, &rec.AlgorithmVersion, &rec.ComputedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *RecommendationSQL) ListFactors(ctx context.Context, recommendationIDs []string) (map[string][]domain.RecommendationFactor, error) {
	out := map[string][]domain.RecommendationFactor{}
	if len(recommendationIDs) == 0 {
		return out, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT recommendation_id, factor_code, raw_value, weight, contribution
		FROM recommendation_factor
		ORDER BY recommendation_id, factor_code`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	wanted := map[string]bool{}
	for _, id := range recommendationIDs {
		wanted[id] = true
	}
	for rows.Next() {
		var f domain.RecommendationFactor
		if err := rows.Scan(&f.RecommendationID, &f.FactorCode, &f.RawValue, &f.Weight, &f.Contribution); err != nil {
			return nil, err
		}
		if !wanted[f.RecommendationID] {
			continue
		}
		out[f.RecommendationID] = append(out[f.RecommendationID], f)
	}
	return out, rows.Err()
}
