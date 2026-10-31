package repository

import (
	"context"
	"database/sql"

	"github.com/MangriMen/UCRP/internal/domain"
)

type EventSQL struct {
	db *sql.DB
}

func NewEventSQL(db *sql.DB) *EventSQL {
	return &EventSQL{db: db}
}

func (r *EventSQL) ListByGuest(ctx context.Context, guestID string, limit, offset int) ([]domain.Event, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(guest_id, ''), COALESCE(origin_guest_id, ''), tenant_id, COALESCE(project_id, ''),
			event_type_id, data_source_id, COALESCE(import_batch_id, ''), external_id, occurred_at,
			amount_minor, currency, COALESCE(match_keys, ''), COALESCE(raw_payload, ''), created_at
		FROM event
		WHERE guest_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2 OFFSET $3`, guestID, limit, offset)
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
