package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/handler"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/config"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
	"github.com/MangriMen/UCRP/internal/platform/logging"
	"github.com/MangriMen/UCRP/internal/platform/queue"
	"github.com/MangriMen/UCRP/internal/platform/storage"
	"github.com/MangriMen/UCRP/internal/recommend"
	"github.com/MangriMen/UCRP/internal/repository"
	"github.com/MangriMen/UCRP/internal/repository/sqlstore"
	"github.com/MangriMen/UCRP/internal/segments"
	"github.com/MangriMen/UCRP/internal/service"
	"github.com/MangriMen/UCRP/migrations"
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

	if err := sqlstore.Migrate(ctx, store.DB, migrations.FS, "common"); err != nil {
		return err
	}

	clock := domain.SystemClock{}
	idg := idgen.UUID{}
	tokens := auth.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL)
	users := repository.NewUserSQL(store.DB)
	if err := seedAdmin(ctx, users, clock, idg); err != nil {
		return err
	}
	catalog := repository.NewImportCatalogSQL(store.DB)
	if err := catalog.EnsureEventTypes(ctx, clock, idg); err != nil {
		return err
	}
	sourceTypes := repository.NewDataSourceTypeSQL(store.DB)
	if err := sourceTypes.EnsureDefaults(ctx, clock, idg); err != nil {
		return err
	}
	blobs, err := storage.NewLocal(cfg.UploadDir)
	if err != nil {
		return err
	}
	authSvc := service.NewAuthService(users, tokens, clock, idg)
	authHandler := handler.NewAuthHandler(authSvc)
	guests := repository.NewGuestSQL(store.DB)
	identifiers := repository.NewGuestIdentifierSQL(store.DB)
	events := repository.NewEventSQL(store.DB)
	profileSvc := service.NewProfileService(guests, identifiers, events)
	guestHandler := handler.NewGuestHandler(profileSvc)
	identitySvc := service.NewIdentityService(repository.NewMergeSQL(store.DB))
	identityHandler := handler.NewIdentityHandler(identitySvc)
	segmentSvc := service.NewSegmentService(repository.NewSegmentSQL(store.DB), segments.NewSqlEvaluator(store.DB))
	segmentHandler := handler.NewSegmentHandler(segmentSvc)
	recoSvc := service.NewRecommendationService(guests, repository.NewRecommendationSQL(store.DB), recommend.NewWeightedRecommender(store.DB))
	recoHandler := handler.NewRecommendationHandler(recoSvc)
	exportSvc := service.NewExportService(
		repository.NewExportSQL(store.DB),
		repository.NewContactLimitSQL(store.DB),
		repository.NewContactLogSQL(store.DB),
		repository.NewSegmentSQL(store.DB),
	)
	exportHandler := handler.NewExportHandler(exportSvc)
	dataSources := repository.NewDataSourceSQL(store.DB)
	imports := service.NewImportService(
		repository.NewImportBatchSQL(store.DB),
		repository.NewRejectedRowSQL(store.DB),
		dataSources,
		events,
		catalog,
		blobs,
		queue.NewDBQueue(store.DB, clock, idg),
	)
	dataSourceSvc := service.NewDataSourceService(dataSources, sourceTypes)
	importHandler := handler.NewImportHandler(imports, dataSourceSvc)
	dataSourceHandler := handler.NewDataSourceHandler(dataSourceSvc)

	r := chi.NewRouter()
	r.Use(httpx.RequestIDMiddleware)
	r.Use(httpx.RecoverMiddleware(log))
	r.Use(httpx.TimeoutMiddleware(cfg.HTTPTimeout))

	r.Get("/healthz", authHandler.Healthz)
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/login", authHandler.Login)
		api.Group(func(priv chi.Router) {
			priv.Use(auth.Middleware(tokens))
			priv.Get("/auth/me", authHandler.Me)
			priv.Route("/guests", func(g chi.Router) {
				g.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/", guestHandler.Search)
				g.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/{id}", guestHandler.Get)
				g.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/{id}/events", guestHandler.Events)
				g.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/{id}/recommendations", recoHandler.List)
				g.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Post("/{id}/recommendations/refresh", recoHandler.Refresh)
			})
			priv.Route("/merge-candidates", func(m chi.Router) {
				m.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/", identityHandler.List)
				m.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/{id}", identityHandler.Get)
				m.With(auth.RequireRole(domain.RoleAnalyst)).Post("/{id}/confirm", identityHandler.Confirm)
				m.With(auth.RequireRole(domain.RoleAnalyst)).Post("/{id}/reject", identityHandler.Reject)
				m.With(auth.RequireRole(domain.RoleAnalyst)).Post("/{id}/revert", identityHandler.Revert)
			})
			priv.Route("/segments", func(s chi.Router) {
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/", segmentHandler.List)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Post("/preview", segmentHandler.Preview)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Post("/", segmentHandler.Create)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/{id}", segmentHandler.Get)
			})
			priv.Route("/exports", func(x chi.Router) {
				x.With(auth.RequireRole(domain.RoleMarketer)).Get("/", exportHandler.List)
				x.With(auth.RequireRole(domain.RoleMarketer)).Post("/preview", exportHandler.Preview)
				x.With(auth.RequireRole(domain.RoleMarketer)).Post("/", exportHandler.Create)
				x.With(auth.RequireRole(domain.RoleMarketer)).Get("/{id}", exportHandler.Get)
				x.With(auth.RequireRole(domain.RoleMarketer)).Get("/{id}/download", exportHandler.Download)
			})
			priv.Route("/imports", func(i chi.Router) {
				i.With(auth.RequireRole(domain.RoleAnalyst)).Post("/", importHandler.Create)
				i.With(auth.RequireRole(domain.RoleAnalyst, domain.RoleManager)).Get("/", importHandler.List)
				i.With(auth.RequireRole(domain.RoleAnalyst, domain.RoleManager)).Get("/{id}", importHandler.Get)
				i.With(auth.RequireRole(domain.RoleAnalyst)).Get("/{id}/rejected-rows", importHandler.RejectedRows)
			})
			priv.Route("/data-sources", func(d chi.Router) {
				d.With(auth.RequireRole(domain.RoleAnalyst, domain.RoleManager, domain.RoleMarketer)).Get("/", dataSourceHandler.List)
				d.With(auth.RequireRole(domain.RoleAnalyst)).Post("/", dataSourceHandler.Create)
				d.With(auth.RequireRole(domain.RoleAnalyst, domain.RoleManager, domain.RoleMarketer)).Get("/{id}", dataSourceHandler.Get)
				d.With(auth.RequireRole(domain.RoleAnalyst)).Post("/{id}/sync", importHandler.Sync)
			})
			priv.With(auth.RequireRole(domain.RoleAnalyst, domain.RoleManager, domain.RoleMarketer)).Get("/data-source-types", dataSourceHandler.ListTypes)
		})
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("api listening", "addr", cfg.HTTPAddr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
