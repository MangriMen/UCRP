package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type GuestSQL struct {
	db *sql.DB
}

func NewGuestSQL(db *sql.DB) *GuestSQL {
	return &GuestSQL{db: db}
}

func (r *GuestSQL) GetByID(ctx context.Context, id string) (*domain.Guest, error) {
	var g domain.Guest
	err := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(display_name, ''), status, COALESCE(merged_into_id, ''), created_at, updated_at, deleted_at
		FROM guest WHERE id = $1`, id).Scan(
		&g.ID, &g.DisplayName, &g.Status, &g.MergedIntoID, &g.CreatedAt, &g.UpdatedAt, &g.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

var guestSorts = map[string]string{
	"createdAt":         "g.created_at ASC, g.id ASC",
	"-createdAt":        "g.created_at DESC, g.id DESC",
	"eventCount":        "COALESCE(s.event_count, 0) ASC, g.id ASC",
	"-eventCount":       "COALESCE(s.event_count, 0) DESC, g.id DESC",
	"totalAmountMinor":  "COALESCE(s.purchase_total, 0) ASC, g.id ASC",
	"-totalAmountMinor": "COALESCE(s.purchase_total, 0) DESC, g.id DESC",
	"firstSeenAt":       "s.first_seen_at ASC NULLS LAST, g.id ASC",
	"-firstSeenAt":      "s.first_seen_at DESC NULLS LAST, g.id DESC",
	"lastSeenAt":        "s.last_seen_at ASC NULLS LAST, g.id ASC",
	"-lastSeenAt":       "s.last_seen_at DESC NULLS LAST, g.id DESC",
}

const guestStatsJoin = `
	FROM guest g
	LEFT JOIN (
		SELECT e.guest_id,
			COUNT(*) AS event_count,
			MIN(e.occurred_at) AS first_seen_at,
			MAX(e.occurred_at) AS last_seen_at,
			COUNT(DISTINCT e.tenant_id) FILTER (WHERE et.counts_as_presence) AS tenant_count,
			COUNT(*) FILTER (WHERE et.code = 'purchase') AS purchase_count,
			COALESCE(SUM(e.amount_minor) FILTER (WHERE et.code = 'purchase'), 0) AS purchase_total,
			COALESCE(MAX(e.currency) FILTER (WHERE et.code = 'purchase'), '') AS currency
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id
		WHERE e.guest_id IS NOT NULL
		GROUP BY e.guest_id
	) s ON s.guest_id = g.id`

func (r *GuestSQL) Search(ctx context.Context, filter domain.GuestFilter, limit, offset int) ([]domain.GuestListEntry, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	order, ok := guestSorts[filter.Sort]
	if !ok {
		order = guestSorts["-createdAt"]
	}
	status := filter.Status
	if status == "" {
		status = domain.GuestActive
	}
	args := []any{string(status)}
	where := []string{"g.status = $1"}
	add := func(condition string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(condition, len(args)))
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		args = append(args, q, escapeLike(q))
		raw, like := len(args)-1, len(args)
		where = append(where, fmt.Sprintf(`(
			g.display_name ILIKE '%%' || $%[2]d || '%%'
			OR EXISTS (
				SELECT 1 FROM guest_identifier i
				WHERE i.guest_id = g.id AND (i.value_normalized = $%[1]d OR i.value_normalized LIKE '%%' || $%[2]d || '%%')
			)
		)`, raw, like))
	}
	var eventConditions []string
	addEvent := func(condition string, value any) {
		args = append(args, value)
		eventConditions = append(eventConditions, fmt.Sprintf(condition, len(args)))
	}
	if filter.SourceID != "" {
		addEvent("ev.data_source_id = $%d", filter.SourceID)
	}
	if filter.TenantID != "" {
		addEvent("ev.tenant_id = $%d", filter.TenantID)
	}
	if filter.PeriodFrom != nil {
		addEvent("ev.occurred_at >= $%d", *filter.PeriodFrom)
	}
	if filter.PeriodTo != nil {
		addEvent("ev.occurred_at < $%d", *filter.PeriodTo)
	}
	if len(eventConditions) > 0 {
		where = append(where, "EXISTS (SELECT 1 FROM event ev WHERE ev.guest_id = g.id AND "+strings.Join(eventConditions, " AND ")+")")
	}
	if filter.SegmentID != "" {
		add(`EXISTS (
			SELECT 1 FROM segment_member m
			JOIN segment_version v ON v.id = m.segment_version_id
			JOIN segment sg ON sg.id = v.segment_id AND sg.current_version_no = v.version_no
			WHERE m.guest_id = g.id AND sg.id = $%d
		)`, filter.SegmentID)
	}
	if filter.FirstSeenFrom != nil {
		add("s.first_seen_at >= $%d", *filter.FirstSeenFrom)
	}
	if filter.LastSeenTo != nil {
		add("s.last_seen_at < $%d", *filter.LastSeenTo)
	}
	if filter.HasPurchases != nil {
		if *filter.HasPurchases {
			where = append(where, "COALESCE(s.purchase_count, 0) > 0")
		} else {
			where = append(where, "COALESCE(s.purchase_count, 0) = 0")
		}
	}
	from := guestStatsJoin + " WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) "+from, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, COALESCE(g.display_name, ''), g.status, COALESCE(g.merged_into_id, ''), g.created_at, g.updated_at, g.deleted_at,
			COALESCE(s.tenant_count, 0), COALESCE(s.event_count, 0), s.first_seen_at, s.last_seen_at,
			COALESCE(s.purchase_count, 0), COALESCE(s.purchase_total, 0), COALESCE(s.currency, '')`+from+`
		ORDER BY `+order+fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.GuestListEntry, 0)
	for rows.Next() {
		var (
			entry       domain.GuestListEntry
			first, last sql.NullTime
		)
		if err := rows.Scan(
			&entry.ID, &entry.DisplayName, &entry.Status, &entry.MergedIntoID, &entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt,
			&entry.TenantCount, &entry.Stats.EventCount, &first, &last,
			&entry.Stats.PurchaseCount, &entry.Stats.PurchaseTotal, &entry.Stats.Currency,
		); err != nil {
			return nil, 0, err
		}
		if first.Valid {
			entry.Stats.FirstSeenAt = &first.Time
		}
		if last.Valid {
			entry.Stats.LastSeenAt = &last.Time
		}
		out = append(out, entry)
	}
	return out, total, rows.Err()
}

func escapeLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(value)
}

func (r *GuestSQL) Insert(ctx context.Context, g *domain.Guest) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO guest (id, display_name, status, merged_into_id, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		g.ID, nullIfEmpty(g.DisplayName), string(g.Status), nullIfEmpty(g.MergedIntoID), g.CreatedAt, g.UpdatedAt, g.DeletedAt,
	)
	return err
}

func (r *GuestSQL) TenantCount(ctx context.Context, guestID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT t.id)
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id
		JOIN tenant t ON t.id = e.tenant_id
		WHERE e.guest_id = $1 AND et.counts_as_presence`, guestID).Scan(&n)
	return n, err
}

func (r *GuestSQL) TenantNames(ctx context.Context, guestID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT t.name
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id
		JOIN tenant t ON t.id = e.tenant_id
		WHERE e.guest_id = $1 AND et.counts_as_presence
		ORDER BY t.name`, guestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

func (r *GuestSQL) MarkMerged(ctx context.Context, loserID, winnerID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE guest SET status = 'merged', merged_into_id = $2, updated_at = now()
		WHERE id = $1`, loserID, winnerID)
	return err
}

func (r *GuestSQL) MarkActive(ctx context.Context, guestID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE guest SET status = 'active', merged_into_id = NULL, updated_at = now()
		WHERE id = $1`, guestID)
	return err
}

func (r *GuestSQL) SoftDelete(ctx context.Context, id string, now time.Time) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE guest SET status = 'deleted', deleted_at = $2, updated_at = $2 WHERE id = $1`, id, now)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *GuestSQL) TenantSummaries(ctx context.Context, guestID string) ([]domain.GuestTenantSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.name, c.name, t.floor,
			COUNT(*) FILTER (WHERE et.counts_as_presence),
			COUNT(*) FILTER (WHERE et.code = 'purchase'),
			COALESCE(SUM(e.amount_minor) FILTER (WHERE et.code = 'purchase'), 0),
			COALESCE(MAX(e.currency) FILTER (WHERE et.code = 'purchase'), ''),
			MAX(e.occurred_at) FILTER (WHERE et.counts_as_presence)
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id
		JOIN tenant t ON t.id = e.tenant_id
		JOIN tenant_category c ON c.id = t.category_id
		WHERE e.guest_id = $1
		GROUP BY t.id, t.name, c.name, t.floor
		ORDER BY t.name`, guestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.GuestTenantSummary, 0)
	for rows.Next() {
		var (
			s         domain.GuestTenantSummary
			lastVisit sql.NullTime
		)
		if err := rows.Scan(&s.TenantID, &s.TenantName, &s.CategoryName, &s.Floor, &s.EventCount, &s.PurchaseCount, &s.PurchaseTotal, &s.Currency, &lastVisit); err != nil {
			return nil, err
		}
		if lastVisit.Valid {
			s.LastVisitAt = &lastVisit.Time
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *GuestSQL) SegmentSummaries(ctx context.Context, guestID string) ([]domain.GuestSegmentSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.id, s.name, v.version_no
		FROM segment_member m
		JOIN segment_version v ON v.id = m.segment_version_id
		JOIN segment s ON s.id = v.segment_id AND s.current_version_no = v.version_no
		WHERE m.guest_id = $1
		ORDER BY s.name`, guestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.GuestSegmentSummary, 0)
	for rows.Next() {
		var s domain.GuestSegmentSummary
		if err := rows.Scan(&s.SegmentID, &s.Name, &s.VersionNo); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *GuestSQL) Stats(ctx context.Context, guestID string) (domain.GuestStats, error) {
	var (
		stats       domain.GuestStats
		first, last sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), MIN(e.occurred_at), MAX(e.occurred_at),
			COUNT(*) FILTER (WHERE et.code = 'purchase'),
			COALESCE(SUM(e.amount_minor) FILTER (WHERE et.code = 'purchase'), 0),
			COALESCE(MAX(e.currency) FILTER (WHERE et.code = 'purchase'), '')
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id
		WHERE e.guest_id = $1`, guestID).
		Scan(&stats.EventCount, &first, &last, &stats.PurchaseCount, &stats.PurchaseTotal, &stats.Currency)
	if err != nil {
		return domain.GuestStats{}, err
	}
	if first.Valid {
		stats.FirstSeenAt = &first.Time
	}
	if last.Valid {
		stats.LastSeenAt = &last.Time
	}
	return stats, nil
}
