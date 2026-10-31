package repository

import (
	"context"
	"database/sql"
	"errors"
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

func (r *GuestSQL) Search(ctx context.Context, q string, limit, offset int) ([]domain.Guest, error) {
	q = strings.TrimSpace(q)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var (
		rows *sql.Rows
		err  error
	)
	if q == "" {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, COALESCE(display_name, ''), status, COALESCE(merged_into_id, ''), created_at, updated_at, deleted_at
			FROM guest WHERE status = 'active'
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`, limit, offset)
	} else {
		like := strings.ReplaceAll(strings.ReplaceAll(q, "%", `\%`), "_", `\_`)
		rows, err = r.db.QueryContext(ctx, `
			SELECT DISTINCT g.id, COALESCE(g.display_name, ''), g.status, COALESCE(g.merged_into_id, ''), g.created_at, g.updated_at, g.deleted_at
			FROM guest g
			LEFT JOIN guest_identifier i ON i.guest_id = g.id
			WHERE g.status = 'active' AND (
				g.display_name ILIKE '%' || $1 || '%'
				OR i.value_normalized = $1
				OR i.value_normalized LIKE '%' || $1 || '%'
			)
			ORDER BY g.created_at DESC
			LIMIT $2 OFFSET $3`, like, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.Guest, 0)
	for rows.Next() {
		var g domain.Guest
		if err := rows.Scan(&g.ID, &g.DisplayName, &g.Status, &g.MergedIntoID, &g.CreatedAt, &g.UpdatedAt, &g.DeletedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
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
