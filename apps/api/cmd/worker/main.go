package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/config"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
	"github.com/MangriMen/UCRP/internal/platform/logging"
	"github.com/MangriMen/UCRP/internal/platform/queue"
	"github.com/MangriMen/UCRP/internal/repository/sqlstore"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := logging.New(cfg.LogLevel)
	if err := run(log, cfg); err != nil {
		log.Error("exit", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger, cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := sqlstore.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()

	clock := domain.SystemClock{}
	q := queue.NewDBQueue(store.DB, clock, idgen.UUID{})
	workerID := "worker-1"
	jobTypes := []string{}

	log.Info("worker started")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		job, err := q.Take(ctx, jobTypes, workerID)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Error("take job", "err", err)
			time.Sleep(time.Second)
			continue
		}
		if job == nil {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(500 * time.Millisecond):
			}
			continue
		}

		log.Info("job taken", "id", job.ID, "type", job.JobType)
		if err := q.Complete(ctx, job.ID); err != nil {
			log.Error("complete job", "id", job.ID, "err", err)
		}
	}
}
