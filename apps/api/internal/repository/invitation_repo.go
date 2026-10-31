package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type InvitationSQL struct{ db *sql.DB }

func NewInvitationSQL(db *sql.DB) *InvitationSQL { return &InvitationSQL{db: db} }

func (r *InvitationSQL) CreateInvitation(ctx context.Context, invite domain.UserInvitation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var targetVersion any
	if invite.Purpose == "recovery" { targetVersion = invite.TargetAuthVersion }
	result, err := tx.ExecContext(ctx, `
		INSERT INTO user_invitation (id, email, display_name, role, tenant_id, token_hash, created_by, created_at, expires_at, policy_version, purpose, target_user_id, target_auth_version)
		SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13 FROM access_policy_state
		WHERE id = 1 AND version = $10 AND (
			($11 = 'invite' AND NOT EXISTS (SELECT 1 FROM app_user WHERE email = $2)) OR
			($11 = 'recovery' AND EXISTS (SELECT 1 FROM app_user WHERE id = $12 AND email = $2 AND role = $4 AND tenant_id IS NOT DISTINCT FROM $5 AND auth_version = $13 AND is_active)))`,
		invite.ID, invite.Email, invite.DisplayName, invite.Role, nullIfEmpty(invite.TenantID), invite.TokenHash,
		invite.CreatedBy, invite.CreatedAt, invite.ExpiresAt, invite.PolicyVersion, invite.Purpose, nullIfEmpty(invite.TargetUserID), targetVersion)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return domain.ErrConflict
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, created_at)
		VALUES ($1,$2,'manager','user.invitation.create','user_invitation',$3,$4)`, invite.AuditID, invite.CreatedBy, invite.ID, invite.CreatedAt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *InvitationSQL) RevokeInvitation(ctx context.Context, id string, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE user_invitation SET revoked_at = $2 WHERE id = $1 AND consumed_at IS NULL`, id, now)
	return err
}

func (r *InvitationSQL) ValidInvitation(ctx context.Context, hash []byte, now time.Time) (bool, error) {
	var valid bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM user_invitation i JOIN access_policy_state s ON s.id = 1 AND s.version = i.policy_version
		JOIN app_user u ON u.id = i.created_by AND u.role = 'manager' AND u.is_active
		WHERE i.token_hash = $1 AND i.consumed_at IS NULL AND i.revoked_at IS NULL AND i.expires_at > $2)`, hash, now).Scan(&valid)
	return valid, err
}

func (r *InvitationSQL) AcceptInvitation(ctx context.Context, hash []byte, userID, passwordHash, auditID string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var version int64
	if err := tx.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1 FOR SHARE`).Scan(&version); err != nil {
		return err
	}
	var creator string
	err = tx.QueryRowContext(ctx, `
		SELECT u.id FROM user_invitation i JOIN app_user u ON u.id = i.created_by
		WHERE i.token_hash = $1 AND u.role = 'manager' AND u.is_active FOR UPDATE OF u`, hash).Scan(&creator)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrUnauthorized
	}
	if err != nil {
		return err
	}
	var invitationID, email, name, role, purpose string
	var tenantID, targetUserID sql.NullString
	var targetVersion sql.NullInt64
	err = tx.QueryRowContext(ctx, `
		UPDATE user_invitation SET consumed_at = $2
		WHERE token_hash = $1 AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at > $2 AND policy_version = $3
		AND EXISTS (SELECT 1 FROM app_user WHERE id = created_by AND role = 'manager' AND is_active)
		RETURNING id, email, display_name, role, tenant_id, purpose, target_user_id, target_auth_version`, hash, now, version).Scan(&invitationID, &email, &name, &role, &tenantID, &purpose, &targetUserID, &targetVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrUnauthorized
	}
	if err != nil {
		return err
	}
	if purpose == "invite" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO app_user (id, email, password_hash, display_name, role, tenant_id, is_active, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,TRUE,$7)`, userID, email, passwordHash, name, role, tenantID, now)
		if err != nil { return err }
	} else if purpose == "recovery" && targetUserID.Valid && targetVersion.Valid {
		userID = targetUserID.String
		result, updateErr := tx.ExecContext(ctx, `
			UPDATE app_user SET password_hash = $2, auth_version = auth_version + 1
			WHERE id = $1 AND email = $3 AND role = $4 AND tenant_id IS NOT DISTINCT FROM $5 AND auth_version = $6 AND is_active`,
			userID, passwordHash, email, role, tenantID, targetVersion.Int64)
		if updateErr != nil { return updateErr }
		count, countErr := result.RowsAffected()
		if countErr != nil { return countErr }
		if count != 1 { return domain.ErrUnauthorized }
	} else {
		return domain.ErrUnauthorized
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, created_at)
		VALUES ($1,$2,$3,'user.invitation.accept','user_invitation',$4,$5)`, auditID, userID, role, invitationID, now)
	if err != nil {
		return err
	}
	return tx.Commit()
}
