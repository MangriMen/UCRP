package fixtures

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
}

func OpenDB(ctx context.Context, databaseURL string) (*DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("UCRP_E2E_DATABASE_URL is required for fixtures")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(5)
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &DB{DB: db}, nil
}

func (d *DB) Exec(ctx context.Context, query string, args ...any) error {
	if _, err := d.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("exec %q: %w", firstLine(query), err)
	}
	return nil
}

func (d *DB) Count(ctx context.Context, query string, args ...any) (int, error) {
	var n int
	if err := d.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("count %q: %w", firstLine(query), err)
	}
	return n, nil
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}
