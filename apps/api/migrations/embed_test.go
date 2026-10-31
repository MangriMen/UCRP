package migrations

import (
	"testing"

	"github.com/pressly/goose/v3"
)

func TestCommonMigrationVersionsAreUnique(t *testing.T) {
	goose.SetBaseFS(FS)
	t.Cleanup(func() { goose.SetBaseFS(nil) })
	if _, err := goose.CollectMigrations("common", 0, goose.MaxVersion); err != nil {
		t.Fatal(err)
	}
}
