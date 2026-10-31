package seed

import (
	"context"
	"database/sql"
	"fmt"
)

type check struct {
	Name  string
	Query string
	Args  []any
	Want  string
}

func (s *Seeder) demoGuestIDs() []string {
	out := make([]string, 0, s.opts.Guests+len(demoPersonas())+12)
	for i := 0; i < s.opts.Guests; i++ {
		out = append(out, demoID("guest", backgroundGuestKey(i)))
	}
	for _, p := range demoPersonas() {
		out = append(out, demoID("guest", p.Key))
	}
	for _, g := range demoMergedGuests() {
		out = append(out, demoID("guest", g.Key))
	}
	for _, g := range demoDeletedGuests() {
		out = append(out, demoID("guest", g.Key))
	}
	return out
}

func (s *Seeder) verify(ctx context.Context) error {
	segmentIDs := make([]string, 0, len(demoSegments()))
	for _, sg := range demoSegments() {
		segmentIDs = append(segmentIDs, demoID("segment", sg.Key))
	}
	guestIDs := s.demoGuestIDs()

	checks := []check{
		{
			Name: "import_batch rows balance",
			Query: `SELECT COUNT(*) FROM import_batch
				WHERE total_rows <> accepted_rows + rejected_rows + duplicate_rows`,
			Want: "0",
		},
		{
			Name: "segment_version member_count matches members",
			Query: `SELECT COUNT(*) FROM segment_version v
				WHERE v.status = 'ready'
				  AND v.member_count IS DISTINCT FROM (
					SELECT COUNT(*) FROM segment_member m WHERE m.segment_version_id = v.id)`,
			Want: "0",
		},
		{
			Name: "segment rules have was and was_not",
			Query: `SELECT COUNT(*) FROM segment_version
				WHERE segment_id = ANY($1)
				  AND (rule LIKE '%"was":[]%' OR rule LIKE '%"was_not":[]%')`,
			Args: []any{segmentIDs},
			Want: "0",
		},
		{
			Name: "recommendation targets unvisited tenants",
			Query: `SELECT COUNT(*) FROM recommendation r
				WHERE r.guest_id = ANY($1)
				  AND EXISTS (
					SELECT 1 FROM event e
					JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
					WHERE e.guest_id = r.guest_id AND e.tenant_id = r.tenant_id)`,
			Args: []any{guestIDs},
			Want: "0",
		},
		{
			Name: "export_job counters balance",
			Query: `SELECT COUNT(*) FROM export_job
				WHERE idempotency_key LIKE 'demo-%'
				  AND status = 'ready'
				  AND total_in_segment <> no_contact_count + blocked_by_limit_count + control_count + exported_count`,
			Want: "0",
		},
		{
			Name: "contact_log only for non-control members",
			Query: `SELECT COUNT(*) FROM contact_log c
				JOIN export_member m ON m.export_id = c.export_id AND m.guest_id = c.guest_id
				WHERE m.is_control`,
			Want: "0",
		},
		{
			Name: "contact_log identifier belongs to guest",
			Query: `SELECT COUNT(*) FROM export_member m
				JOIN guest_identifier i ON i.id = m.identifier_id
				WHERE i.guest_id <> m.guest_id`,
			Want: "0",
		},
		{
			Name:  "merge_candidate ordering",
			Query: `SELECT COUNT(*) FROM merge_candidate WHERE guest_a_id >= guest_b_id`,
			Want:  "0",
		},
		{
			Name:  "unmerge_block ordering",
			Query: `SELECT COUNT(*) FROM unmerge_block WHERE guest_a_id >= guest_b_id`,
			Want:  "0",
		},
		{
			Name: "merged guests have target",
			Query: `SELECT COUNT(*) FROM guest
				WHERE (status = 'merged') <> (merged_into_id IS NOT NULL)`,
			Want: "0",
		},
		{
			Name: "event match_keys match identifiers",
			Query: `SELECT COUNT(*) FROM event e
				JOIN guest_identifier i ON i.guest_id = e.guest_id
				WHERE e.match_keys IS NOT NULL
				  AND e.match_keys <> '{}'
				  AND e.match_keys::jsonb ? i.id_type
				  AND e.match_keys::jsonb ->> i.id_type <> i.value_normalized`,
			Want: "0",
		},
		{
			Name: "reversal events reference purchase",
			Query: `SELECT COUNT(*) FROM event e
				JOIN event_type et ON et.id = e.event_type_id
				WHERE et.code = 'reversal' AND e.reverses_event_id IS NULL`,
			Want: "0",
		},
		{
			Name: "duplicate guest identifiers",
			Query: `SELECT COUNT(*) FROM (
					SELECT guest_id, id_type, value_normalized FROM guest_identifier
					GROUP BY 1,2,3 HAVING COUNT(*) > 1) d`,
			Want: "0",
		},
		{
			Name: "duplicate events per source",
			Query: `SELECT COUNT(*) FROM (
					SELECT data_source_id, event_type_id, external_id FROM event
					GROUP BY 1,2,3 HAVING COUNT(*) > 1) d`,
			Want: "0",
		},
		{
			Name: "orphan foreign keys",
			Query: `SELECT
					(SELECT COUNT(*) FROM event e LEFT JOIN tenant t ON t.id = e.tenant_id WHERE t.id IS NULL)
				  + (SELECT COUNT(*) FROM event e LEFT JOIN event_type et ON et.id = e.event_type_id WHERE et.id IS NULL)
				  + (SELECT COUNT(*) FROM guest_identifier i LEFT JOIN guest g ON g.id = i.guest_id WHERE g.id IS NULL)
				  + (SELECT COUNT(*) FROM segment_member m LEFT JOIN guest g ON g.id = m.guest_id WHERE g.id IS NULL)
				  + (SELECT COUNT(*) FROM export_member m LEFT JOIN guest_identifier i ON i.id = m.identifier_id WHERE i.id IS NULL)`,
			Want: "0",
		},
	}

	s.println("\nverification")
	failed := 0
	for _, c := range checks {
		got, err := s.scalar(ctx, c.Query, c.Args...)
		if err != nil {
			return fmt.Errorf("verify %q: %w", c.Name, err)
		}
		status := "ok"
		if got != c.Want {
			status = "FAIL"
			failed++
		}
		s.printf("  %-46s %-4s (got %s, want %s)\n", c.Name, status, got, c.Want)
	}
	if failed > 0 {
		return fmt.Errorf("%d verification checks failed", failed)
	}
	return s.report(ctx)
}

func (s *Seeder) scalar(ctx context.Context, query string, args ...any) (string, error) {
	var value sql.NullString
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&value); err != nil {
		return "", err
	}
	if !value.Valid {
		return "", nil
	}
	return value.String, nil
}

func (s *Seeder) report(ctx context.Context) error {
	tenantIDs := make([]string, 0, len(demoTenants()))
	for _, t := range demoTenants() {
		tenantIDs = append(tenantIDs, demoID("tenant", t.Code))
	}
	sourceIDs := make([]string, 0, len(demoSources()))
	for _, d := range demoSources() {
		sourceIDs = append(sourceIDs, demoID("data_source", d.Key))
	}
	segmentIDs := make([]string, 0, len(demoSegments()))
	for _, sg := range demoSegments() {
		segmentIDs = append(segmentIDs, demoID("segment", sg.Key))
	}
	guestIDs := s.demoGuestIDs()

	rows := []struct {
		label string
		query string
		args  []any
	}{
		{"tenant_category", `SELECT COUNT(*) FROM tenant_category WHERE id = ANY($1)`, []any{s.categoryIDs()}},
		{"tenant_category_relation", `SELECT COUNT(*) FROM tenant_category_relation WHERE category_id = ANY($1)`, []any{s.categoryIDs()}},
		{"tenant", `SELECT COUNT(*) FROM tenant WHERE id = ANY($1)`, []any{tenantIDs}},
		{"project", `SELECT COUNT(*) FROM project WHERE tenant_id = ANY($1)`, []any{tenantIDs}},
		{"event_type", `SELECT COUNT(*) FROM event_type`, nil},
		{"data_source_type", `SELECT COUNT(*) FROM data_source_type`, nil},
		{"data_source", `SELECT COUNT(*) FROM data_source WHERE id = ANY($1)`, []any{sourceIDs}},
		{"app_user", `SELECT COUNT(*) FROM app_user WHERE id = ANY($1)`, []any{s.userIDs()}},
		{"guest", `SELECT COUNT(*) FROM guest WHERE id = ANY($1)`, []any{guestIDs}},
		{"guest_identifier", `SELECT COUNT(*) FROM guest_identifier WHERE guest_id = ANY($1)`, []any{guestIDs}},
		{"import_batch", `SELECT COUNT(*) FROM import_batch WHERE data_source_id = ANY($1)`, []any{sourceIDs}},
		{"rejected_row", `SELECT COUNT(*) FROM rejected_row WHERE import_batch_id IN (SELECT id FROM import_batch WHERE data_source_id = ANY($1))`, []any{sourceIDs}},
		{"event", `SELECT COUNT(*) FROM event WHERE data_source_id = ANY($1)`, []any{sourceIDs}},
		{"merge_candidate", `SELECT COUNT(*) FROM merge_candidate WHERE guest_a_id = ANY($1)`, []any{guestIDs}},
		{"unmerge_block", `SELECT COUNT(*) FROM unmerge_block WHERE guest_a_id = ANY($1)`, []any{guestIDs}},
		{"segment", `SELECT COUNT(*) FROM segment WHERE id = ANY($1)`, []any{segmentIDs}},
		{"segment_version", `SELECT COUNT(*) FROM segment_version WHERE segment_id = ANY($1)`, []any{segmentIDs}},
		{"segment_member", `SELECT COUNT(*) FROM segment_member WHERE segment_version_id IN (SELECT id FROM segment_version WHERE segment_id = ANY($1))`, []any{segmentIDs}},
		{"recommendation", `SELECT COUNT(*) FROM recommendation WHERE guest_id = ANY($1)`, []any{guestIDs}},
		{"recommendation_factor", `SELECT COUNT(*) FROM recommendation_factor WHERE recommendation_id IN (SELECT id FROM recommendation WHERE guest_id = ANY($1))`, []any{guestIDs}},
		{"campaign", `SELECT COUNT(*) FROM campaign WHERE id = ANY($1)`, []any{s.campaignIDs()}},
		{"contact_limit", `SELECT COUNT(*) FROM contact_limit WHERE id = ANY($1)`, []any{s.limitIDs()}},
		{"export_job", `SELECT COUNT(*) FROM export_job WHERE idempotency_key LIKE 'demo-%'`, nil},
		{"export_member", `SELECT COUNT(*) FROM export_member WHERE export_id IN (SELECT id FROM export_job WHERE idempotency_key LIKE 'demo-%')`, nil},
		{"contact_log", `SELECT COUNT(*) FROM contact_log WHERE export_id IN (SELECT id FROM export_job WHERE idempotency_key LIKE 'demo-%')`, nil},
		{"audit_log", `SELECT COUNT(*) FROM audit_log WHERE id = ANY($1)`, []any{s.auditIDs()}},
		{"job_queue", `SELECT COUNT(*) FROM job_queue WHERE id = ANY($1)`, []any{s.jobIDs()}},
	}

	s.println("\ndemo row counts")
	for _, r := range rows {
		count, err := s.scalar(ctx, r.query, r.args...)
		if err != nil {
			return err
		}
		s.printf("  %-26s %s\n", r.label, count)
	}
	return nil
}

func (s *Seeder) categoryIDs() []string {
	out := make([]string, 0, len(demoCategories()))
	for _, c := range demoCategories() {
		out = append(out, demoID("tenant_category", c.Code))
	}
	return out
}

func (s *Seeder) userIDs() []string {
	out := make([]string, 0, len(demoUsers()))
	for _, u := range demoUsers() {
		out = append(out, demoID("app_user", u.Key))
	}
	return out
}

func (s *Seeder) campaignIDs() []string {
	out := make([]string, 0, len(demoCampaigns()))
	for _, c := range demoCampaigns() {
		out = append(out, demoID("campaign", c.Key))
	}
	return out
}

func (s *Seeder) limitIDs() []string {
	out := make([]string, 0, len(demoContactLimits()))
	for _, l := range demoContactLimits() {
		out = append(out, demoID("contact_limit", l.Key))
	}
	return out
}

func (s *Seeder) auditIDs() []string {
	out := make([]string, 0, len(demoAuditEntries()))
	for _, a := range demoAuditEntries() {
		out = append(out, demoID("audit_log", a.Key))
	}
	return out
}

func (s *Seeder) jobIDs() []string {
	out := make([]string, 0, len(demoJobs()))
	for _, j := range demoJobs() {
		out = append(out, demoID("job_queue", j.Key))
	}
	return out
}
