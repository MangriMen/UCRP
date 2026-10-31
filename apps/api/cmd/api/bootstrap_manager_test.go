package main

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
	"github.com/MangriMen/UCRP/internal/repository/sqlstore"
	"github.com/MangriMen/UCRP/migrations"
)

func TestBootstrapManagerCredentials(t *testing.T) {
	for _, test := range []struct {
		name, email, password string
		enabled, failure      bool
	}{
		{"absent", "", "", false, false},
		{"email only", "manager@example.test", "", false, true},
		{"password only", "", "test-password", false, true},
		{"blank email only", "  ", "", false, true},
		{"whitespace email", "  ", "test-password", false, true},
		{"both", " manager@example.test ", "test-password", true, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			values := map[string]string{"UCRP_BOOTSTRAP_MANAGER_EMAIL": test.email, "UCRP_BOOTSTRAP_MANAGER_PASSWORD": test.password}
			credentials, enabled, err := readBootstrapManagerCredentials(func(key string) string { return values[key] })
			if enabled != test.enabled || (err != nil) != test.failure {
				t.Fatalf("enabled=%t error=%v", enabled, err)
			}
			if err != nil && strings.Contains(err.Error(), test.password) && test.password != "" {
				t.Fatal("error exposes the password")
			}
			if enabled && (credentials.email != strings.TrimSpace(test.email) || credentials.password != test.password) {
				t.Fatal("credentials changed unexpectedly")
			}
		})
	}
}

func TestBootstrapManagerPostgres(t *testing.T) {
	databaseURL := os.Getenv("UCRP_BOOTSTRAP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("UCRP_BOOTSTRAP_TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}
	adminDB := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = adminDB.Close() })
	schema := "ucrp_bootstrap_test_" + strings.ReplaceAll(idgen.UUID{}.NewID(ctx), "-", "")
	if _, err := adminDB.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if _, err := adminDB.ExecContext(cleanupCtx, `DROP SCHEMA "`+schema+`" CASCADE`); err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
	})
	config.RuntimeParams["search_path"] = schema
	db := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = db.Close() })
	if err := sqlstore.Migrate(ctx, db, migrations.FS, "common"); err != nil {
		t.Fatalf("migrate isolated schema: %v", err)
	}
	t.Setenv("UCRP_BOOTSTRAP_MANAGER_EMAIL", "bootstrap@example.test")
	t.Setenv("UCRP_BOOTSTRAP_MANAGER_PASSWORD", "bootstrap-test-password")
	const workers = 6
	errors := make(chan error, workers)
	var wait sync.WaitGroup
	for i := 0; i < workers; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			errors <- bootstrapManager(ctx, db, domain.SystemClock{}, idgen.UUID{})
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent bootstrap: %v", err)
		}
	}
	var managers, grants, items, audits, version int
	var hash string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*), MIN(password_hash) FROM app_user WHERE role = 'manager'`).Scan(&managers, &hash); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM access_grant WHERE subject_type = 'user' AND status = 'active'`).Scan(&grants); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM access_grant_item i JOIN access_grant g ON g.id = i.grant_id WHERE g.subject_type = 'user'`).Scan(&items); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log WHERE action = 'access.grant.create' AND actor_role = 'system' AND details LIKE '%bootstrap_env%'`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if managers != 1 || grants != 1 || items != 16 || audits != 1 || version != 2 || !auth.CheckPassword(hash, "bootstrap-test-password") {
		t.Fatalf("unexpected bootstrap state: managers=%d grants=%d items=%d audits=%d version=%d", managers, grants, items, audits, version)
	}
	var scoped bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM access_grant_scope s JOIN access_grant g ON g.id = s.grant_id
		WHERE g.subject_type = 'user' AND s.scope_type = 'principal_role' AND s.scope_value = 'manager')`).Scan(&scoped); err != nil || !scoped {
		t.Fatalf("missing manager role scope: %v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE app_user SET is_active = FALSE WHERE role = 'manager'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE access_grant SET status = 'revoked' WHERE subject_type = 'user'`); err != nil {
		t.Fatal(err)
	}
	if err := bootstrapManager(ctx, db, domain.SystemClock{}, idgen.UUID{}); err != nil {
		t.Fatalf("repeat bootstrap: %v", err)
	}
	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM access_grant WHERE subject_type = 'user'`).Scan(&status); err != nil || status != "revoked" {
		t.Fatalf("bootstrap restored revoked grant: status=%q err=%v", status, err)
	}
	var nextVersion int
	if err := db.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1`).Scan(&nextVersion); err != nil || nextVersion != version {
		t.Fatalf("repeat bootstrap changed policy version: %d, %v", nextVersion, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE app_user SET role = 'analyst' WHERE role = 'manager'`); err != nil {
		t.Fatal(err)
	}
	if err := bootstrapManager(ctx, db, domain.SystemClock{}, idgen.UUID{}); err == nil {
		t.Fatal("bootstrap accepted an email already owned by another role")
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM app_user WHERE role = 'manager'`).Scan(&managers); err != nil || managers != 0 {
		t.Fatalf("failed bootstrap left manager account: %d, %v", managers, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1`).Scan(&nextVersion); err != nil || nextVersion != version {
		t.Fatalf("failed bootstrap changed policy version: %d, %v", nextVersion, err)
	}
}
