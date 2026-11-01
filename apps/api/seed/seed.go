package seed

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"
)

const (
	demoPassword  = "demo12345"
	defaultGuests = 92
	minGuests     = 10
	maxGuests     = 2000
)

type Options struct {
	Guests int
	Reset  bool
	DryRun bool
	Out    io.Writer
}

type Seeder struct {
	db   *sql.DB
	opts Options
	now  time.Time
	out  io.Writer

	categories  map[string]string
	tenants     map[string]string
	projects    map[string]string
	users       map[string]string
	sources     map[string]string
	eventTypes  map[string]string
	guests      map[string]string
	identifiers map[string]string
	identValues map[string]string
	segments    map[string]string
	versions    map[string]string
	campaigns   map[string]string
	batches     map[string]string
	exports     map[string]string
}

func Run(ctx context.Context, db *sql.DB, opts Options) error {
	if opts.Guests <= 0 {
		opts.Guests = defaultGuests
	}
	if opts.Guests < minGuests {
		opts.Guests = minGuests
	}
	if opts.Guests > maxGuests {
		opts.Guests = maxGuests
	}
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	s := &Seeder{
		db:          db,
		opts:        opts,
		now:         time.Now().UTC().Truncate(time.Second),
		out:         opts.Out,
		categories:  map[string]string{},
		tenants:     map[string]string{},
		projects:    map[string]string{},
		users:       map[string]string{},
		sources:     map[string]string{},
		eventTypes:  map[string]string{},
		guests:      map[string]string{},
		identifiers: map[string]string{},
		identValues: map[string]string{},
		segments:    map[string]string{},
		versions:    map[string]string{},
		campaigns:   map[string]string{},
		batches:     map[string]string{},
		exports:     map[string]string{},
	}
	if opts.DryRun {
		return s.describe()
	}
	if err := s.cleanup(ctx); err != nil {
		return fmt.Errorf("cleanup: %w", err)
	}
	if opts.Reset {
		s.println("demo data removed")
		return nil
	}

	stages := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"reference", s.seedReference},
		{"users", s.seedUsers},
		{"tenant grants", s.seedTenantGrants},
		{"data sources", s.seedSources},
		{"import batches", s.seedBatches},
		{"guests", s.seedGuests},
		{"events", s.seedEvents},
		{"identity", s.seedIdentity},
		{"segments", s.seedSegments},
		{"recommendations", s.seedRecommendations},
		{"campaigns", s.seedCampaigns},
		{"contact limits", s.seedContactLimits},
		{"exports", s.seedExports},
		{"audit", s.seedAudit},
		{"jobs", s.seedJobs},
	}
	for _, st := range stages {
		if err := st.fn(ctx); err != nil {
			return fmt.Errorf("%s: %w", st.name, err)
		}
		s.printf("  %-18s ok\n", st.name)
	}
	return s.verify(ctx)
}

func (s *Seeder) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(s.out, format, args...)
}

func (s *Seeder) println(args ...any) {
	_, _ = fmt.Fprintln(s.out, args...)
}

func (s *Seeder) describe() error {
	guests := s.buildGuests()
	events := s.buildEvents()
	s.printf("demo data plan (guests %d)\n", s.opts.Guests)
	s.printf("  tenant_category          %d\n", len(demoCategories()))
	s.printf("  tenant_category_relation %d\n", len(demoRelations()))
	s.printf("  mall                     %d\n", len(demoMalls()))
	s.printf("  tenant                   %d\n", len(demoTenants()))
	s.printf("  project                  %d\n", len(demoProjects()))
	s.printf("  app_user                 %d\n", len(demoUsers()))
	s.printf("  data_source              %d\n", len(demoSources()))
	s.printf("  import_batch             %d\n", len(demoBatches()))
	s.printf("  guest                    %d\n", len(guests))
	s.printf("  event                    %d\n", len(events))
	s.printf("  segment                  %d\n", len(demoSegments()))
	s.printf("  campaign                 %d\n", len(demoCampaigns()))
	s.printf("  export_job               %d\n", len(demoExports()))
	return nil
}

func (s *Seeder) exec(ctx context.Context, query string, args ...any) error {
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Seeder) leftoverGuestIDs(ctx context.Context, sourceIDs []string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT guest_id FROM guest_identifier WHERE data_source_id = ANY($1)
		UNION
		SELECT DISTINCT guest_id FROM event WHERE data_source_id = ANY($1) AND guest_id IS NOT NULL`,
		sourceIDs)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *Seeder) cleanup(ctx context.Context) error {
	categoryIDs := make([]string, 0, len(demoCategories()))
	for _, c := range demoCategories() {
		categoryIDs = append(categoryIDs, demoID("tenant_category", c.Code))
	}
	tenantIDs := make([]string, 0, len(demoTenants()))
	for _, t := range demoTenants() {
		tenantIDs = append(tenantIDs, demoID("tenant", t.Code))
	}
	mallIDs := make([]string, 0, len(demoMalls()))
	for _, m := range demoMalls() {
		mallIDs = append(mallIDs, demoID("mall", m.Code))
	}
	projectIDs := make([]string, 0, len(demoProjects()))
	for _, p := range demoProjects() {
		projectIDs = append(projectIDs, demoID("project", p.Tenant+"/"+p.Code))
	}
	userIDs := make([]string, 0, len(demoUsers()))
	for _, u := range demoUsers() {
		userIDs = append(userIDs, demoID("app_user", u.Key))
	}
	grantIDs := make([]string, 0, len(demoTenantGrants()))
	for _, g := range demoTenantGrants() {
		grantIDs = append(grantIDs, demoID("access_grant", g.UserKey))
	}
	sourceIDs := make([]string, 0, len(demoSources()))
	for _, d := range demoSources() {
		sourceIDs = append(sourceIDs, demoID("data_source", d.Key))
	}
	batchIDs := make([]string, 0, len(demoBatches()))
	for _, b := range demoBatches() {
		batchIDs = append(batchIDs, demoID("import_batch", b.Key))
	}
	segmentIDs := make([]string, 0, len(demoSegments()))
	for _, sg := range demoSegments() {
		segmentIDs = append(segmentIDs, demoID("segment", sg.Key))
	}
	campaignIDs := make([]string, 0, len(demoCampaigns()))
	for _, c := range demoCampaigns() {
		campaignIDs = append(campaignIDs, demoID("campaign", c.Key))
	}
	limitIDs := make([]string, 0, len(demoContactLimits()))
	for _, l := range demoContactLimits() {
		limitIDs = append(limitIDs, demoID("contact_limit", l.Key))
	}
	candidateIDs := make([]string, 0, len(demoMergeCandidates()))
	for _, c := range demoMergeCandidates() {
		candidateIDs = append(candidateIDs, demoID("merge_candidate", c.Key))
	}
	blockIDs := make([]string, 0, len(demoMergeBlocks()))
	for _, b := range demoMergeBlocks() {
		blockIDs = append(blockIDs, demoID("merge_block", b.Key))
	}
	auditIDs := make([]string, 0, len(demoAuditEntries()))
	for _, a := range demoAuditEntries() {
		auditIDs = append(auditIDs, demoID("audit_log", a.Key))
	}
	jobIDs := make([]string, 0, len(demoJobs()))
	for _, j := range demoJobs() {
		jobIDs = append(jobIDs, demoID("job_queue", j.Key))
	}

	guestIDs := make([]string, 0, s.opts.Guests+len(demoPersonas())+12)
	for i := 0; i < s.opts.Guests; i++ {
		guestIDs = append(guestIDs, demoID("guest", backgroundGuestKey(i)))
	}
	for _, p := range demoPersonas() {
		guestIDs = append(guestIDs, demoID("guest", p.Key))
	}
	for _, g := range demoMergedGuests() {
		guestIDs = append(guestIDs, demoID("guest", g.Key))
	}
	for _, g := range demoDeletedGuests() {
		guestIDs = append(guestIDs, demoID("guest", g.Key))
	}
	leftovers, err := s.leftoverGuestIDs(ctx, sourceIDs)
	if err != nil {
		return err
	}
	guestIDs = append(guestIDs, leftovers...)

	steps := []struct {
		query string
		args  []any
	}{
		{`DELETE FROM contact_log WHERE export_id IN (SELECT id FROM export_job WHERE idempotency_key LIKE 'demo-%')`, nil},
		{`DELETE FROM export_member WHERE export_id IN (SELECT id FROM export_job WHERE idempotency_key LIKE 'demo-%')`, nil},
		{`DELETE FROM export_job WHERE idempotency_key LIKE 'demo-%'`, nil},
		{`DELETE FROM campaign WHERE id = ANY($1)`, []any{campaignIDs}},
		{`DELETE FROM contact_limit WHERE id = ANY($1)`, []any{limitIDs}},
		{`DELETE FROM segment_member WHERE segment_version_id IN (SELECT id FROM segment_version WHERE segment_id = ANY($1))`, []any{segmentIDs}},
		{`DELETE FROM segment_version WHERE segment_id = ANY($1)`, []any{segmentIDs}},
		{`DELETE FROM segment WHERE id = ANY($1)`, []any{segmentIDs}},
		{`DELETE FROM recommendation WHERE guest_id = ANY($1) OR tenant_id = ANY($2)`, []any{guestIDs, tenantIDs}},
		{`DELETE FROM merge_block WHERE id = ANY($1)`, []any{blockIDs}},
		{`DELETE FROM merge_candidate WHERE id = ANY($1)`, []any{candidateIDs}},
		{`UPDATE event SET reverses_event_id = NULL WHERE data_source_id = ANY($1)`, []any{sourceIDs}},
		{`DELETE FROM event WHERE data_source_id = ANY($1)`, []any{sourceIDs}},
		{`DELETE FROM guest_identifier WHERE data_source_id = ANY($1)`, []any{sourceIDs}},
		{`UPDATE guest SET status = 'active', merged_into_id = NULL WHERE id = ANY($1)`, []any{guestIDs}},
		{`DELETE FROM guest WHERE id = ANY($1)`, []any{guestIDs}},
		{`DELETE FROM rejected_row WHERE import_batch_id = ANY($1)`, []any{batchIDs}},
		{`DELETE FROM import_batch WHERE id = ANY($1)`, []any{batchIDs}},
		{`DELETE FROM data_source WHERE id = ANY($1)`, []any{sourceIDs}},
		{`DELETE FROM project WHERE id = ANY($1)`, []any{projectIDs}},
		{`DELETE FROM audit_log WHERE id = ANY($1)`, []any{auditIDs}},
		{`DELETE FROM job_queue WHERE id = ANY($1)`, []any{jobIDs}},
		{`DELETE FROM access_grant_scope WHERE grant_id = ANY($1)`, []any{grantIDs}},
		{`DELETE FROM access_grant_item WHERE grant_id = ANY($1)`, []any{grantIDs}},
		{`DELETE FROM access_grant WHERE id = ANY($1)`, []any{grantIDs}},
		{`DELETE FROM app_user WHERE id = ANY($1)`, []any{userIDs}},
		{`DELETE FROM tenant WHERE id = ANY($1)`, []any{tenantIDs}},
		{`DELETE FROM mall WHERE id = ANY($1)`, []any{mallIDs}},
		{`DELETE FROM tenant_category_relation WHERE category_id = ANY($1)`, []any{categoryIDs}},
		{`DELETE FROM tenant_category WHERE id = ANY($1)`, []any{categoryIDs}},
	}
	for _, step := range steps {
		if err := s.exec(ctx, step.query, step.args...); err != nil {
			return err
		}
	}
	return nil
}
