package seed

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/MangriMen/UCRP/internal/repository"
	"github.com/MangriMen/UCRP/internal/service"
)

type limitSpec struct {
	Key         string
	Channel     string
	MaxContacts int
	PeriodDays  int
}

func demoContactLimits() []limitSpec {
	return []limitSpec{
		{"limit-sms", "sms", 2, 7},
		{"limit-email", "email", 3, 14},
		{"limit-global", "", 2, 7},
	}
}

func (s *Seeder) seedContactLimits(ctx context.Context) error {
	for _, l := range demoContactLimits() {
		var channel any
		if l.Channel != "" {
			channel = l.Channel
		}
		if err := s.exec(ctx, `
			INSERT INTO contact_limit (id, channel, max_contacts, period_days, is_active, created_by, created_at, updated_at)
			VALUES ($1,$2,$3,$4,TRUE,$5,$6,$6)
			ON CONFLICT (id) DO NOTHING`,
			demoID("contact_limit", l.Key), channel, l.MaxContacts, l.PeriodDays,
			s.users["marketer"], s.now.AddDate(0, -1, 0)); err != nil {
			return err
		}
	}
	return nil
}

type exportSpec struct {
	Key            string
	Segment        string
	Campaign       string
	Channel        string
	Format         string
	ControlPercent int
	IdempotencyKey string
}

func demoExports() []exportSpec {
	return []exportSpec{
		{"exp-cinema-sms", "seg-sleep-cinema", "camp-cinema", "sms", "csv", 33, "demo-export-1"},
		{"exp-single-sms", "tpl-single", "", "sms", "csv", 0, "demo-export-2"},
		{"exp-cinema-email", "seg-sleep-cinema", "camp-cinema", "email", "api", 20, "demo-export-3"},
	}
}

func (s *Seeder) seedExports(ctx context.Context) error {
	if err := s.seedBackdatedExport(ctx); err != nil {
		return err
	}
	svc := service.NewExportService(
		repository.NewExportSQL(s.db),
		repository.NewContactLimitSQL(s.db),
		repository.NewContactLogSQL(s.db),
		repository.NewSegmentSQL(s.db),
	)
	for _, e := range demoExports() {
		dto, err := svc.Create(ctx, service.ExportCreateRequest{
			SegmentID:      s.segments[e.Segment],
			Channel:        e.Channel,
			Format:         e.Format,
			ControlPercent: e.ControlPercent,
			CampaignID:     s.campaigns[e.Campaign],
			IdempotencyKey: e.IdempotencyKey,
		}, s.users["marketer"])
		if err != nil {
			return fmt.Errorf("export %s: %w", e.Key, err)
		}
		s.exports[e.Key] = dto.ID
	}
	return s.seedPendingExports(ctx)
}

func (s *Seeder) seedBackdatedExport(ctx context.Context) error {
	guestID := s.guests["g-sms"]
	identifierID := s.identifiers["g-sms/phone"]
	if guestID == "" || identifierID == "" {
		return fmt.Errorf("backdated export: g-sms identifiers missing")
	}
	segmentID := s.segments["seg-sleep-cinema"]
	versionID := s.segmentCurrentVersion("seg-sleep-cinema")
	if segmentID == "" || versionID == "" {
		return fmt.Errorf("backdated export: segment missing")
	}
	exportID := demoID("export_job", "exp-cinema-sms-old")
	s.exports["exp-cinema-sms-old"] = exportID
	createdAt := s.now.AddDate(0, 0, -5)
	if err := s.exec(ctx, `
		INSERT INTO export_job (id, segment_id, segment_version_id, campaign_id, channel, format, control_percent, control_seed,
			idempotency_key, status, total_in_segment, no_contact_count, blocked_by_limit_count, control_count, exported_count,
			created_by, created_at, finished_at)
		VALUES ($1,$2,$3,$4,'sms','csv',0,$5,'demo-export-4','ready',1,0,0,0,1,$6,$7,$7)
		ON CONFLICT (id) DO NOTHING`,
		exportID, segmentID, versionID, s.campaigns["camp-cinema"], hashSeed("demo-export-4"),
		s.users["marketer"], createdAt); err != nil {
		return err
	}
	if err := s.exec(ctx, `
		INSERT INTO export_member (export_id, guest_id, identifier_id, is_control)
		VALUES ($1,$2,$3,FALSE)
		ON CONFLICT DO NOTHING`, exportID, guestID, identifierID); err != nil {
		return err
	}
	if err := s.exec(ctx, `
		INSERT INTO contact_log (id, guest_id, export_id, contacted_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT DO NOTHING`,
		demoID("contact_log", "exp-cinema-sms-old/g-sms"), guestID, exportID, createdAt); err != nil {
		return err
	}
	return s.seedSecondSMSContact(ctx, guestID, identifierID, segmentID, versionID)
}

func (s *Seeder) seedSecondSMSContact(ctx context.Context, guestID, identifierID, segmentID, versionID string) error {
	exportID := demoID("export_job", "exp-cinema-sms-older")
	s.exports["exp-cinema-sms-older"] = exportID
	createdAt := s.now.AddDate(0, 0, -6)
	if err := s.exec(ctx, `
		INSERT INTO export_job (id, segment_id, segment_version_id, campaign_id, channel, format, control_percent, control_seed,
			idempotency_key, status, total_in_segment, no_contact_count, blocked_by_limit_count, control_count, exported_count,
			created_by, created_at, finished_at)
		VALUES ($1,$2,$3,$4,'sms','csv',0,$5,'demo-export-7','ready',1,0,0,0,1,$6,$7,$7)
		ON CONFLICT (id) DO NOTHING`,
		exportID, segmentID, versionID, s.campaigns["camp-cinema"], hashSeed("demo-export-7"),
		s.users["marketer"], createdAt); err != nil {
		return err
	}
	if err := s.exec(ctx, `
		INSERT INTO export_member (export_id, guest_id, identifier_id, is_control)
		VALUES ($1,$2,$3,FALSE)
		ON CONFLICT DO NOTHING`, exportID, guestID, identifierID); err != nil {
		return err
	}
	return s.exec(ctx, `
		INSERT INTO contact_log (id, guest_id, export_id, contacted_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT DO NOTHING`,
		demoID("contact_log", "exp-cinema-sms-older/g-sms"), guestID, exportID, createdAt)
}

func (s *Seeder) seedPendingExports(ctx context.Context) error {
	pending := []struct {
		Key            string
		Segment        string
		Channel        string
		Format         string
		Status         string
		ControlPercent int
		IdempotencyKey string
		DaysAgo        int
	}{
		{"exp-single-queued", "seg-single-archived", "sms", "csv", "queued", 0, "demo-export-5", 1},
		{"exp-high-value-failed", "tpl-high-value", "push", "csv", "failed", 0, "demo-export-6", 2},
	}
	for _, p := range pending {
		id := demoID("export_job", p.Key)
		s.exports[p.Key] = id
		segmentID := s.segments[p.Segment]
		versionID := s.segmentCurrentVersion(p.Segment)
		if segmentID == "" || versionID == "" {
			return fmt.Errorf("pending export %s: segment missing", p.Key)
		}
		createdAt := s.now.AddDate(0, 0, -p.DaysAgo)
		var finishedAt any
		if p.Status == "failed" {
			finishedAt = createdAt.Add(time.Minute)
		}
		if err := s.exec(ctx, `
			INSERT INTO export_job (id, segment_id, segment_version_id, channel, format, control_percent, control_seed,
				idempotency_key, status, total_in_segment, no_contact_count, blocked_by_limit_count, control_count, exported_count,
				created_by, created_at, finished_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,0,0,0,0,0,$10,$11,$12)
			ON CONFLICT (id) DO NOTHING`,
			id, segmentID, versionID, p.Channel, p.Format, p.ControlPercent, hashSeed(p.IdempotencyKey),
			p.IdempotencyKey, p.Status, s.users["marketer"], createdAt, finishedAt); err != nil {
			return err
		}
	}
	return nil
}

func hashSeed(key string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return int64(h.Sum64())
}
