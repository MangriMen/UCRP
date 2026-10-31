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
	defer rows.Close()
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
