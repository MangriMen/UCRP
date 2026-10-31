package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/identity"
	"github.com/MangriMen/UCRP/internal/platform/config"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
	"github.com/MangriMen/UCRP/internal/platform/logging"
	"github.com/MangriMen/UCRP/internal/platform/queue"
	"github.com/MangriMen/UCRP/internal/platform/storage"
	"github.com/MangriMen/UCRP/internal/recommend"
	"github.com/MangriMen/UCRP/internal/repository"
	"github.com/MangriMen/UCRP/internal/repository/sqlstore"
	"github.com/MangriMen/UCRP/internal/segments"
	"github.com/MangriMen/UCRP/internal/service"
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
	defer func() {
		if err := store.Close(); err != nil {
			log.Error("close store", "err", err)
		}
	}()

	clock := domain.SystemClock{}
	q := queue.NewDBQueue(store.DB, clock, idgen.UUID{})
	workerID := "worker-1"
	blobs, err := storage.NewLocal(cfg.UploadDir)
	if err != nil {
		return err
	}
	dataSources := repository.NewDataSourceSQL(store.DB)
	catalog := repository.NewImportCatalogSQL(store.DB)
	events := repository.NewEventSQL(store.DB)
	imports := service.NewImportService(
		repository.NewImportBatchSQL(store.DB),
		repository.NewRejectedRowSQL(store.DB),
		dataSources,
		events,
		catalog,
		blobs,
		q,
	)
	resolver := identity.NewResolver(
		repository.NewGuestSQL(store.DB),
		repository.NewGuestIdentifierSQL(store.DB),
		events,
		repository.NewMergeSQL(store.DB),
		repository.NewMergeBlockSQL(store.DB),
	)
	segmentRepo := repository.NewSegmentSQL(store.DB)
	segmentSvc := service.NewSegmentService(segmentRepo, segments.NewSqlEvaluator(store.DB), q)
	recommender := recommend.NewWeightedRecommender(store.DB, cfg.Recommend)
	presence := repository.NewPresenceSQL(store.DB)
	jobTypes := []string{"import.process", "connector.pull", "identity.resolve", "segment.recompute", "recommend.refresh"}
	maxAttempts := 3
	expiryCtx, cancelExpiry := context.WithCancel(ctx)
	defer cancelExpiry()
	go runGrantExpiry(expiryCtx, service.NewGrantExpiry(repository.NewAccessSQL(store.DB), clock), log)

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
		if err := handleJob(ctx, imports, resolver, segmentSvc, segmentRepo, recommender, presence, q, job); err != nil {
			log.Error("job failed", "id", job.ID, "type", job.JobType, "err", err)
			if job.Attempts >= maxAttempts {
				if failErr := q.Fail(ctx, job.ID, err.Error(), time.Time{}); failErr != nil {
					log.Error("fail job", "id", job.ID, "err", failErr)
				}
				continue
			}
			retryAfter := time.Now().UTC().Add(time.Minute * time.Duration(job.Attempts))
			if failErr := q.Fail(ctx, job.ID, err.Error(), retryAfter); failErr != nil {
				log.Error("retry job", "id", job.ID, "err", failErr)
			}
			continue
		}
		if err := q.Complete(ctx, job.ID); err != nil {
			log.Error("complete job", "id", job.ID, "err", err)
		}
	}
}

func handleJob(
	ctx context.Context,
	imports *service.ImportService,
	resolver *identity.Resolver,
	segmentSvc *service.SegmentService,
	segmentRepo domain.SegmentRepo,
	recommender *recommend.WeightedRecommender,
	presence *repository.PresenceSQL,
	q domain.Queue,
	job *domain.Job,
) error {
	var payload struct {
		BatchID   string `json:"batch_id"`
		SegmentID string `json:"segment_id"`
		VersionID string `json:"version_id"`
		GuestID   string `json:"guest_id"`
	}
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return err
	}
	switch job.JobType {
	case "identity.resolve":
		if payload.BatchID == "" {
			return domain.ErrValidation
		}
		if err := resolver.ResolveBatch(ctx, payload.BatchID); err != nil {
			return err
		}
		if err := enqueueRecommendRefreshForBatch(ctx, presence, q, payload.BatchID); err != nil {
			return err
		}
		return enqueueActiveSegmentRecompute(ctx, segmentRepo, q)
	case "connector.pull":
		if payload.BatchID == "" {
			return domain.ErrValidation
		}
		return imports.ProcessBatch(ctx, payload.BatchID)
	case "segment.recompute":
		if payload.SegmentID == "" {
			return domain.ErrValidation
		}
		return segmentSvc.RecomputeVersion(ctx, payload.SegmentID, payload.VersionID)
	case "recommend.refresh":
		if payload.GuestID == "" {
			return domain.ErrValidation
		}
		_, err := recommender.Refresh(ctx, payload.GuestID)
		return err
	default:
		if payload.BatchID == "" {
			return domain.ErrValidation
		}
		return imports.ProcessBatch(ctx, payload.BatchID)
	}
}

func enqueueRecommendRefreshForBatch(ctx context.Context, presence *repository.PresenceSQL, q domain.Queue, batchID string) error {
	guestIDs, err := presence.GuestIDsByBatch(ctx, batchID)
	if err != nil {
		return err
	}
	for _, guestID := range guestIDs {
		payload, err := json.Marshal(map[string]string{"guest_id": guestID})
		if err != nil {
			return err
		}
		if err := q.Enqueue(ctx, domain.Job{
			JobType:  "recommend.refresh",
			Payload:  string(payload),
			DedupKey: "reco:" + guestID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func enqueueActiveSegmentRecompute(ctx context.Context, segmentsRepo domain.SegmentRepo, q domain.Queue) error {
	items, err := segmentsRepo.List(ctx, 100, 0)
	if err != nil {
		return err
	}
	for _, seg := range items {
		if seg.Status != domain.SegmentActive || seg.IsTemplate {
			continue
		}
		payload, err := json.Marshal(map[string]string{"segment_id": seg.ID})
		if err != nil {
			return err
		}
		if err := q.Enqueue(ctx, domain.Job{
			JobType:  "segment.recompute",
			Payload:  string(payload),
			DedupKey: "segment:" + seg.ID + ":auto",
		}); err != nil {
			return err
		}
	}
	return nil
}
