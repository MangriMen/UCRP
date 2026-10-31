package repository

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestAccessExpiryPostgres(t *testing.T) {
	databaseURL := os.Getenv("UCRP_ACCESS_EXPIRY_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("UCRP_ACCESS_EXPIRY_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	admin, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	schema := "ucrp_expiry_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.ExecContext(context.Background(), `DROP SCHEMA `+schema+` CASCADE`); err != nil {
			t.Errorf("drop schema: %v", err)
		}
	})
	parsed, err := url.Parse(databaseURL)
	if err != nil || parsed.Scheme == "" {
		t.Fatalf("test database URL: %v", err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := sql.Open("pgx", parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	statements := []string{
		`CREATE TABLE access_policy_state (id INTEGER PRIMARY KEY, version BIGINT NOT NULL, updated_at TIMESTAMP NOT NULL)`,
		`INSERT INTO access_policy_state VALUES (1, 7, CURRENT_TIMESTAMP)`,
		`CREATE TABLE access_grant (id TEXT PRIMARY KEY, status TEXT NOT NULL, subject_type TEXT NOT NULL, subject_id TEXT NOT NULL, valid_to TIMESTAMP, policy_version BIGINT NOT NULL, updated_at TIMESTAMP NOT NULL)`,
		`CREATE TABLE audit_log (id TEXT PRIMARY KEY, actor_user_id TEXT, actor_role TEXT NOT NULL, action TEXT NOT NULL, entity_type TEXT NOT NULL, entity_id TEXT NOT NULL, details TEXT, created_at TIMESTAMP NOT NULL)`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	for _, grant := range []struct {
		id, status string
		expiry     *time.Time
	}{
		{"active", "active", expiryTime(now.Add(-time.Second))},
		{"suspended", "suspended", expiryTime(now)},
		{"draft", "draft", expiryTime(now.Add(-time.Minute))},
		{"future", "active", expiryTime(now.Add(time.Second))},
		{"indefinite", "active", nil},
		{"revoked", "revoked", expiryTime(now.Add(-time.Second))},
	} {
		_, err := db.ExecContext(ctx, `INSERT INTO access_grant VALUES ($1, $2, 'tenant', 'tenant-1', $3, 7, $4)`, grant.id, grant.status, grant.expiry, now)
		if err != nil {
			t.Fatal(err)
		}
	}
	repo := NewAccessSQL(db)
	var wg sync.WaitGroup
	counts := make(chan int, 2)
	errorsFound := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count, err := repo.ExpireDue(ctx, now, 100)
			counts <- count
			errorsFound <- err
		}()
	}
	wg.Wait()
	close(counts)
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatal(err)
		}
	}
	total := 0
	for count := range counts {
		total += count
	}
	if total != 3 {
		t.Fatalf("expired count = %d, want 3", total)
	}
	var version, distinctVersions, audits int
	if err := db.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT policy_version) FROM access_grant WHERE status = 'expired'`).Scan(&distinctVersions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log WHERE action = 'access.grant.expire' AND actor_role = 'system'`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if version != 10 || distinctVersions != 3 || audits != 3 {
		t.Fatalf("version = %d, distinct grant versions = %d, audits = %d", version, distinctVersions, audits)
	}
	for id, want := range map[string]string{"active": "expired", "suspended": "expired", "draft": "expired", "future": "active", "indefinite": "active", "revoked": "revoked"} {
		var got string
		if err := db.QueryRowContext(ctx, `SELECT status FROM access_grant WHERE id = $1`, id).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s: status = %s, want %s", id, got, want)
		}
	}
	if count, err := repo.ExpireDue(ctx, now, 100); err != nil || count != 0 {
		t.Fatalf("repeat: count = %d, error = %v", count, err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE audit_log ADD CONSTRAINT reject_expiry CHECK (action <> 'access.grant.expire') NOT VALID`); err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO access_grant VALUES ('rollback', 'active', 'tenant', 'tenant-1', $1, 10, $2)`, now, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ExpireDue(ctx, now, 100); err == nil {
		t.Fatal("expected audit insert failure")
	}
	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM access_grant WHERE id = 'rollback'`).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if status != "active" || version != 10 {
		t.Fatalf("rollback: status = %s, version = %d", status, version)
	}
}

func expiryTime(value time.Time) *time.Time { return &value }
