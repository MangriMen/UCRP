package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MangriMen/UCRP/internal/domain"
)

type AuditSQL struct {
	db *sql.DB
}

func NewAuditSQL(db *sql.DB) *AuditSQL {
	return &AuditSQL{db: db}
}

func (r *AuditSQL) List(ctx context.Context, filter domain.AuditFilter, limit, offset int) ([]domain.AuditLog, int, error) {
	limit, offset = clampPage(limit, offset)
	where := ` WHERE TRUE`
	args := make([]any, 0, 8)
	add := func(condition string, value any) {
		args = append(args, value)
		where += ` AND ` + condition + itoa(len(args))
	}
	if filter.EntityType != "" {
		add("entity_type = $", filter.EntityType)
	}
	if filter.EntityID != "" {
		add("entity_id = $", filter.EntityID)
	}
	if filter.Actor != "" {
		add("actor_user_id = $", filter.Actor)
	}
	if filter.Action != "" {
		add("action = $", filter.Action)
	}
	if filter.From != nil {
		add("created_at >= $", *filter.From)
	}
	if filter.To != nil {
		add("created_at <= $", *filter.To)
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	query := `
		SELECT id, COALESCE(actor_user_id, ''), actor_role, action, entity_type, entity_id,
			COALESCE(details, ''), COALESCE(request_id, ''), created_at
		FROM audit_log` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(len(args)-1) + ` OFFSET $` + itoa(len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.AuditLog, 0)
	for rows.Next() {
		var a domain.AuditLog
		if err := rows.Scan(&a.ID, &a.ActorUserID, &a.ActorRole, &a.Action, &a.EntityType, &a.EntityID, &a.Details, &a.RequestID, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	return out, total, rows.Err()
}

func (r *AuditSQL) Insert(ctx context.Context, a *domain.AuditLog) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, request_id, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.ID, nullIfEmpty(a.ActorUserID), a.ActorRole, a.Action, a.EntityType, a.EntityID,
		nullIfEmpty(a.Details), nullIfEmpty(a.RequestID), a.CreatedAt,
	)
	return err
}

const (
	campaignColumns = `c.id, c.name, COALESCE(c.target_tenant_id, ''), COALESCE(t.name, ''), COALESCE(c.segment_id, ''), COALESCE(sg.name, ''),
		c.attribution_days, c.status, c.starts_at, c.ends_at, COALESCE(c.created_by, ''), c.created_at, c.created_at`
	campaignFrom = `campaign c
		LEFT JOIN tenant t ON t.id = c.target_tenant_id
		LEFT JOIN segment sg ON sg.id = c.segment_id`
)

type CampaignSQL struct {
	db *sql.DB
}

func NewCampaignSQL(db *sql.DB) *CampaignSQL {
	return &CampaignSQL{db: db}
}

func (r *CampaignSQL) List(ctx context.Context, filter domain.CampaignFilter, limit, offset int) ([]domain.Campaign, int, error) {
	limit, offset = clampPage(limit, offset)
	where := ` WHERE TRUE`
	args := make([]any, 0, 4)
	if filter.Status != "" {
		args = append(args, filter.Status)
		where += ` AND c.status = $` + itoa(len(args))
	}
	if filter.SegmentID != "" {
		args = append(args, filter.SegmentID)
		where += ` AND c.segment_id = $` + itoa(len(args))
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+campaignFrom+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+campaignColumns+`
		FROM `+campaignFrom+where+` ORDER BY c.created_at DESC LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.Campaign, 0)
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *c)
	}
	return out, total, rows.Err()
}

func (r *CampaignSQL) GetByID(ctx context.Context, id string) (*domain.Campaign, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+campaignColumns+`
		FROM `+campaignFrom+` WHERE c.id = $1`, id)
	c, err := scanCampaign(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return c, err
}

func (r *CampaignSQL) Insert(ctx context.Context, c *domain.Campaign) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO campaign (id, name, target_tenant_id, segment_id, attribution_days, status, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		c.ID, c.Name, nullIfEmpty(c.TenantID), nullIfEmpty(c.SegmentID), c.AttributionDays, c.Status, nullIfEmpty(c.CreatedBy), c.CreatedAt,
	)
	return err
}

func (r *CampaignSQL) Update(ctx context.Context, c *domain.Campaign) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE campaign SET name = $2, target_tenant_id = $3, segment_id = $4, attribution_days = $5, status = $6
		WHERE id = $1`,
		c.ID, c.Name, nullIfEmpty(c.TenantID), nullIfEmpty(c.SegmentID), c.AttributionDays, c.Status,
	)
	return err
}

func (r *CampaignSQL) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM campaign WHERE id = $1`, id)
	return err
}

func (r *CampaignSQL) HasExports(ctx context.Context, campaignID string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM export_job WHERE campaign_id = $1`, campaignID).Scan(&n)
	return n > 0, err
}

func (r *CampaignSQL) Conversion(ctx context.Context, filter domain.ConversionFilter) (domain.Conversion, error) {
	var out domain.Conversion
	err := r.db.QueryRowContext(ctx, `
		WITH members AS (
			SELECT m.guest_id, m.is_control, c.target_tenant_id, c.attribution_days,
				COALESCE(cl.contacted_at, j.created_at) AS contacted_at
			FROM export_job j
			JOIN campaign c ON c.id = j.campaign_id
			JOIN export_member m ON m.export_id = j.id
			LEFT JOIN contact_log cl ON cl.export_id = j.id AND cl.guest_id = m.guest_id
			WHERE (m.is_control OR cl.id IS NOT NULL)
				AND ($1 = '' OR c.id = $1)
				AND ($2::timestamp IS NULL OR j.created_at >= $2)
				AND ($3::timestamp IS NULL OR j.created_at <= $3)
		),
		scored AS (
			SELECT mb.is_control, EXISTS (
				SELECT 1 FROM event e
				WHERE e.guest_id = mb.guest_id AND e.tenant_id = mb.target_tenant_id
					AND e.occurred_at > mb.contacted_at
					AND e.occurred_at <= mb.contacted_at + make_interval(days => mb.attribution_days::int)
			) AS converted
			FROM members mb
		)
		SELECT
			COUNT(*) FILTER (WHERE NOT is_control),
			COUNT(*) FILTER (WHERE is_control),
			COUNT(*) FILTER (WHERE NOT is_control AND converted),
			COUNT(*) FILTER (WHERE is_control AND converted)
		FROM scored`, filter.CampaignID, nullTime(filter.From), nullTime(filter.To),
	).Scan(&out.Recipients, &out.ControlGroup, &out.Conversions, &out.ControlConversions)
	return out, err
}

func scanCampaign(s scanner) (*domain.Campaign, error) {
	var (
		c        domain.Campaign
		startsAt sql.NullTime
		endsAt   sql.NullTime
	)
	if err := s.Scan(&c.ID, &c.Name, &c.TenantID, &c.TenantName, &c.SegmentID, &c.SegmentName, &c.AttributionDays, &c.Status, &startsAt, &endsAt, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	if startsAt.Valid {
		c.StartsAt = &startsAt.Time
	}
	if endsAt.Valid {
		c.EndsAt = &endsAt.Time
	}
	return &c, nil
}

func itoa(n int) string {
	return strconvItoa(n)
}

func strconvItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
