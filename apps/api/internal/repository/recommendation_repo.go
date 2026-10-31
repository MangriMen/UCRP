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

func (r *RecommendationSQL) TopTenantsForSegment(ctx context.Context, segmentID string, limit int) ([]domain.SegmentTenantScore, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.tenant_id, COUNT(DISTINCT m.guest_id)::float8 AS score, COUNT(DISTINCT m.guest_id) AS guests
		FROM segment_member m
		JOIN segment_version v ON v.id = m.segment_version_id
		JOIN segment s ON s.id = v.segment_id AND s.current_version_no = v.version_no
		JOIN event e ON e.guest_id = m.guest_id
		JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
		WHERE s.id = $1
		  AND NOT EXISTS (
			SELECT 1 FROM event e0
			JOIN event_type et0 ON et0.id = e0.event_type_id AND et0.counts_as_presence
			WHERE e0.guest_id = m.guest_id AND e0.tenant_id = e.tenant_id
		  )
		GROUP BY e.tenant_id
		ORDER BY score DESC, e.tenant_id
		LIMIT $2`, segmentID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.SegmentTenantScore, 0)
	for rows.Next() {
		var s domain.SegmentTenantScore
		if err := rows.Scan(&s.TenantID, &s.Score, &s.Guests); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
