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
	defer func() {
		if err := store.Close(); err != nil {
			log.Error("close store", "err", err)
		}
	}()

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
	if err := bootstrapManager(ctx, store.DB, clock, idg); err != nil {
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
	accessSvc := service.NewAccessGrantService(repository.NewAccessSQL(store.DB), clock, idg, users)
	accessSvc.SetTenantRepository(repository.NewTenantSQL(store.DB))
	accessSvc.SetAuditRepository(repository.NewAuditSQL(store.DB))
	authSvc.SetAccessService(accessSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	accessHandler := handler.NewAccessGrantHandler(accessSvc)
	invitationHandler := handler.NewInvitationHandler(service.NewInvitationService(repository.NewInvitationSQL(store.DB), accessSvc, nil, clock, idg, users))
	userHandler := handler.NewUserHandler(authSvc)
	auditHandler := handler.NewAuditHandler(service.NewAuditService(repository.NewAuditSQL(store.DB)))
	campaignHandler := handler.NewCampaignHandler(service.NewCampaignService(repository.NewCampaignSQL(store.DB)))
	analyticsSvc := service.NewAnalyticsService(store.DB)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc)
	tenantAccessHandler := handler.NewTenantAccessHandler(accessSvc, analyticsSvc)
	q := queue.NewDBQueue(store.DB, clock, idg)
	guests := repository.NewGuestSQL(store.DB)
	identifiers := repository.NewGuestIdentifierSQL(store.DB)
	events := repository.NewEventSQL(store.DB)
	contacts := repository.NewContactLogSQL(store.DB)
	profileSvc := service.NewProfileService(guests, identifiers, events, contacts)
	guestHandler := handler.NewGuestHandler(profileSvc)
	tenantGuestSvc := service.NewTenantGuestService(repository.NewTenantGuestSQL(store.DB), accessSvc)
	tenantGuestHandler := handler.NewTenantGuestHandler(tenantGuestSvc, guestHandler)
	mergeSQL := repository.NewMergeSQL(store.DB)
	blocks := repository.NewMergeBlockSQL(store.DB)
	identitySvc := service.NewIdentityService(mergeSQL, blocks, mergeSQL, q)
	identityHandler := handler.NewIdentityHandler(identitySvc, profileSvc)
	segmentRepo := repository.NewSegmentSQL(store.DB)
	if err := segmentRepo.EnsureTemplates(ctx, clock, idg); err != nil {
		return err
	}
	segmentSvc := service.NewSegmentService(segmentRepo, segments.NewSqlEvaluator(store.DB), q)
	segmentHandler := handler.NewSegmentHandler(segmentSvc)
	recommender := recommend.NewWeightedRecommender(store.DB, cfg.Recommend)
	recoSvc := service.NewRecommendationService(guests, repository.NewRecommendationSQL(store.DB), recommender, q)
	recoHandler := handler.NewRecommendationHandler(recoSvc)
	exportSvc := service.NewExportService(
		repository.NewExportSQL(store.DB),
		repository.NewContactLimitSQL(store.DB),
		contacts,
		segmentRepo,
	)
	exportHandler := handler.NewExportHandler(exportSvc)
	limitSvc := service.NewContactLimitService(repository.NewContactLimitSQL(store.DB))
	limitHandler := handler.NewContactLimitHandler(limitSvc)
	dataSources := repository.NewDataSourceSQL(store.DB)
	imports := service.NewImportService(
		repository.NewImportBatchSQL(store.DB),
		repository.NewRejectedRowSQL(store.DB),
		dataSources,
		events,
		catalog,
		blobs,
		q,
	)
	dataSourceSvc := service.NewDataSourceService(dataSources, sourceTypes)
	importHandler := handler.NewImportHandler(imports, dataSourceSvc)
	dataSourceHandler := handler.NewDataSourceHandler(dataSourceSvc)
	refSvc := service.NewReferenceService(
		repository.NewTenantSQL(store.DB),
		repository.NewTenantCategorySQL(store.DB),
		repository.NewProjectSQL(store.DB),
		repository.NewEventTypeSQL(store.DB),
	)
	refHandler := handler.NewReferenceHandler(refSvc)
	tenantReferenceHandler := handler.NewTenantReferenceHandler(accessSvc, refSvc, refHandler)

	r := chi.NewRouter()
	r.Use(httpx.RequestIDMiddleware)
	r.Use(httpx.RecoverMiddleware(log))
	r.Use(httpx.TimeoutMiddleware(cfg.HTTPTimeout))

	r.Get("/healthz", authHandler.Healthz)
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/login", authHandler.Login)
		api.Post("/auth/refresh", authHandler.Refresh)
		api.Post("/auth/logout", authHandler.Logout)
		api.Post("/auth/invitations/accept", invitationHandler.Accept)
		api.Group(func(priv chi.Router) {
			priv.Use(auth.Middleware(tokens))
			priv.Use(auth.RequireActiveUser(users))
			priv.Use(handler.DeniedAuditMiddleware(repository.NewAuditSQL(store.DB), idg, clock))
			priv.Use(handler.PermissionMiddleware(accessSvc))
			priv.Get("/auth/me", authHandler.Me)
			priv.Route("/access", func(a chi.Router) {
				a.Get("/permissions", accessHandler.Permissions)
				a.Get("/presets", accessHandler.Presets)
				a.Get("/effective", accessHandler.EffectiveSubject)
				a.Post("/simulate", accessHandler.Simulate)
				a.Get("/audit", accessHandler.Audit)
				a.Get("/grants", accessHandler.List)
				a.Post("/grants", accessHandler.Create)
				a.Get("/grants/{id}", accessHandler.Get)
				a.Patch("/grants/{id}", accessHandler.Update)
				a.Post("/grants/{id}/activate", accessHandler.Activate)
				a.Post("/grants/{id}/suspend", accessHandler.Suspend)
				a.Post("/grants/{id}/revoke", accessHandler.Revoke)
			})
			priv.Route("/guests", func(g chi.Router) {
				g.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/", guestHandler.Search)
				g.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst, domain.RoleTenant)).Get("/{id}", tenantGuestHandler.Get)
				g.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst, domain.RoleTenant)).Get("/{id}/events", tenantGuestHandler.Events)
				g.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/{id}/contacts", guestHandler.Contacts)
				g.With(auth.RequireRole(domain.RoleAnalyst)).Post("/{id}/merge", identityHandler.Merge)
				g.With(auth.RequireRole(domain.RoleAnalyst)).Delete("/{id}", guestHandler.Delete)
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
			priv.Route("/merge-blocks", func(u chi.Router) {
				u.With(auth.RequireRole(domain.RoleAnalyst)).Get("/", identityHandler.ListBlocks)
				u.With(auth.RequireRole(domain.RoleAnalyst)).Delete("/{id}", identityHandler.DeleteBlock)
			})
			priv.Route("/segments", func(s chi.Router) {
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/", segmentHandler.List)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/templates", segmentHandler.Templates)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Post("/preview", segmentHandler.Preview)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Post("/", segmentHandler.Create)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/{id}", segmentHandler.Get)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Put("/{id}", segmentHandler.Update)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Delete("/{id}", segmentHandler.Delete)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/{id}/members", segmentHandler.Members)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/{id}/versions", segmentHandler.Versions)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Post("/{id}/recompute", segmentHandler.Recompute)
				s.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/{id}/recommendations", recoHandler.SegmentRecommendations)
			})
			priv.Route("/exports", func(x chi.Router) {
				x.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/", exportHandler.List)
				x.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/{id}", exportHandler.Get)
				x.With(auth.RequireRole(domain.RoleMarketer)).Post("/preview", exportHandler.Preview)
				x.With(auth.RequireRole(domain.RoleMarketer)).Post("/", exportHandler.Create)
				x.With(auth.RequireRole(domain.RoleMarketer)).Get("/{id}/members", exportHandler.Members)
				x.With(auth.RequireRole(domain.RoleMarketer)).Get("/{id}/download", exportHandler.Download)
			})
			priv.Route("/contact-limits", func(c chi.Router) {
				c.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/", limitHandler.List)
				c.With(auth.RequireRole(domain.RoleMarketer)).Post("/", limitHandler.Create)
				c.With(auth.RequireRole(domain.RoleMarketer)).Patch("/{id}", limitHandler.Update)
				c.With(auth.RequireRole(domain.RoleMarketer)).Delete("/{id}", limitHandler.Delete)
			})
			priv.Route("/tenants", func(t chi.Router) {
				t.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst, domain.RoleTenant)).Get("/", tenantReferenceHandler.List)
				t.With(auth.RequireRole(domain.RoleAnalyst)).Post("/", refHandler.CreateTenant)
				t.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst, domain.RoleTenant)).Get("/{id}", tenantReferenceHandler.Get)
				t.With(auth.RequireRole(domain.RoleAnalyst)).Patch("/{id}", refHandler.UpdateTenant)
				t.With(auth.RequireRole(domain.RoleAnalyst)).Delete("/{id}", refHandler.DeleteTenant)
			})
			priv.Route("/tenant-categories", func(c chi.Router) {
				c.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/", refHandler.ListCategories)
				c.With(auth.RequireRole(domain.RoleAnalyst)).Post("/", refHandler.CreateCategory)
				c.With(auth.RequireRole(domain.RoleAnalyst)).Patch("/{id}", refHandler.UpdateCategory)
				c.With(auth.RequireRole(domain.RoleAnalyst)).Delete("/{id}", refHandler.DeleteCategory)
				c.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/{id}/relations", refHandler.ListRelations)
				c.With(auth.RequireRole(domain.RoleAnalyst)).Put("/{id}/relations", refHandler.ReplaceRelations)
			})
			priv.Route("/projects", func(p chi.Router) {
				p.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/", refHandler.ListProjects)
				p.With(auth.RequireRole(domain.RoleAnalyst)).Post("/", refHandler.CreateProject)
				p.With(auth.RequireRole(domain.RoleAnalyst)).Patch("/{id}", refHandler.UpdateProject)
				p.With(auth.RequireRole(domain.RoleAnalyst)).Delete("/{id}", refHandler.DeleteProject)
			})
			priv.Route("/event-types", func(e chi.Router) {
				e.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/", refHandler.ListEventTypes)
				e.With(auth.RequireRole(domain.RoleAnalyst)).Post("/", refHandler.CreateEventType)
				e.With(auth.RequireRole(domain.RoleAnalyst)).Patch("/{id}", refHandler.UpdateEventType)
				e.With(auth.RequireRole(domain.RoleAnalyst)).Delete("/{id}", refHandler.DeleteEventType)
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
				d.With(auth.RequireRole(domain.RoleAnalyst)).Patch("/{id}", dataSourceHandler.Update)
				d.With(auth.RequireRole(domain.RoleAnalyst)).Delete("/{id}", dataSourceHandler.Delete)
				d.With(auth.RequireRole(domain.RoleAnalyst)).Post("/{id}/sync", importHandler.Sync)
			})
			priv.With(auth.RequireRole(domain.RoleMarketer, domain.RoleAnalyst)).Get("/segment-templates", segmentHandler.Templates)
			priv.With(auth.RequireRole(domain.RoleAnalyst, domain.RoleManager, domain.RoleMarketer)).Get("/data-source-types", dataSourceHandler.ListTypes)
			priv.With(auth.RequireRole(domain.RoleAnalyst)).Post("/data-source-types", dataSourceHandler.CreateType)
			priv.With(auth.RequireRole(domain.RoleAnalyst)).Patch("/data-source-types/{id}", dataSourceHandler.UpdateType)
			priv.Route("/users", func(u chi.Router) {
				u.With(auth.RequireRole(domain.RoleManager)).Post("/invitations", invitationHandler.Issue)
				u.With(auth.RequireRole(domain.RoleManager)).Post("/{id}/recovery", invitationHandler.Recover)
				u.With(auth.RequireRole(domain.RoleManager)).Get("/", userHandler.List)
				u.With(auth.RequireRole(domain.RoleManager)).Post("/", userHandler.Create)
				u.With(auth.RequireRole(domain.RoleManager)).Get("/{id}", userHandler.Get)
				u.With(auth.RequireRole(domain.RoleManager)).Patch("/{id}", userHandler.Update)
				u.With(auth.RequireRole(domain.RoleManager)).Delete("/{id}", userHandler.Delete)
			})
			priv.With(auth.RequireRole(domain.RoleManager, domain.RoleAnalyst)).Get("/audit-logs", auditHandler.List)
			priv.Route("/campaigns", func(c chi.Router) {
				c.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/", campaignHandler.List)
				c.With(auth.RequireRole(domain.RoleMarketer)).Post("/", campaignHandler.Create)
				c.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/{id}", campaignHandler.Get)
				c.With(auth.RequireRole(domain.RoleMarketer)).Patch("/{id}", campaignHandler.Update)
				c.With(auth.RequireRole(domain.RoleMarketer)).Delete("/{id}", campaignHandler.Delete)
			})
			priv.Route("/dashboard", func(d chi.Router) {
				d.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/cross-tenant", analyticsHandler.CrossTenantReport)
				d.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst)).Get("/metrics", analyticsHandler.Metrics)
			})
			priv.With(auth.RequireRole(domain.RoleMarketer, domain.RoleManager, domain.RoleAnalyst, domain.RoleTenant)).Get("/tenants/{id}/aggregates", tenantAccessHandler.Aggregates)
			priv.With(auth.RequireRole(domain.RoleManager, domain.RoleAnalyst)).Get("/data-quality/sources", analyticsHandler.DataQuality)
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
