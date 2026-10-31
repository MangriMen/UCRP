package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type EventSQL struct {
	db *sql.DB
}

func NewEventSQL(db *sql.DB) *EventSQL {
	return &EventSQL{db: db}
}

var guestEventSorts = map[string]string{
	"occurredAt":   "e.occurred_at ASC, e.id ASC",
	"-occurredAt":  "e.occurred_at DESC, e.id DESC",
	"amountMinor":  "e.amount_minor ASC NULLS LAST, e.id ASC",
	"-amountMinor": "e.amount_minor DESC NULLS LAST, e.id DESC",
}

func (r *EventSQL) ListByGuest(ctx context.Context, guestID string, filter domain.EventFilter, limit, offset int) ([]domain.GuestEvent, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	order, ok := guestEventSorts[filter.Sort]
	if !ok {
		order = guestEventSorts["-occurredAt"]
	}
	where := []string{"e.guest_id = $1"}
	args := []any{guestID}
	add := func(condition string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(condition, len(args)))
	}
	if filter.TypeCode != "" {
		add("et.code = $%d", filter.TypeCode)
	}
	if filter.TenantID != "" {
		add("e.tenant_id = $%d", filter.TenantID)
	}
	if filter.From != nil {
		add("e.occurred_at >= $%d", *filter.From)
	}
	if filter.To != nil {
		add("e.occurred_at < $%d", *filter.To)
	}
	if filter.AmountMin != nil {
		add("e.amount_minor >= $%d", *filter.AmountMin)
	}
	if filter.AmountMax != nil {
		add("e.amount_minor <= $%d", *filter.AmountMax)
	}
	condition := strings.Join(where, " AND ")
	from := `
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id
		JOIN tenant t ON t.id = e.tenant_id
		LEFT JOIN project p ON p.id = e.project_id
		WHERE ` + condition

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) "+from, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, COALESCE(e.guest_id, ''), COALESCE(e.origin_guest_id, ''), e.tenant_id, COALESCE(e.project_id, ''),
			e.event_type_id, e.data_source_id, COALESCE(e.import_batch_id, ''), e.external_id, e.occurred_at,
			e.amount_minor, e.currency, COALESCE(e.match_keys, ''), COALESCE(e.raw_payload, ''), e.created_at,
			et.code, t.name, COALESCE(p.name, '')`+from+`
		ORDER BY `+order+fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.GuestEvent, 0)
	for rows.Next() {
		var (
			e      domain.GuestEvent
			amount sql.NullInt64
		)
		if err := rows.Scan(
			&e.ID, &e.GuestID, &e.OriginGuestID, &e.TenantID, &e.ProjectID,
			&e.EventTypeID, &e.DataSourceID, &e.ImportBatchID, &e.ExternalID, &e.OccurredAt,
			&amount, &e.Currency, &e.MatchKeys, &e.RawPayload, &e.CreatedAt,
			&e.EventTypeCode, &e.TenantName, &e.ProjectName,
		); err != nil {
			return nil, 0, err
		}
		if amount.Valid {
			v := amount.Int64
			e.AmountMinor = &v
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func (r *EventSQL) ActionLog(ctx context.Context, guestID string, limit int) ([]domain.GuestActionLogEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.occurred_at, t.name, et.code, e.amount_minor, e.currency
		FROM event e
		JOIN tenant t ON t.id = e.tenant_id
		JOIN event_type et ON et.id = e.event_type_id
		WHERE e.guest_id = $1
		ORDER BY e.occurred_at DESC
		LIMIT $2`, guestID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.GuestActionLogEntry, 0)
	for rows.Next() {
		var entry domain.GuestActionLogEntry
		var amount sql.NullInt64
		if err := rows.Scan(&entry.OccurredAt, &entry.TenantName, &entry.EventType, &amount, &entry.Currency); err != nil {
			return nil, err
		}
		if amount.Valid {
			v := amount.Int64
			entry.AmountMinor = &v
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

func (r *EventSQL) ListMovement(ctx context.Context, guestID string, from, to *time.Time, limit int) ([]domain.MovementStep, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, tenant_id, tenant_name, mall_id, mall_name, floor, occurred_at FROM (
			SELECT e.id, et.code, t.id AS tenant_id, t.name AS tenant_name, m.id AS mall_id, m.name AS mall_name,
				t.floor, e.occurred_at
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			JOIN tenant t ON t.id = e.tenant_id
			JOIN mall m ON m.id = t.mall_id AND m.is_active
			WHERE e.guest_id = $1
				AND ($2::timestamp IS NULL OR e.occurred_at >= $2)
				AND ($3::timestamp IS NULL OR e.occurred_at < $3)
			ORDER BY e.occurred_at DESC, e.id DESC
			LIMIT $4
		) recent
		ORDER BY occurred_at, id`, guestID, nullTime(from), nullTime(to), limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.MovementStep, 0)
	for rows.Next() {
		var step domain.MovementStep
		if err := rows.Scan(
			&step.EventID, &step.EventTypeCode, &step.TenantID, &step.TenantName,
			&step.MallID, &step.MallName, &step.Floor, &step.OccurredAt,
		); err != nil {
			return nil, err
		}
		out = append(out, step)
	}
	return out, rows.Err()
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func (r *EventSQL) Insert(ctx context.Context, e *domain.Event) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO event (
			id, guest_id, origin_guest_id, tenant_id, project_id, event_type_id, data_source_id,
			import_batch_id, external_id, occurred_at, amount_minor, currency, match_keys, raw_payload, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		e.ID, nullIfEmpty(e.GuestID), nullIfEmpty(e.OriginGuestID), e.TenantID, nullIfEmpty(e.ProjectID),
		e.EventTypeID, e.DataSourceID, nullIfEmpty(e.ImportBatchID), e.ExternalID, e.OccurredAt,
		e.AmountMinor, e.Currency, nullIfEmpty(e.MatchKeys), nullIfEmpty(e.RawPayload), e.CreatedAt,
	)
	return err
}

func (r *EventSQL) InsertIgnoreDuplicate(ctx context.Context, e *domain.Event) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO event (
			id, guest_id, origin_guest_id, tenant_id, project_id, event_type_id, data_source_id,
			import_batch_id, external_id, occurred_at, amount_minor, currency, match_keys, raw_payload, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (data_source_id, event_type_id, external_id) DO NOTHING`,
		e.ID, nullIfEmpty(e.GuestID), nullIfEmpty(e.OriginGuestID), e.TenantID, nullIfEmpty(e.ProjectID),
		e.EventTypeID, e.DataSourceID, nullIfEmpty(e.ImportBatchID), e.ExternalID, e.OccurredAt,
		e.AmountMinor, e.Currency, nullIfEmpty(e.MatchKeys), nullIfEmpty(e.RawPayload), e.CreatedAt,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *EventSQL) ListUnmatchedByBatch(ctx context.Context, batchID string, limit int) ([]domain.Event, error) {
	if limit <= 0 || limit > 50000 {
		limit = 5000
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(guest_id, ''), COALESCE(origin_guest_id, ''), tenant_id, COALESCE(project_id, ''),
			event_type_id, data_source_id, COALESCE(import_batch_id, ''), external_id, occurred_at,
			amount_minor, currency, COALESCE(match_keys, ''), COALESCE(raw_payload, ''), created_at
		FROM event
		WHERE import_batch_id = $1 AND guest_id IS NULL
		ORDER BY id
		LIMIT $2`, batchID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.Event, 0)
	for rows.Next() {
		var (
			e      domain.Event
			amount sql.NullInt64
		)
		if err := rows.Scan(
			&e.ID, &e.GuestID, &e.OriginGuestID, &e.TenantID, &e.ProjectID,
			&e.EventTypeID, &e.DataSourceID, &e.ImportBatchID, &e.ExternalID, &e.OccurredAt,
			&amount, &e.Currency, &e.MatchKeys, &e.RawPayload, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		if amount.Valid {
			v := amount.Int64
			e.AmountMinor = &v
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *EventSQL) AttachGuest(ctx context.Context, eventID, guestID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE event SET guest_id = $2 WHERE id = $1 AND guest_id IS NULL`, eventID, guestID)
	return err
}

func (r *EventSQL) MoveToGuest(ctx context.Context, fromGuestID, toGuestID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE event SET guest_id = $2
		WHERE guest_id = $1 AND origin_guest_id = $1`, fromGuestID, toGuestID)
	return err
}

func (r *EventSQL) RestoreOrigin(ctx context.Context, originGuestID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE event SET guest_id = origin_guest_id
		WHERE origin_guest_id = $1 AND guest_id <> origin_guest_id`, originGuestID)
	return err
}
