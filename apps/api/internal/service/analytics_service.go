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

type CrossTenantDistribution struct {
	One       int `json:"one"`
	Two       int `json:"two"`
	ThreePlus int `json:"threePlus"`
}

type CrossTenantPoint struct {
	CrossTenantDistribution
	Period          string  `json:"period"`
	CrossTenantRate float64 `json:"crossTenantRate"`
}

type CrossTenantResponse struct {
	From            string                  `json:"from"`
	To              string                  `json:"to"`
	Granularity     string                  `json:"granularity"`
	Distribution    CrossTenantDistribution `json:"distribution"`
	CrossTenantRate float64                 `json:"crossTenantRate"`
	Series          []CrossTenantPoint      `json:"series"`
}

func (s *AnalyticsService) CrossTenantReport(ctx context.Context, from, to time.Time, granularity string) (*CrossTenantResponse, error) {
	if to.IsZero() {
		to = time.Now().UTC()
	}
	if from.IsZero() {
		from = to.AddDate(0, -6, 0)
	}
	if from.After(to) {
		return nil, domain.ErrValidation
	}
	if to.Sub(from) > 24*31*24*time.Hour {
		return nil, domain.ErrValidation
	}
	switch granularity {
	case "", "day", "week", "month":
	default:
		return nil, domain.ErrValidation
	}
	if granularity == "" {
		granularity = "month"
	}

	var dist CrossTenantDistribution
	var rate float64
	err := s.db.QueryRowContext(ctx, `
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
		FROM guest_tenants`, from, to).Scan(&dist.One, &dist.Two, &dist.ThreePlus, &rate)
	if err != nil {
		return nil, err
	}

	series, err := s.crossTenantSeries(ctx, from, to, granularity)
	if err != nil {
		return nil, err
	}
	return &CrossTenantResponse{
		From:            from.UTC().Format(time.RFC3339),
		To:              to.UTC().Format(time.RFC3339),
		Granularity:     granularity,
		Distribution:    dist,
		CrossTenantRate: rate,
		Series:          series,
	}, nil
}

func (s *AnalyticsService) crossTenantSeries(ctx context.Context, from, to time.Time, granularity string) ([]CrossTenantPoint, error) {
	var bucketSQL string
	switch granularity {
	case "day":
		bucketSQL = `to_char(e.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')`
	case "week":
		bucketSQL = `to_char(date_trunc('week', e.occurred_at AT TIME ZONE 'UTC'), 'IYYY-IW')`
	default:
		bucketSQL = `to_char(e.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM')`
	}
	rows, err := s.db.QueryContext(ctx, `
		WITH ev AS (
			SELECT e.guest_id, e.tenant_id, `+bucketSQL+` AS period
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			WHERE e.guest_id IS NOT NULL AND e.occurred_at >= $1 AND e.occurred_at <= $2
		),
		per_guest AS (
			SELECT period, guest_id, COUNT(DISTINCT tenant_id) AS tn
			FROM ev
			GROUP BY period, guest_id
		)
		SELECT period,
			COALESCE(SUM(CASE WHEN tn = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tn = 2 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tn >= 3 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tn >= 2 THEN 1 ELSE 0 END), 0)::float8 / NULLIF(COUNT(*), 0)
		FROM per_guest
		GROUP BY period
		ORDER BY period`, from, to)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]CrossTenantPoint, 0)
	for rows.Next() {
		var p CrossTenantPoint
		var rate sql.NullFloat64
		if err := rows.Scan(&p.Period, &p.One, &p.Two, &p.ThreePlus, &rate); err != nil {
			return nil, err
		}
		if rate.Valid {
			p.CrossTenantRate = rate.Float64
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type MetricsResponse struct {
	From                         string  `json:"from"`
	To                           string  `json:"to"`
	MatchRate                    float64 `json:"matchRate"`
	DisputedMergeRate            float64 `json:"disputedMergeRate"`
	SegmentToExportMedianSeconds float64 `json:"segmentToExportMedianSeconds"`
	SegmentToExportP95Seconds    float64 `json:"segmentToExportP95Seconds"`
	DisputedRateReliable         bool
	ExportLatencyReliable        bool
}

func (s *AnalyticsService) Metrics(ctx context.Context, from, to time.Time) (*MetricsResponse, error) {
	if to.IsZero() {
		to = time.Now().UTC()
	}
	if from.IsZero() {
		from = to.AddDate(0, -6, 0)
	}
	if from.After(to) {
		return nil, domain.ErrValidation
	}
	out := &MetricsResponse{
		From: from.UTC().Format(time.RFC3339),
		To:   to.UTC().Format(time.RFC3339),
	}
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN guest_id IS NOT NULL THEN 1 ELSE 0 END)::float8 / NULLIF(COUNT(*), 0), 0)
		FROM event`).Scan(&out.MatchRate)
	if err != nil {
		return nil, err
	}
	var mergeTotal, mergeRejected int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE resolution = 'reverted')
		FROM merge_candidate`).Scan(&mergeTotal, &mergeRejected)
	if err != nil {
		return nil, err
	}
	if mergeTotal > 0 {
		out.DisputedMergeRate = float64(mergeRejected) / float64(mergeTotal)
		out.DisputedRateReliable = mergeTotal >= 100
	}
	err = s.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(percentile_cont(0.5) WITHIN GROUP (
				ORDER BY EXTRACT(EPOCH FROM (j.created_at - s.computed_at))
			), 0),
			COALESCE(percentile_cont(0.95) WITHIN GROUP (
				ORDER BY EXTRACT(EPOCH FROM (j.created_at - s.computed_at))
			), 0),
			COUNT(*)
		FROM export_job j
		JOIN segment_version s ON s.id = j.segment_version_id
		WHERE s.computed_at IS NOT NULL AND j.created_at >= s.computed_at`).Scan(&out.SegmentToExportMedianSeconds, &out.SegmentToExportP95Seconds, &mergeTotal)
	if err != nil {
		return nil, err
	}
	out.ExportLatencyReliable = mergeTotal >= 30
	return out, nil
}

type TenantAggregateResponse struct {
	NewGuests  int  `json:"newGuests"`
	Suppressed bool `json:"suppressed"`
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
		return &TenantAggregateResponse{NewGuests: 0, Suppressed: true}, nil
	}
	return &TenantAggregateResponse{NewGuests: n}, nil
}

type DataQualityItem struct {
	DataSourceID           string  `json:"dataSourceId"`
	Name                   string  `json:"name"`
	LastImportAt           string  `json:"lastImportAt,omitempty"`
	RejectedShare30d       float64 `json:"rejectedShare30d"`
	PendingMergeCandidates int     `json:"pendingMergeCandidates"`
	HasWarning             bool    `json:"hasWarning"`
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
				  AND (
				    m.guest_a_id IN (SELECT guest_id FROM guest_identifier WHERE data_source_id = d.id)
				    OR m.guest_b_id IN (SELECT guest_id FROM guest_identifier WHERE data_source_id = d.id)
				  )
			), 0)
		FROM data_source d
		ORDER BY d.name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]DataQualityItem, 0)
	for rows.Next() {
		var it DataQualityItem
		var last sql.NullTime
		if err := rows.Scan(&it.DataSourceID, &it.Name, &last, &it.RejectedShare30d, &it.PendingMergeCandidates); err != nil {
			return nil, err
		}
		if last.Valid {
			it.LastImportAt = last.Time.UTC().Format(time.RFC3339)
			it.HasWarning = time.Since(last.Time) > 7*24*time.Hour || it.RejectedShare30d > 0.10
		} else {
			it.HasWarning = true
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
