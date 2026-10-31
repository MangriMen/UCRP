package service

import (
	"context"
	"fmt"
	"time"
)

type TenantAggregateQuery struct {
	TenantID      string
	From          *time.Time
	To            *time.Time
	Threshold     int
	ShowNewGuests bool
}

type TenantAggregateDay struct {
	Date           string `json:"date"`
	VisitsCount    *int   `json:"visitsCount"`
	NewGuestsCount *int   `json:"newGuestsCount"`
}

type TenantAggregateReport struct {
	From                 *string              `json:"from,omitempty"`
	To                   *string              `json:"to,omitempty"`
	VisitsCount          *int                 `json:"visitsCount"`
	GuestsCount          *int                 `json:"guestsCount"`
	NewGuestsCount       *int                 `json:"newGuestsCount"`
	Suppressed           bool                 `json:"suppressed"`
	SuppressionThreshold int                  `json:"suppressionThreshold"`
	Series               []TenantAggregateDay `json:"series"`
}

const dateLayout = "2006-01-02"

const presenceEvents = `
	FROM event e
	JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
	WHERE e.tenant_id = $1 AND e.guest_id IS NOT NULL`

func (s *AnalyticsService) TenantAggregate(ctx context.Context, q TenantAggregateQuery) (*TenantAggregateReport, error) {
	threshold := max(q.Threshold, 5)
	report := &TenantAggregateReport{SuppressionThreshold: threshold, Series: []TenantAggregateDay{}}
	args := []any{q.TenantID}
	window := ""
	if q.From != nil {
		args = append(args, *q.From)
		window += fmt.Sprintf(" AND e.occurred_at >= $%d", len(args))
		from := q.From.Format(dateLayout)
		report.From = &from
	}
	if q.To != nil {
		args = append(args, *q.To)
		window += fmt.Sprintf(" AND e.occurred_at < $%d", len(args))
		to := q.To.AddDate(0, 0, -1).Format(dateLayout)
		report.To = &to
	}

	var visits, guests int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(DISTINCT e.guest_id) `+presenceEvents+window, args...).Scan(&visits, &guests); err != nil {
		return nil, err
	}
	if guests >= threshold {
		report.VisitsCount, report.GuestsCount = &visits, &guests
	} else {
		report.Suppressed = true
	}

	if q.ShowNewGuests {
		fresh, err := s.countNewGuests(ctx, q.TenantID, q.From, q.To)
		if err != nil {
			return nil, err
		}
		if fresh >= threshold {
			report.NewGuestsCount = &fresh
		} else {
			report.Suppressed = true
		}
	}

	if q.From != nil && q.To != nil {
		series, err := s.tenantSeries(ctx, q, threshold)
		if err != nil {
			return nil, err
		}
		report.Series = series
	}
	return report, nil
}

func (s *AnalyticsService) countNewGuests(ctx context.Context, tenantID string, from, to *time.Time) (int, error) {
	args := []any{tenantID}
	window := ""
	if from != nil {
		args = append(args, *from)
		window += fmt.Sprintf(" AND first_at >= $%d", len(args))
	}
	if to != nil {
		args = append(args, *to)
		window += fmt.Sprintf(" AND first_at < $%d", len(args))
	}
	var fresh int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT e.guest_id, MIN(e.occurred_at) AS first_at `+presenceEvents+`
			GROUP BY e.guest_id
		) first_visit WHERE TRUE`+window, args...).Scan(&fresh)
	return fresh, err
}

func (s *AnalyticsService) tenantSeries(ctx context.Context, q TenantAggregateQuery, threshold int) ([]TenantAggregateDay, error) {
	rows, err := s.db.QueryContext(ctx, `
		WITH days AS (
			SELECT d::date AS day FROM generate_series($2::timestamp::date, ($3::timestamp::date - 1), interval '1 day') d
		),
		daily AS (
			SELECT e.occurred_at::date AS day, COUNT(*) AS visits, COUNT(DISTINCT e.guest_id) AS guests
			`+presenceEvents+` AND e.occurred_at >= $2::timestamp AND e.occurred_at < $3::timestamp
			GROUP BY 1
		),
		first_visit AS (
			SELECT e.guest_id, MIN(e.occurred_at)::date AS day `+presenceEvents+`
			GROUP BY e.guest_id
		),
		fresh AS (
			SELECT day, COUNT(*) AS guests FROM first_visit WHERE day >= $2::timestamp::date AND day < $3::timestamp::date GROUP BY day
		)
		SELECT days.day, COALESCE(daily.visits, 0), COALESCE(daily.guests, 0), COALESCE(fresh.guests, 0)
		FROM days
		LEFT JOIN daily ON daily.day = days.day
		LEFT JOIN fresh ON fresh.day = days.day
		ORDER BY days.day`, q.TenantID, *q.From, *q.To)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	series := []TenantAggregateDay{}
	for rows.Next() {
		var (
			day                   time.Time
			visits, guests, fresh int
		)
		if err := rows.Scan(&day, &visits, &guests, &fresh); err != nil {
			return nil, err
		}
		point := TenantAggregateDay{Date: day.Format(dateLayout)}
		if guests >= threshold {
			point.VisitsCount = &visits
		}
		if q.ShowNewGuests && fresh >= threshold {
			point.NewGuestsCount = &fresh
		}
		series = append(series, point)
	}
	return series, rows.Err()
}
