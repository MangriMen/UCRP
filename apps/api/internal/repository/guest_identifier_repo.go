package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/MangriMen/UCRP/internal/domain"
)

type GuestIdentifierSQL struct {
	db *sql.DB
}

func NewGuestIdentifierSQL(db *sql.DB) *GuestIdentifierSQL {
	return &GuestIdentifierSQL{db: db}
}

func (r *GuestIdentifierSQL) ListByGuest(ctx context.Context, guestID string) ([]domain.GuestIdentifier, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, guest_id, origin_guest_id, id_type, value_normalized, data_source_id, first_seen_at
		FROM guest_identifier
		WHERE guest_id = $1
		ORDER BY first_seen_at`, guestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.GuestIdentifier, 0)
	for rows.Next() {
		var gi domain.GuestIdentifier
		if err := rows.Scan(&gi.ID, &gi.GuestID, &gi.OriginGuestID, &gi.IDType, &gi.ValueNormalized, &gi.DataSourceID, &gi.FirstSeenAt); err != nil {
			return nil, err
		}
		out = append(out, gi)
	}
	return out, rows.Err()
}

func (r *GuestIdentifierSQL) ListByGuests(ctx context.Context, guestIDs []string) ([]domain.GuestIdentifier, error) {
	out := make([]domain.GuestIdentifier, 0)
	if len(guestIDs) == 0 {
		return out, nil
	}
	placeholders := make([]string, len(guestIDs))
	args := make([]any, len(guestIDs))
	for i, id := range guestIDs {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, guest_id, origin_guest_id, id_type, value_normalized, data_source_id, first_seen_at
		FROM guest_identifier
		WHERE guest_id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY guest_id, first_seen_at, id`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var gi domain.GuestIdentifier
		if err := rows.Scan(&gi.ID, &gi.GuestID, &gi.OriginGuestID, &gi.IDType, &gi.ValueNormalized, &gi.DataSourceID, &gi.FirstSeenAt); err != nil {
			return nil, err
		}
		out = append(out, gi)
	}
	return out, rows.Err()
}

func (r *GuestIdentifierSQL) FindByValue(ctx context.Context, idType, value string) ([]domain.GuestIdentifier, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, guest_id, origin_guest_id, id_type, value_normalized, data_source_id, first_seen_at
		FROM guest_identifier
		WHERE id_type = $1 AND value_normalized = $2`, idType, value)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.GuestIdentifier, 0)
	for rows.Next() {
		var gi domain.GuestIdentifier
		if err := rows.Scan(&gi.ID, &gi.GuestID, &gi.OriginGuestID, &gi.IDType, &gi.ValueNormalized, &gi.DataSourceID, &gi.FirstSeenAt); err != nil {
			return nil, err
		}
		out = append(out, gi)
	}
	return out, rows.Err()
}

func (r *GuestIdentifierSQL) Insert(ctx context.Context, gi *domain.GuestIdentifier) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO guest_identifier (id, guest_id, origin_guest_id, id_type, value_normalized, data_source_id, first_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		gi.ID, gi.GuestID, gi.OriginGuestID, gi.IDType, gi.ValueNormalized, gi.DataSourceID, gi.FirstSeenAt,
	)
	return err
}

func (r *GuestIdentifierSQL) MoveToGuest(ctx context.Context, fromGuestID, toGuestID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE guest_identifier SET guest_id = $2
		WHERE guest_id = $1 AND origin_guest_id = $1`, fromGuestID, toGuestID)
	return err
}

func (r *GuestIdentifierSQL) RestoreOrigin(ctx context.Context, originGuestID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE guest_identifier SET guest_id = origin_guest_id
		WHERE origin_guest_id = $1 AND guest_id <> origin_guest_id`, originGuestID)
	return err
}
