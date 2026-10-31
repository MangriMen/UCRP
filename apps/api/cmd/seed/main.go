package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/MangriMen/UCRP/internal/platform/config"
	"github.com/MangriMen/UCRP/internal/repository/sqlstore"
	"github.com/MangriMen/UCRP/migrations"
	"github.com/MangriMen/UCRP/seed"
)

func main() {
	var (
		dsn    = flag.String("dsn", "", "database URL (defaults to UCRP_DATABASE_URL)")
		reset  = flag.Bool("reset", false, "remove demo data and exit")
		dryRun = flag.Bool("dry-run", false, "print the plan without touching the database")
		guests = flag.Int("guests", 92, "number of background guests to generate (10..2000)")
	)
	flag.Parse()

	if err := run(*dsn, *reset, *dryRun, *guests); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

func run(dsn string, reset, dryRun bool, guests int) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if dsn == "" {
		dsn = os.Getenv("UCRP_DATABASE_URL")
	}
	if dsn == "" && !dryRun {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("database url is required: set -dsn or UCRP_DATABASE_URL")
		}
		dsn = cfg.DatabaseURL
	}

	if dryRun {
		return seed.Run(ctx, nil, seed.Options{Guests: guests, DryRun: true, Out: os.Stdout})
	}

	store, err := sqlstore.Open(dsn)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	if err := sqlstore.Migrate(ctx, store.DB, migrations.FS, "common"); err != nil {
		return err
	}

	return seed.Run(ctx, store.DB, seed.Options{
		Guests: guests,
		Reset:  reset,
		Out:    os.Stdout,
	})
}
