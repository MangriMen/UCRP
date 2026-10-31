package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

const (
	defaultTenantPageSize = 50
	maxTenantPageSize     = 200
	minTenantThreshold    = 5
)

var tenantSorts = map[string]string{
	"name":            "t.name ASC",
	"floor":           "t.floor ASC NULLS LAST",
	"priority":        "t.priority ASC",
	"createdAt":       "t.created_at ASC",
	"guestsCount":     "guests_count ASC NULLS LAST",
	"visitsCount":     "visits_count ASC NULLS LAST",
	"newGuestsCount":  "new_guests_count ASC NULLS LAST",
	"-name":           "t.name DESC",
	"-floor":          "t.floor DESC NULLS LAST",
	"-priority":       "t.priority DESC",
	"-createdAt":      "t.created_at DESC",
	"-guestsCount":    "guests_count DESC NULLS LAST",
	"-visitsCount":    "visits_count DESC NULLS LAST",
	"-newGuestsCount": "new_guests_count DESC NULLS LAST",
}

func (r *TenantSQL) Search(ctx context.Context, filter domain.TenantFilter, limit, offset int) ([]domain.TenantListEntry, int, error) {
	if limit <= 0 {
		limit = defaultTenantPageSize
	}
	if limit > maxTenantPageSize {
		limit = maxTenantPageSize
	}
	if offset < 0 {
		offset = 0
	}
	order, ok := tenantSorts[filter.Sort]
	if !ok {
		order = tenantSorts["name"]
	}
	threshold := filter.KThreshold
	if threshold < minTenantThreshold {
		threshold = minTenantThreshold
	}

	var args []any
	where := []string{"t.deleted_at IS NULL"}
	add := func(condition string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(condition, len(args)))
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		add("(t.name ILIKE '%%' || $%[1]d || '%%' OR t.code ILIKE '%%' || $%[1]d || '%%')", escapeLike(q))
	}
	if filter.CategoryID != "" {
		add("t.category_id = $%d", filter.CategoryID)
	}
	if filter.Floor != nil {
		add("t.floor = $%d", *filter.Floor)
	}
	if filter.IsActive != nil {
		add("t.is_active = $%d", *filter.IsActive)
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenant t WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	var visitWindow, firstWindow strings.Builder
	addWindow := func(column string, sink *strings.Builder) {
		if filter.PeriodFrom != nil {
			args = append(args, *filter.PeriodFrom)
			fmt.Fprintf(sink, " AND %s >= $%d", column, len(args))
		}
		if filter.PeriodTo != nil {
			args = append(args, *filter.PeriodTo)
			fmt.Fprintf(sink, " AND %s < $%d", column, len(args))
		}
	}
	addWindow("occurred_at", &visitWindow)
	addWindow("first_at", &firstWindow)
	args = append(args, threshold, limit, offset)
	thresholdArg, limitArg, offsetArg := len(args)-2, len(args)-1, len(args)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		WITH presence AS (
			SELECT e.tenant_id, e.guest_id, e.occurred_at
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			WHERE e.guest_id IS NOT NULL
		),
		in_period AS (
			SELECT tenant_id, COUNT(*) AS visits, COUNT(DISTINCT guest_id) AS guests
			FROM presence
			WHERE TRUE%[1]s
			GROUP BY tenant_id
		),
		first_visit AS (
			SELECT tenant_id, guest_id, MIN(occurred_at) AS first_at
			FROM presence
			GROUP BY tenant_id, guest_id
		),
		fresh AS (
			SELECT tenant_id, COUNT(*) AS guests
			FROM first_visit
			WHERE TRUE%[2]s
			GROUP BY tenant_id
		)
		SELECT t.id, t.code, t.name, t.mall_id, t.category_id, t.kind, t.floor, t.priority, t.is_active,
			t.created_at, t.updated_at, c.name,
			CASE WHEN COALESCE(p.guests, 0) < $%[3]d THEN NULL ELSE p.guests END AS guests_count,
			CASE WHEN COALESCE(p.guests, 0) < $%[3]d THEN NULL ELSE p.visits END AS visits_count,
			CASE WHEN COALESCE(f.guests, 0) < $%[3]d THEN NULL ELSE f.guests END AS new_guests_count,
			(SELECT COUNT(*) FROM data_source d WHERE d.tenant_id = t.id AND d.is_active) AS sources_count,
			(SELECT COUNT(*) FROM data_source d WHERE d.tenant_id = t.id AND d.is_active AND COALESCE((
				SELECT b.status FROM import_batch b WHERE b.data_source_id = d.id ORDER BY b.created_at DESC LIMIT 1
			), '') IN ('failed', 'completed_with_errors')) AS problem_sources_count
		FROM tenant t
		JOIN tenant_category c ON c.id = t.category_id
		LEFT JOIN in_period p ON p.tenant_id = t.id
		LEFT JOIN fresh f ON f.tenant_id = t.id
		WHERE %[4]s
		ORDER BY %[5]s, t.name ASC, t.id ASC
		LIMIT $%[6]d OFFSET $%[7]d`,
		visitWindow.String(), firstWindow.String(), thresholdArg, whereSQL, order, limitArg, offsetArg,
	), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.TenantListEntry, 0)
	for rows.Next() {
		var (
			entry                 domain.TenantListEntry
			created, updated      time.Time
			guests, visits, fresh *int
		)
		if err := rows.Scan(
			&entry.ID, &entry.Code, &entry.Name, &entry.MallID, &entry.CategoryID, &entry.Kind, &entry.Floor,
			&entry.Priority, &entry.IsActive, &created, &updated, &entry.CategoryName,
			&guests, &visits, &fresh, &entry.Stats.SourcesCount, &entry.Stats.ProblemSourcesCount,
		); err != nil {
			return nil, 0, err
		}
		entry.CreatedAt, entry.UpdatedAt = created, updated
		entry.Stats.GuestsCount, entry.Stats.VisitsCount, entry.Stats.NewGuestsCount = guests, visits, fresh
		out = append(out, entry)
	}
	return out, total, rows.Err()
}
