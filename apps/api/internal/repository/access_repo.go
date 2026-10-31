package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type AccessSQL struct {
	db *sql.DB
}

func NewAccessSQL(db *sql.DB) *AccessSQL {
	return &AccessSQL{db: db}
}

func (r *AccessSQL) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT code, domain, kind, title, COALESCE(description, ''), is_sensitive, min_k, is_active
		FROM permission ORDER BY domain, code`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	permissions := make([]domain.Permission, 0)
	for rows.Next() {
		var permission domain.Permission
		var minK sql.NullInt64
		if err := rows.Scan(&permission.Code, &permission.Domain, &permission.Kind, &permission.Title, &permission.Description, &permission.IsSensitive, &minK, &permission.IsActive); err != nil {
			return nil, err
		}
		if minK.Valid {
			value := int(minK.Int64)
			permission.MinK = &value
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *AccessSQL) RolePermissions(ctx context.Context, role domain.Role) ([]domain.AccessRolePermission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT rp.permission_code, rp.effect FROM role_permission rp
		JOIN permission p ON p.code = rp.permission_code AND p.is_active
		WHERE rp.role = $1 ORDER BY rp.permission_code`, role)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	permissions := make([]domain.AccessRolePermission, 0)
	for rows.Next() {
		var permission domain.AccessRolePermission
		if err := rows.Scan(&permission.PermissionCode, &permission.Effect); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *AccessSQL) PolicyVersion(ctx context.Context) (int64, error) {
	var version int64
	err := r.db.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1`).Scan(&version)
	return version, err
}

const accessGrantColumns = `id, subject_type, subject_id, COALESCE(preset_code, ''), status,
	valid_from, valid_to, COALESCE(created_by, ''), COALESCE(approved_by, ''), COALESCE(reason, ''),
	policy_version, created_at, updated_at`

func (r *AccessSQL) ActiveGrants(ctx context.Context, principal domain.Principal, now time.Time) ([]domain.AccessGrant, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+accessGrantColumns+`
		FROM access_grant
		WHERE status = 'active' AND valid_from <= $1 AND (valid_to IS NULL OR valid_to > $1)
		AND ((subject_type = 'user' AND subject_id = $2)
		  OR (subject_type = 'tenant' AND subject_id = $3 AND $3 <> '')
		  OR (subject_type = 'role' AND subject_id = $4))
		ORDER BY id`, now, principal.UserID, principal.TenantID, principal.Role)
	if err != nil {
		return nil, err
	}
	return r.collectGrants(ctx, rows)
}

func (r *AccessSQL) GetGrant(ctx context.Context, id string) (*domain.AccessGrant, error) {
	grant, err := scanAccessGrant(r.db.QueryRowContext(ctx, `SELECT `+accessGrantColumns+` FROM access_grant WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := r.loadGrantDetails(ctx, grant); err != nil {
		return nil, err
	}
	return grant, nil
}

func (r *AccessSQL) ListGrants(ctx context.Context, filter domain.AccessGrantFilter) ([]domain.AccessGrant, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+accessGrantColumns+`
		FROM access_grant
		WHERE ($1 = '' OR subject_type = $1) AND ($2 = '' OR subject_id = $2) AND ($3 = '' OR status = $3)
		ORDER BY created_at DESC, id LIMIT $4 OFFSET $5`, filter.SubjectType, filter.SubjectID, filter.Status, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	return r.collectGrants(ctx, rows)
}

func (r *AccessSQL) collectGrants(ctx context.Context, rows *sql.Rows) ([]domain.AccessGrant, error) {
	defer func() { _ = rows.Close() }()
	grants := make([]domain.AccessGrant, 0)
	for rows.Next() {
		grant, err := scanAccessGrant(rows)
		if err != nil {
			return nil, err
		}
		grants = append(grants, *grant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range grants {
		if err := r.loadGrantDetails(ctx, &grants[i]); err != nil {
			return nil, err
		}
	}
	return grants, nil
}

func scanAccessGrant(s scanner) (*domain.AccessGrant, error) {
	var grant domain.AccessGrant
	var validTo sql.NullTime
	err := s.Scan(&grant.ID, &grant.SubjectType, &grant.SubjectID, &grant.PresetCode, &grant.Status,
		&grant.ValidFrom, &validTo, &grant.CreatedBy, &grant.ApprovedBy, &grant.Reason,
		&grant.PolicyVersion, &grant.CreatedAt, &grant.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if validTo.Valid {
		grant.ValidTo = &validTo.Time
	}
	return &grant, nil
}

func (r *AccessSQL) loadGrantDetails(ctx context.Context, grant *domain.AccessGrant) error {
	items, err := r.db.QueryContext(ctx, `
		SELECT permission_code, effect FROM access_grant_item
		WHERE grant_id = $1 ORDER BY permission_code`, grant.ID)
	if err != nil {
		return err
	}
	defer func() { _ = items.Close() }()
	grant.Items = make([]domain.AccessGrantItem, 0)
	for items.Next() {
		var item domain.AccessGrantItem
		if err := items.Scan(&item.PermissionCode, &item.Effect); err != nil {
			return err
		}
		grant.Items = append(grant.Items, item)
	}
	if err := items.Err(); err != nil {
		return err
	}
	if err := items.Close(); err != nil {
		return err
	}
	scopes, err := r.db.QueryContext(ctx, `
		SELECT scope_type, scope_value FROM access_grant_scope WHERE grant_id = $1 ORDER BY scope_type, scope_value`, grant.ID)
	if err != nil {
		return err
	}
	defer func() { _ = scopes.Close() }()
	grant.Scopes = make([]domain.AccessGrantScope, 0)
	for scopes.Next() {
		var scope domain.AccessGrantScope
		if err := scopes.Scan(&scope.Type, &scope.Value); err != nil {
			return err
		}
		grant.Scopes = append(grant.Scopes, scope)
	}
	return scopes.Err()
}

func (r *AccessSQL) CreateGrant(ctx context.Context, grant *domain.AccessGrant, expectedVersion int64, audit *domain.AuditLog) error {
	return r.saveGrant(ctx, grant, expectedVersion, audit, true)
}

func (r *AccessSQL) UpdateGrant(ctx context.Context, grant *domain.AccessGrant, expectedVersion int64, audit *domain.AuditLog) error {
	return r.saveGrant(ctx, grant, expectedVersion, audit, false)
}

func (r *AccessSQL) saveGrant(ctx context.Context, grant *domain.AccessGrant, expectedVersion int64, audit *domain.AuditLog, create bool) error {
	if grant == nil || audit == nil {
		return domain.ErrValidation
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var version int64
	err = tx.QueryRowContext(ctx, `
		UPDATE access_policy_state SET version = version + 1, updated_at = $1
		WHERE id = 1 AND version = $2 RETURNING version`, grant.UpdatedAt, expectedVersion).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrConflict
	}
	if err != nil {
		return err
	}
	if err := validateAccessSubject(ctx, tx, grant.SubjectType, grant.SubjectID); err != nil {
		return err
	}
	if create {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO access_grant (id, subject_type, subject_id, preset_code, status, valid_from, valid_to,
				created_by, approved_by, reason, policy_version, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			grant.ID, grant.SubjectType, grant.SubjectID, nullIfEmpty(grant.PresetCode), grant.Status,
			grant.ValidFrom, grant.ValidTo, nullIfEmpty(grant.CreatedBy), nullIfEmpty(grant.ApprovedBy), nullIfEmpty(grant.Reason),
			version, grant.CreatedAt, grant.UpdatedAt)
	} else {
		var result sql.Result
		result, err = tx.ExecContext(ctx, `
			UPDATE access_grant SET preset_code = $2, status = $3, valid_from = $4, valid_to = $5,
				approved_by = $6, reason = $7, policy_version = $8, updated_at = $9
			WHERE id = $1 AND subject_type = $10 AND subject_id = $11`,
			grant.ID, nullIfEmpty(grant.PresetCode), grant.Status, grant.ValidFrom, grant.ValidTo,
			nullIfEmpty(grant.ApprovedBy), nullIfEmpty(grant.Reason), version, grant.UpdatedAt, grant.SubjectType, grant.SubjectID)
		if err == nil {
			var affected int64
			affected, err = result.RowsAffected()
			if err == nil && affected == 0 {
				return domain.ErrNotFound
			}
		}
	}
	if err != nil {
		return err
	}
	if !create {
		if _, err := tx.ExecContext(ctx, `DELETE FROM access_grant_item WHERE grant_id = $1`, grant.ID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM access_grant_scope WHERE grant_id = $1`, grant.ID); err != nil {
			return err
		}
	}
	for _, item := range grant.Items {
		if _, err := tx.ExecContext(ctx, `INSERT INTO access_grant_item (grant_id, permission_code, effect) VALUES ($1,$2,$3)`, grant.ID, item.PermissionCode, item.Effect); err != nil {
			return err
		}
	}
	for _, scope := range grant.Scopes {
		if _, err := tx.ExecContext(ctx, `INSERT INTO access_grant_scope (grant_id, scope_type, scope_value) VALUES ($1,$2,$3)`, grant.ID, scope.Type, scope.Value); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, request_id, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, audit.ID, nullIfEmpty(audit.ActorUserID), audit.ActorRole,
		audit.Action, audit.EntityType, audit.EntityID, nullIfEmpty(audit.Details), nullIfEmpty(audit.RequestID), audit.CreatedAt)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	grant.PolicyVersion = version
	return nil
}

func validateAccessSubject(ctx context.Context, tx *sql.Tx, subjectType, subjectID string) error {
	var exists bool
	var err error
	switch subjectType {
	case "user":
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM app_user WHERE id = $1)`, subjectID).Scan(&exists)
	case "tenant":
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM tenant WHERE id = $1)`, subjectID).Scan(&exists)
	case "role":
		switch domain.Role(subjectID) {
		case domain.RoleMarketer, domain.RoleManager, domain.RoleTenant, domain.RoleAnalyst, "system":
			exists = true
		}
	default:
		return domain.ErrValidation
	}
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}
	return nil
}
