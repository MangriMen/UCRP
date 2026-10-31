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

func (r *AuditSQL) List(ctx context.Context, entityType, entityID, actor string, limit, offset int) ([]domain.AuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `
		SELECT id, COALESCE(actor_user_id, ''), actor_role, action, entity_type, entity_id,
			COALESCE(details, ''), COALESCE(request_id, ''), created_at
		FROM audit_log WHERE 1=1`
	args := make([]any, 0, 5)
	if entityType != "" {
		args = append(args, entityType)
		query += ` AND entity_type = $` + itoa(len(args))
	}
	if entityID != "" {
		args = append(args, entityID)
		query += ` AND entity_id = $` + itoa(len(args))
	}
	if actor != "" {
		args = append(args, actor)
		query += ` AND actor_user_id = $` + itoa(len(args))
	}
	args = append(args, limit, offset)
	query += ` ORDER BY created_at DESC LIMIT $` + itoa(len(args)-1) + ` OFFSET $` + itoa(len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.AuditLog, 0)
	for rows.Next() {
		var a domain.AuditLog
		if err := rows.Scan(&a.ID, &a.ActorUserID, &a.ActorRole, &a.Action, &a.EntityType, &a.EntityID, &a.Details, &a.RequestID, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
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

type CampaignSQL struct {
	db *sql.DB
}

func NewCampaignSQL(db *sql.DB) *CampaignSQL {
	return &CampaignSQL{db: db}
}

func (r *CampaignSQL) List(ctx context.Context, limit, offset int) ([]domain.Campaign, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, tenant_id, segment_id, attribution_days, status, COALESCE(created_by, ''), created_at, updated_at
		FROM campaign ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.Campaign, 0)
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (r *CampaignSQL) GetByID(ctx context.Context, id string) (*domain.Campaign, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, tenant_id, segment_id, attribution_days, status, COALESCE(created_by, ''), created_at, updated_at
		FROM campaign WHERE id = $1`, id)
	c, err := scanCampaign(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return c, err
}

func (r *CampaignSQL) Insert(ctx context.Context, c *domain.Campaign) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO campaign (id, name, tenant_id, segment_id, attribution_days, status, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		c.ID, c.Name, c.TenantID, c.SegmentID, c.AttributionDays, c.Status, nullIfEmpty(c.CreatedBy), c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *CampaignSQL) Update(ctx context.Context, c *domain.Campaign) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE campaign SET name = $2, segment_id = $3, attribution_days = $4, status = $5, updated_at = $6
		WHERE id = $1`,
		c.ID, c.Name, c.SegmentID, c.AttributionDays, c.Status, c.UpdatedAt,
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

func scanCampaign(s scanner) (*domain.Campaign, error) {
	var c domain.Campaign
	if err := s.Scan(&c.ID, &c.Name, &c.TenantID, &c.SegmentID, &c.AttributionDays, &c.Status, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
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
