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
	defer rows.Close()
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
