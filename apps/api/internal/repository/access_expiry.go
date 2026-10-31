package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/MangriMen/UCRP/internal/domain"
)

type expiredGrant struct {
	id          string
	status      string
	subjectType string
	subjectID   string
	validTo     time.Time
}

func (r *AccessSQL) ExpireDue(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 || limit > 1000 {
		return 0, domain.ErrValidation
	}
	now = now.UTC()
	var due bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM access_grant
		WHERE status IN ('draft', 'active', 'suspended') AND valid_to IS NOT NULL AND valid_to <= $1)`, now).Scan(&due); err != nil {
		return 0, err
	}
	if !due {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var version int64
	if err := tx.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1 FOR UPDATE`).Scan(&version); err != nil {
		return 0, err
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT id, status, subject_type, subject_id, valid_to FROM access_grant
		WHERE status IN ('draft', 'active', 'suspended') AND valid_to IS NOT NULL AND valid_to <= $1
		ORDER BY valid_to, id LIMIT $2 FOR UPDATE SKIP LOCKED`, now, limit)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	grants := make([]expiredGrant, 0, limit)
	for rows.Next() {
		var grant expiredGrant
		if err := rows.Scan(&grant.id, &grant.status, &grant.subjectType, &grant.subjectID, &grant.validTo); err != nil {
			return 0, err
		}
		grants = append(grants, grant)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	for _, grant := range grants {
		version++
		if _, err := tx.ExecContext(ctx, `
			UPDATE access_grant SET status = 'expired', policy_version = $2, updated_at = $3 WHERE id = $1`,
			grant.id, version, now); err != nil {
			return 0, err
		}
		details, err := json.Marshal(map[string]any{
			"previousStatus": grant.status, "subjectType": grant.subjectType, "subjectId": grant.subjectID,
			"validTo": grant.validTo, "policyVersion": version,
		})
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, created_at)
			VALUES ($1, NULL, 'system', 'access.grant.expire', 'access_grant', $2, $3, $4)`,
			uuid.NewString(), grant.id, string(details), now); err != nil {
			return 0, err
		}
	}
	if len(grants) > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE access_policy_state SET version = $1, updated_at = $2 WHERE id = 1`, version, now); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(grants), nil
}
