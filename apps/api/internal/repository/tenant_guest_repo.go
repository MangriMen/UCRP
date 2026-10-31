package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type TenantGuestSQL struct {
	db *sql.DB
}

func NewTenantGuestSQL(db *sql.DB) *TenantGuestSQL {
	return &TenantGuestSQL{db: db}
}

func (r *TenantGuestSQL) Get(ctx context.Context, tenantID, guestID string, since time.Time) (*domain.TenantGuest, error) {
	var guest domain.TenantGuest
	err := r.db.QueryRowContext(ctx, `
		SELECT g.id, MIN(e.occurred_at) FROM guest g
		JOIN event e ON e.guest_id = g.id AND e.tenant_id = $2 AND e.occurred_at >= $3
		WHERE g.id = $1 AND g.status = 'active' GROUP BY g.id`,
		guestID, tenantID, since).Scan(&guest.ID, &guest.FirstSeenAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &guest, nil
}

func (r *TenantGuestSQL) Identifiers(ctx context.Context, tenantID, guestID string, since time.Time, limit int) ([]domain.TenantGuestIdentifier, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT i.id, i.id_type, i.value_normalized, i.first_seen_at
		FROM guest_identifier i JOIN data_source d ON d.id = i.data_source_id
		WHERE i.guest_id = $1 AND d.tenant_id = $2 AND i.first_seen_at >= $3
		ORDER BY i.first_seen_at DESC, i.id LIMIT $4`, guestID, tenantID, since, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	identifiers := make([]domain.TenantGuestIdentifier, 0)
	for rows.Next() {
		var identifier domain.TenantGuestIdentifier
		if err := rows.Scan(&identifier.ID, &identifier.Type, &identifier.ValueNormalized, &identifier.FirstSeenAt); err != nil {
			return nil, err
		}
		identifiers = append(identifiers, identifier)
	}
	return identifiers, rows.Err()
}

func (r *TenantGuestSQL) Events(ctx context.Context, tenantID, guestID string, since time.Time, limit, offset int) ([]domain.TenantGuestEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, e.event_type_id, e.tenant_id, COALESCE(p.id, ''), e.external_id, e.occurred_at, e.amount_minor, e.currency
		FROM event e LEFT JOIN project p ON p.id = e.project_id AND p.tenant_id = e.tenant_id
		WHERE e.guest_id = $1 AND e.tenant_id = $2 AND e.occurred_at >= $3
		ORDER BY e.occurred_at DESC, e.id LIMIT $4 OFFSET $5`, guestID, tenantID, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	events := make([]domain.TenantGuestEvent, 0)
	for rows.Next() {
		var event domain.TenantGuestEvent
		var amount sql.NullInt64
		if err := rows.Scan(&event.ID, &event.EventTypeID, &event.TenantID, &event.ProjectID, &event.ExternalID, &event.OccurredAt, &amount, &event.Currency); err != nil {
			return nil, err
		}
		if amount.Valid {
			value := amount.Int64
			event.AmountMinor = &value
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
