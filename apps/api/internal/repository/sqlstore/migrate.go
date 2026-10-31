package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

func Migrate(ctx context.Context, db *sql.DB, fsys fs.FS, dir string) error {
	goose.SetBaseFS(fsys)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.UpContext(ctx, db, dir); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
