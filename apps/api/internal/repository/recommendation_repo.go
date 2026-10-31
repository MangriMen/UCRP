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

func (r *RecommendationSQL) ReplaceForGuest(ctx context.Context, guestID string, items []domain.Recommendation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM recommendation WHERE guest_id = $1`, guestID); err != nil {
		return err
	}
	for _, it := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO recommendation (id, guest_id, tenant_id, rank, score, explanation, algorithm_version, computed_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			it.ID, it.GuestID, it.TenantID, it.Rank, it.Score, it.Explanation, it.AlgorithmVersion, it.ComputedAt,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}
