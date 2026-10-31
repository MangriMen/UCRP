package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type AnalyticsService struct {
	db *sql.DB
}

func NewAnalyticsService(db *sql.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

type PenetrationBucket struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type PenetrationResponse struct {
	One       int                 `json:"one"`
	Two       int                 `json:"two"`
	ThreePlus int                 `json:"three_plus"`
	Share     float64             `json:"share"`
	Buckets   []PenetrationBucket `json:"buckets"`
}

func (s *AnalyticsService) Penetration(ctx context.Context, from, to time.Time) (*PenetrationResponse, error) {
	if to.IsZero() {
		to = time.Now().UTC()
	}
	if from.IsZero() {
		from = to.AddDate(0, -1, 0)
	}
	if from.After(to) {
		return nil, domain.ErrValidation
	}
	if to.Sub(from) > 24*31*24*time.Hour {
		return nil, domain.ErrValidation
	}
	rows, err := s.db.QueryContext(ctx, `
		WITH guest_tenants AS (
			SELECT e.guest_id, COUNT(DISTINCT e.tenant_id) AS tn
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			WHERE e.guest_id IS NOT NULL AND e.occurred_at >= $1 AND e.occurred_at <= $2
			GROUP BY e.guest_id
		)
		SELECT
			COALESCE(SUM(CASE WHEN tn = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tn = 2 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tn >= 3 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tn >= 2 THEN 1 ELSE 0 END), 0)::float8 / NULLIF(COUNT(*), 0)
		FROM guest_tenants`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out PenetrationResponse
	if err := rows.Scan(&out.One, &out.Two, &out.ThreePlus, &out.Share); err != nil {
		return nil, err
	}
	out.Buckets = []PenetrationBucket{
		{Label: "1", Count: out.One},
		{Label: "2", Count: out.Two},
		{Label: "3+", Count: out.ThreePlus},
	}
	return &out, nil
}

type MetricsResponse struct {
	MatchRate          float64 `json:"match_rate"`
	ContestedMergeRate float64 `json:"contested_merge_rate"`
	AvgSegmentToExport float64 `json:"avg_segment_to_export_minutes"`
}

func (s *AnalyticsService) Metrics(ctx context.Context) (*MetricsResponse, error) {
	var out MetricsResponse
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN guest_id IS NOT NULL THEN 1 ELSE 0 END), 0)::float8 / NULLIF(COUNT(*), 0)
		FROM event`).Scan(&out.MatchRate)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE WHEN resolution = 'rejected' THEN 1 ELSE 0 END), 0)::float8 / NULLIF(COUNT(*), 0)
		FROM merge_candidate`).Scan(&out.ContestedMergeRate)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (j.created_at - s.computed_at)) / 60.0), 0)
		FROM export_job j
		JOIN segment_version s ON s.id = j.segment_version_id
		WHERE s.computed_at IS NOT NULL AND j.created_at >= s.computed_at`).Scan(&out.AvgSegmentToExport)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type TenantAggregateResponse struct {
	NewGuests int  `json:"new_guests"`
	Hidden    bool `json:"hidden"`
}

func (s *AnalyticsService) TenantAggregates(ctx context.Context, tenantID string, claimTenantID string, role domain.Role) (*TenantAggregateResponse, error) {
	if role == domain.RoleTenant {
		if claimTenantID == "" || claimTenantID != tenantID {
			return nil, domain.ErrForbidden
		}
	}
	var n int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT e.guest_id)
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
		WHERE e.tenant_id = $1 AND e.guest_id IS NOT NULL`, tenantID).Scan(&n)
	if err != nil {
		return nil, err
	}
	const k = 5
	if n < k {
		return &TenantAggregateResponse{NewGuests: 0, Hidden: true}, nil
	}
	return &TenantAggregateResponse{NewGuests: n}, nil
}

type DataQualityItem struct {
	DataSourceID  string  `json:"data_source_id"`
	Name          string  `json:"name"`
	LastImportAt  string  `json:"last_import_at,omitempty"`
	RejectedRate  float64 `json:"rejected_rate"`
	PendingMerges int     `json:"pending_merges"`
	Warning       bool    `json:"warning"`
}

func (s *AnalyticsService) DataQuality(ctx context.Context) ([]DataQualityItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT d.id, d.name, d.last_import_at,
			COALESCE((
				SELECT SUM(b.rejected_rows)::float8 / NULLIF(SUM(b.total_rows), 0)
				FROM import_batch b
				WHERE b.data_source_id = d.id AND b.created_at >= now() - interval '30 days'
			), 0),
			COALESCE((
				SELECT COUNT(*) FROM merge_candidate m
				WHERE m.resolution = 'pending'
				  AND m.guest_a_id IN (SELECT guest_id FROM guest_identifier WHERE data_source_id = d.id)
				  OR m.guest_b_id IN (SELECT guest_id FROM guest_identifier WHERE data_source_id = d.id)
			), 0)
		FROM data_source d
		ORDER BY d.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]DataQualityItem, 0)
	for rows.Next() {
		var it DataQualityItem
		var last sql.NullTime
		if err := rows.Scan(&it.DataSourceID, &it.Name, &last, &it.RejectedRate, &it.PendingMerges); err != nil {
			return nil, err
		}
		if last.Valid {
			it.LastImportAt = last.Time.UTC().Format(time.RFC3339)
			it.Warning = time.Since(last.Time) > 7*24*time.Hour || it.RejectedRate > 0.10
		} else {
			it.Warning = true
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
