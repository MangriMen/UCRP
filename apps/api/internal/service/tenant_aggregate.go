package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

type TenantAggregateMetrics struct {
	NewGuests      bool
	RepeatVisits   bool
	Revenue        bool
	AvgCheck       bool
	FirstPurchases bool
}

type TenantAggregateQuery struct {
	TenantID  string
	From      *time.Time
	To        *time.Time
	Threshold int
	Metrics   TenantAggregateMetrics
}

type TenantAggregateDay struct {
	Date              string `json:"date"`
	VisitsCount       *int   `json:"visitsCount"`
	NewGuestsCount    *int   `json:"newGuestsCount"`
	RepeatVisitsCount *int   `json:"repeatVisitsCount"`
	RevenueTotalMinor *int64 `json:"revenueTotalMinor"`
}

type TenantAggregateReport struct {
	From                 *string              `json:"from,omitempty"`
	To                   *string              `json:"to,omitempty"`
	VisitsCount          *int                 `json:"visitsCount"`
	GuestsCount          *int                 `json:"guestsCount"`
	NewGuestsCount       *int                 `json:"newGuestsCount"`
	RepeatVisitsCount    *int                 `json:"repeatVisitsCount"`
	RevenueTotalMinor    *int64               `json:"revenueTotalMinor"`
	AvgCheckMinor        *int64               `json:"avgCheckMinor"`
	FirstPurchasesCount  *int                 `json:"firstPurchasesCount"`
	Suppressed           bool                 `json:"suppressed"`
	SuppressionThreshold int                  `json:"suppressionThreshold"`
	Series               []TenantAggregateDay `json:"series"`
}

const dateLayout = "2006-01-02"

const presenceEvents = `
	FROM event e
	JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
	WHERE e.tenant_id = $1 AND e.guest_id IS NOT NULL`

const numberedPresence = `
	SELECT e.id, e.guest_id, e.occurred_at,
		ROW_NUMBER() OVER (PARTITION BY e.guest_id ORDER BY e.occurred_at, e.id) AS visit_number` + presenceEvents

const purchaseEvents = `
	FROM event e
	JOIN event_type et ON et.id = e.event_type_id AND et.code = 'purchase'
	WHERE e.tenant_id = $1 AND e.guest_id IS NOT NULL AND e.amount_minor IS NOT NULL`

const guestHistory = `
	SELECT e.guest_id, MIN(e.occurred_at) AS first_at,
		MIN(e.occurred_at) FILTER (WHERE et.counts_as_presence) AS first_presence_at,
		MIN(e.occurred_at) FILTER (WHERE et.code = 'purchase') AS first_purchase_at
	FROM event e
	JOIN event_type et ON et.id = e.event_type_id
	WHERE e.tenant_id = $1 AND e.guest_id IS NOT NULL
	GROUP BY e.guest_id`

type aggregateReveal struct {
	threshold  int
	suppressed bool
}

func (r *aggregateReveal) count(value, guests int) *int {
	if guests < r.threshold {
		r.suppressed = true
		return nil
	}
	return &value
}

func repeatVisitGuests(repeatGuests, firstVisits int) int {
	if firstVisits == 0 {
		return repeatGuests
	}
	return min(repeatGuests, firstVisits)
}

func (r *aggregateReveal) amount(value int64, guests int) *int64 {
	if guests < r.threshold {
		r.suppressed = true
		return nil
	}
	return &value
}

func periodWindow(column string, from, to *time.Time, args []any) (string, []any) {
	window := ""
	if from != nil {
		args = append(args, *from)
		window += fmt.Sprintf(" AND %s >= $%d", column, len(args))
	}
	if to != nil {
		args = append(args, *to)
		window += fmt.Sprintf(" AND %s < $%d", column, len(args))
	}
	return window, args
}

func (s *AnalyticsService) TenantAggregate(ctx context.Context, q TenantAggregateQuery) (*TenantAggregateReport, error) {
	reveal := &aggregateReveal{threshold: max(q.Threshold, 5)}
	report := &TenantAggregateReport{SuppressionThreshold: reveal.threshold, Series: []TenantAggregateDay{}}
	if q.From != nil {
		from := q.From.Format(dateLayout)
		report.From = &from
	}
	if q.To != nil {
		to := q.To.AddDate(0, 0, -1).Format(dateLayout)
		report.To = &to
	}

	window, args := periodWindow("e.occurred_at", q.From, q.To, []any{q.TenantID})
	var visits, guests int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(DISTINCT e.guest_id) `+presenceEvents+window, args...).Scan(&visits, &guests); err != nil {
		return nil, err
	}
	report.VisitsCount = reveal.count(visits, guests)
	report.GuestsCount = reveal.count(guests, guests)

	var fresh, firstBuyers int
	switch {
	case q.Metrics.NewGuests && q.Metrics.FirstPurchases:
		var err error
		if fresh, firstBuyers, err = s.countGuestFirsts(ctx, q); err != nil {
			return nil, err
		}
	case q.Metrics.NewGuests:
		var err error
		if fresh, err = s.countNewGuests(ctx, q); err != nil {
			return nil, err
		}
	case q.Metrics.FirstPurchases:
		var err error
		if firstBuyers, err = s.countFirstPurchases(ctx, q); err != nil {
			return nil, err
		}
	}
	if q.Metrics.NewGuests {
		report.NewGuestsCount = reveal.count(fresh, fresh)
	}
	if q.Metrics.RepeatVisits {
		repeats, repeatGuests, err := s.countRepeatVisits(ctx, q)
		if err != nil {
			return nil, err
		}
		report.RepeatVisitsCount = reveal.count(repeats, repeatVisitGuests(repeatGuests, visits-repeats))
	}
	if q.Metrics.Revenue || q.Metrics.AvgCheck {
		revenue, purchases, buyers, err := s.sumPurchases(ctx, q)
		if err != nil {
			return nil, err
		}
		if q.Metrics.Revenue {
			report.RevenueTotalMinor = reveal.amount(revenue, buyers)
		}
		if q.Metrics.AvgCheck {
			report.AvgCheckMinor = reveal.amount(averageCheck(revenue, purchases), buyers)
		}
	}
	if q.Metrics.FirstPurchases {
		report.FirstPurchasesCount = reveal.count(firstBuyers, firstBuyers)
	}
	report.Suppressed = reveal.suppressed

	if q.From != nil && q.To != nil {
		series, err := s.tenantSeries(ctx, q, reveal.threshold)
		if err != nil {
			return nil, err
		}
		report.Series = series
	}
	return report, nil
}

func averageCheck(revenue int64, purchases int) int64 {
	if purchases == 0 {
		return 0
	}
	return int64(math.Round(float64(revenue) / float64(purchases)))
}

func (s *AnalyticsService) countNewGuests(ctx context.Context, q TenantAggregateQuery) (int, error) {
	window, args := periodWindow("first_at", q.From, q.To, []any{q.TenantID})
	var fresh int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT e.guest_id, MIN(e.occurred_at) AS first_at `+presenceEvents+`
			GROUP BY e.guest_id
		) first_visit WHERE TRUE`+window, args...).Scan(&fresh)
	return fresh, err
}

func (s *AnalyticsService) countGuestFirsts(ctx context.Context, q TenantAggregateQuery) (fresh, firstBuyers int, err error) {
	window, args := periodWindow("first_presence_at", q.From, q.To, []any{q.TenantID})
	purchaseWindow := strings.ReplaceAll(window, "first_presence_at", "first_at")
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FILTER (WHERE first_presence_at IS NOT NULL`+window+`),
			COUNT(*) FILTER (WHERE first_purchase_at = first_at`+purchaseWindow+`)
		FROM (`+guestHistory+`) history`, args...).Scan(&fresh, &firstBuyers)
	return fresh, firstBuyers, err
}

func (s *AnalyticsService) countRepeatVisits(ctx context.Context, q TenantAggregateQuery) (visits, guests int, err error) {
	window, args := periodWindow("occurred_at", q.From, q.To, []any{q.TenantID})
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(DISTINCT guest_id) FROM (`+numberedPresence+`) numbered
		WHERE visit_number > 1`+window, args...).Scan(&visits, &guests)
	return visits, guests, err
}

func (s *AnalyticsService) sumPurchases(ctx context.Context, q TenantAggregateQuery) (revenue int64, purchases, buyers int, err error) {
	window, args := periodWindow("e.occurred_at", q.From, q.To, []any{q.TenantID})
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(e.amount_minor), 0), COUNT(*), COUNT(DISTINCT e.guest_id) `+purchaseEvents+window,
		args...).Scan(&revenue, &purchases, &buyers)
	return revenue, purchases, buyers, err
}

func (s *AnalyticsService) countFirstPurchases(ctx context.Context, q TenantAggregateQuery) (int, error) {
	window, args := periodWindow("first_at", q.From, q.To, []any{q.TenantID})
	var buyers int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (`+guestHistory+`) history
		WHERE first_purchase_at = first_at`+window, args...).Scan(&buyers)
	return buyers, err
}

func (s *AnalyticsService) tenantSeries(ctx context.Context, q TenantAggregateQuery, threshold int) ([]TenantAggregateDay, error) {
	ctes := `
		days AS (
			SELECT d::date AS day FROM generate_series($2::timestamp::date, ($3::timestamp::date - 1), interval '1 day') d
		),
		daily AS (
			SELECT e.occurred_at::date AS day, COUNT(*) AS visits, COUNT(DISTINCT e.guest_id) AS guests
			` + presenceEvents + ` AND e.occurred_at >= $2::timestamp AND e.occurred_at < $3::timestamp
			GROUP BY 1
		),
		first_visit AS (
			SELECT e.guest_id, MIN(e.occurred_at)::date AS day ` + presenceEvents + `
			GROUP BY e.guest_id
		),
		fresh AS (
			SELECT day, COUNT(*) AS guests FROM first_visit WHERE day >= $2::timestamp::date AND day < $3::timestamp::date GROUP BY day
		)`
	joins := `
		LEFT JOIN daily ON daily.day = days.day
		LEFT JOIN fresh ON fresh.day = days.day`
	repeatColumns := "0, 0"
	if q.Metrics.RepeatVisits {
		ctes += `,
		repeats AS (
			SELECT occurred_at::date AS day, COUNT(*) AS visits, COUNT(DISTINCT guest_id) AS guests
			FROM (` + numberedPresence + `) numbered
			WHERE visit_number > 1 AND occurred_at >= $2::timestamp AND occurred_at < $3::timestamp
			GROUP BY 1
		)`
		joins += `
		LEFT JOIN repeats ON repeats.day = days.day`
		repeatColumns = "COALESCE(repeats.visits, 0), COALESCE(repeats.guests, 0)"
	}
	revenueColumns := "0, 0"
	if q.Metrics.Revenue {
		ctes += `,
		revenue AS (
			SELECT e.occurred_at::date AS day, SUM(e.amount_minor) AS total, COUNT(DISTINCT e.guest_id) AS guests
			` + purchaseEvents + ` AND e.occurred_at >= $2::timestamp AND e.occurred_at < $3::timestamp
			GROUP BY 1
		)`
		joins += `
		LEFT JOIN revenue ON revenue.day = days.day`
		revenueColumns = "COALESCE(revenue.total, 0), COALESCE(revenue.guests, 0)"
	}
	rows, err := s.db.QueryContext(ctx, `
		WITH `+ctes+`
		SELECT days.day, COALESCE(daily.visits, 0), COALESCE(daily.guests, 0), COALESCE(fresh.guests, 0),
			`+repeatColumns+`, `+revenueColumns+`
		FROM days`+joins+`
		ORDER BY days.day`, q.TenantID, *q.From, *q.To)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	series := []TenantAggregateDay{}
	for rows.Next() {
		var (
			day                                       time.Time
			visits, guests, fresh                     int
			repeatVisits, repeatGuests, revenueBuyers int
			revenue                                   int64
		)
		if err := rows.Scan(&day, &visits, &guests, &fresh, &repeatVisits, &repeatGuests, &revenue, &revenueBuyers); err != nil {
			return nil, err
		}
		reveal := &aggregateReveal{threshold: threshold}
		point := TenantAggregateDay{Date: day.Format(dateLayout), VisitsCount: reveal.count(visits, guests)}
		if q.Metrics.NewGuests {
			point.NewGuestsCount = reveal.count(fresh, fresh)
		}
		if q.Metrics.RepeatVisits {
			point.RepeatVisitsCount = reveal.count(repeatVisits, repeatVisitGuests(repeatGuests, visits-repeatVisits))
		}
		if q.Metrics.Revenue {
			point.RevenueTotalMinor = reveal.amount(revenue, revenueBuyers)
		}
		series = append(series, point)
	}
	return series, rows.Err()
}
