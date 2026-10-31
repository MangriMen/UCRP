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
		SELECT r.id, r.guest_id, r.tenant_id, COALESCE(t.name, ''), COALESCE(c.name, ''), t.floor,
			r.rank, r.score, r.explanation, r.algorithm_version, r.computed_at
		FROM recommendation r
		LEFT JOIN tenant t ON t.id = r.tenant_id
		LEFT JOIN tenant_category c ON c.id = t.category_id
		WHERE r.guest_id = $1
		ORDER BY r.rank`, guestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.Recommendation, 0)
	for rows.Next() {
		var rec domain.Recommendation
		var floor sql.NullInt64
		if err := rows.Scan(&rec.ID, &rec.GuestID, &rec.TenantID, &rec.TenantName, &rec.CategoryName, &floor,
			&rec.Rank, &rec.Score, &rec.Explanation, &rec.AlgorithmVersion, &rec.ComputedAt); err != nil {
			return nil, err
		}
		if floor.Valid {
			value := int(floor.Int64)
			rec.Floor = &value
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	factors, err := r.listFactorsByGuest(ctx, guestID)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Factors = factors[out[i].ID]
	}
	return out, nil
}

func (r *RecommendationSQL) listFactorsByGuest(ctx context.Context, guestID string) (map[string][]domain.RecommendationFactor, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT f.recommendation_id, f.factor_code, f.raw_value, f.weight, f.contribution
		FROM recommendation_factor f
		JOIN recommendation r ON r.id = f.recommendation_id
		WHERE r.guest_id = $1
		ORDER BY f.recommendation_id, f.contribution DESC, f.factor_code`, guestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string][]domain.RecommendationFactor{}
	for rows.Next() {
		var f domain.RecommendationFactor
		if err := rows.Scan(&f.RecommendationID, &f.FactorCode, &f.RawValue, &f.Weight, &f.Contribution); err != nil {
			return nil, err
		}
		out[f.RecommendationID] = append(out[f.RecommendationID], f)
	}
	return out, rows.Err()
}
