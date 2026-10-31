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
	authSvc := service.NewAuthService(users, tokens, clock, idg)
	authHandler := handler.NewAuthHandler(authSvc)
	guests := repository.NewGuestSQL(store.DB)
	identifiers := repository.NewGuestIdentifierSQL(store.DB)
	events := repository.NewEventSQL(store.DB)
	profileSvc := service.NewProfileService(guests, identifiers, events)
	guestHandler := handler.NewGuestHandler(profileSvc)
	identitySvc := service.NewIdentityService(
		repository.NewMergeCandidateSQL(store.DB),
		repository.NewUnmergeBlockSQL(store.DB),
		guests,
		identifiers,
		events,
	)
	identityHandler := handler.NewIdentityHandler(identitySvc)
	segmentSvc := service.NewSegmentService(repository.NewSegmentSQL(store.DB), segments.NewSqlEvaluator(store.DB))
	segmentHandler := handler.NewSegmentHandler(segmentSvc)

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
				g.Get("/", guestHandler.Search)
				g.Get("/{id}", guestHandler.Get)
				g.Get("/{id}/events", guestHandler.Events)
			})
			priv.Route("/merge-candidates", func(m chi.Router) {
				m.Get("/", identityHandler.List)
				m.Get("/{id}", identityHandler.Get)
				m.Post("/{id}/confirm", identityHandler.Confirm)
				m.Post("/{id}/reject", identityHandler.Reject)
				m.Post("/{id}/revert", identityHandler.Revert)
			})
			priv.Route("/segments", func(s chi.Router) {
				s.Get("/", segmentHandler.List)
				s.Post("/preview", segmentHandler.Preview)
				s.Post("/", segmentHandler.Create)
				s.Get("/{id}", segmentHandler.Get)
			})
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
