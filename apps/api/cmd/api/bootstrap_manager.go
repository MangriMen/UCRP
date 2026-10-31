package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
)

type bootstrapManagerCredentials struct {
	email    string
	password string
}

func readBootstrapManagerCredentials(getenv func(string) string) (bootstrapManagerCredentials, bool, error) {
	rawEmail := getenv("UCRP_BOOTSTRAP_MANAGER_EMAIL")
	email := strings.TrimSpace(rawEmail)
	password := getenv("UCRP_BOOTSTRAP_MANAGER_PASSWORD")
	if rawEmail == "" && password == "" {
		return bootstrapManagerCredentials{}, false, nil
	}
	if email == "" || password == "" {
		return bootstrapManagerCredentials{}, false, errors.New("UCRP_BOOTSTRAP_MANAGER_EMAIL and UCRP_BOOTSTRAP_MANAGER_PASSWORD must be provided together")
	}
	return bootstrapManagerCredentials{email: email, password: password}, true, nil
}

func bootstrapManager(ctx context.Context, db *sql.DB, clock domain.Clock, ids domain.IDGen) error {
	credentials, enabled, err := readBootstrapManagerCredentials(os.Getenv)
	if err != nil || !enabled {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("bootstrap manager transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var version int64
	if err := tx.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1 FOR UPDATE`).Scan(&version); err != nil {
		return fmt.Errorf("bootstrap manager policy lock: %w", err)
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM app_user WHERE role = 'manager')`).Scan(&exists); err != nil {
		return fmt.Errorf("bootstrap manager lookup: %w", err)
	}
	if exists {
		return tx.Commit()
	}
	hash, err := auth.HashPassword(credentials.password)
	if err != nil {
		return fmt.Errorf("bootstrap manager password: %w", err)
	}
	now := clock.Now().UTC()
	userID, grantID, auditID := ids.NewID(ctx), ids.NewID(ctx), ids.NewID(ctx)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO app_user (id, email, password_hash, display_name, role, is_active, created_at)
		VALUES ($1, $2, $3, 'Manager', 'manager', TRUE, $4)`, userID, credentials.email, hash, now); err != nil {
		return fmt.Errorf("bootstrap manager account: %w", err)
	}
	version++
	if _, err := tx.ExecContext(ctx, `
		UPDATE access_policy_state SET version = $1, updated_at = $2 WHERE id = 1`, version, now); err != nil {
		return fmt.Errorf("bootstrap manager policy version: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO access_grant (id, subject_type, subject_id, status, valid_from, reason, policy_version, created_at, updated_at)
		VALUES ($1, 'user', $2, 'active', $3, 'Initial manager bootstrap', $4, $3, $3)`, grantID, userID, now, version); err != nil {
		return fmt.Errorf("bootstrap manager grant: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO access_grant_item (grant_id, permission_code, effect)
		SELECT $1, rp.permission_code, rp.effect
		FROM role_permission rp JOIN permission p ON p.code = rp.permission_code AND p.is_active
		WHERE rp.role = 'manager' AND rp.effect = 'allow'
		AND (rp.permission_code LIKE 'field.guest.%' OR rp.permission_code LIKE 'field.event.%'
		  OR rp.permission_code IN ('guest.profile.read', 'guest.profile.search', 'guest.event.read',
		                            'guest.identifier.read', 'guest.contact_history.read'))`, grantID)
	if err != nil {
		return fmt.Errorf("bootstrap manager permissions: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("bootstrap manager permission count: %w", err)
	}
	if count != 16 {
		return errors.New("bootstrap manager permissions do not match the bootstrap catalog")
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO access_grant_scope (grant_id, scope_type, scope_value) VALUES ($1, 'principal_role', 'manager')`, grantID); err != nil {
		return fmt.Errorf("bootstrap manager grant scope: %w", err)
	}
	details, err := json.Marshal(map[string]any{"origin": "bootstrap_env", "subjectId": userID, "policyVersion": version})
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO audit_log (id, actor_role, action, entity_type, entity_id, details, created_at)
		VALUES ($1, 'system', 'access.grant.create', 'access_grant', $2, $3, $4)`, auditID, grantID, string(details), now); err != nil {
		return fmt.Errorf("bootstrap manager audit: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("bootstrap manager commit: %w", err)
	}
	return nil
}
