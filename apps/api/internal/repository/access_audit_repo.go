package repository

import (
	"context"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

const accessAuditWhere = `
	WHERE (
		(entity_type = 'access_grant' AND (action LIKE 'access.grant.%' OR ($1 = '' AND action = 'access.simulate')))
		OR ($1 = '' AND entity_type = 'access' AND action = 'access.denied')
	)
	AND ($1 = '' OR entity_id = $1)
	AND ($2 = '' OR actor_user_id = $2)
	AND ($3::timestamp IS NULL OR created_at >= $3::timestamp)
	AND ($4::timestamp IS NULL OR created_at <= $4::timestamp)`

func (r *AuditSQL) ListAccessAudit(ctx context.Context, grantID, actorID string, from, to *time.Time, limit, offset int) ([]domain.AuditLog, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log`+accessAuditWhere,
		grantID, actorID, from, to).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(actor_user_id, ''), actor_role, action, entity_type, entity_id,
			COALESCE(details, ''), COALESCE(request_id, ''), created_at
		FROM audit_log`+accessAuditWhere+`
		ORDER BY created_at DESC, id DESC LIMIT $5 OFFSET $6`, grantID, actorID, from, to, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	entries := make([]domain.AuditLog, 0)
	for rows.Next() {
		var entry domain.AuditLog
		if err := rows.Scan(&entry.ID, &entry.ActorUserID, &entry.ActorRole, &entry.Action, &entry.EntityType, &entry.EntityID, &entry.Details, &entry.RequestID, &entry.CreatedAt); err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}
