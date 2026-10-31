package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

func (r *ContactLogSQL) ListByGuest(ctx context.Context, guestID string, limit, offset int) ([]domain.ContactLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT cl.id, cl.guest_id, cl.export_id, ej.channel, cl.contacted_at
		FROM contact_log cl
		JOIN export_job ej ON ej.id = cl.export_id
		WHERE cl.guest_id = $1
		ORDER BY cl.contacted_at DESC LIMIT $2 OFFSET $3`, guestID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.ContactLog, 0)
	for rows.Next() {
		var l domain.ContactLog
		if err := rows.Scan(&l.ID, &l.GuestID, &l.ExportID, &l.Channel, &l.ContactedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

type TenantSQL struct {
	db *sql.DB
}

func NewTenantSQL(db *sql.DB) *TenantSQL {
	return &TenantSQL{db: db}
}

func (r *TenantSQL) List(ctx context.Context, limit, offset int) ([]domain.Tenant, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, tenantSelect+` ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanTenants(rows)
}

func (r *TenantSQL) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	row := r.db.QueryRowContext(ctx, tenantSelect+` WHERE id = $1 AND deleted_at IS NULL`, id)
	t, err := scanTenant(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

func (r *TenantSQL) Insert(ctx context.Context, t *domain.Tenant) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant (id, code, name, mall_id, category_id, kind, floor, priority, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		t.ID, t.Code, t.Name, t.MallID, t.CategoryID, t.Kind, t.Floor, t.Priority, t.IsActive, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return domain.ErrConflict
	}
	return err
}

func (r *TenantSQL) Update(ctx context.Context, t *domain.Tenant) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant SET name = $2, mall_id = $3, category_id = $4, kind = $5, floor = $6, priority = $7, is_active = $8, updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL`,
		t.ID, t.Name, t.MallID, t.CategoryID, t.Kind, t.Floor, t.Priority, t.IsActive, t.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TenantSQL) SoftDelete(ctx context.Context, id string, now time.Time) error {
	var events int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM event WHERE tenant_id = $1`, id).Scan(&events); err != nil {
		return err
	}
	if events > 0 {
		return domain.ErrConflict
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant SET is_active = FALSE, deleted_at = $2, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`, id, now)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

const tenantSelect = `
	SELECT id, code, name, mall_id, category_id, kind, floor, priority, is_active, created_at, updated_at, deleted_at
	FROM tenant`

func scanTenants(rows *sql.Rows) ([]domain.Tenant, error) {
	out := make([]domain.Tenant, 0)
	for rows.Next() {
		t, err := scanTenant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func scanTenant(s scanner) (*domain.Tenant, error) {
	var (
		t   domain.Tenant
		del sql.NullTime
	)
	if err := s.Scan(&t.ID, &t.Code, &t.Name, &t.MallID, &t.CategoryID, &t.Kind, &t.Floor, &t.Priority, &t.IsActive, &t.CreatedAt, &t.UpdatedAt, &del); err != nil {
		return nil, err
	}
	if del.Valid {
		t.DeletedAt = &del.Time
	}
	return &t, nil
}

type TenantCategorySQL struct {
	db *sql.DB
}

func NewTenantCategorySQL(db *sql.DB) *TenantCategorySQL {
	return &TenantCategorySQL{db: db}
}

func (r *TenantCategorySQL) List(ctx context.Context) ([]domain.TenantCategory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, is_active, created_at FROM tenant_category WHERE is_active ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.TenantCategory, 0)
	for rows.Next() {
		var c domain.TenantCategory
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *TenantCategorySQL) GetByID(ctx context.Context, id string) (*domain.TenantCategory, error) {
	var c domain.TenantCategory
	err := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, is_active, created_at FROM tenant_category WHERE id = $1`, id).
		Scan(&c.ID, &c.Code, &c.Name, &c.IsActive, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &c, err
}

func (r *TenantCategorySQL) Insert(ctx context.Context, c *domain.TenantCategory) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_category (id, code, name, is_active, created_at) VALUES ($1,$2,$3,$4,$5)`,
		c.ID, c.Code, c.Name, c.IsActive, c.CreatedAt,
	)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return domain.ErrConflict
	}
	return err
}

func (r *TenantCategorySQL) Update(ctx context.Context, c *domain.TenantCategory) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_category SET name = $2, is_active = $3 WHERE id = $1`, c.ID, c.Name, c.IsActive)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TenantCategorySQL) Delete(ctx context.Context, id string) error {
	var tenants int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenant WHERE category_id = $1 AND deleted_at IS NULL`, id).Scan(&tenants); err != nil {
		return err
	}
	if tenants > 0 {
		return domain.ErrConflict
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM tenant_category WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TenantCategorySQL) ListRelations(ctx context.Context, categoryID string) ([]domain.CategoryRelation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT category_id, related_category_id, relation_type, weight_pct
		FROM tenant_category_relation WHERE category_id = $1`, categoryID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.CategoryRelation, 0)
	for rows.Next() {
		var rel domain.CategoryRelation
		if err := rows.Scan(&rel.CategoryID, &rel.RelatedCategoryID, &rel.RelationType, &rel.WeightPct); err != nil {
			return nil, err
		}
		out = append(out, rel)
	}
	return out, rows.Err()
}

func (r *TenantCategorySQL) ReplaceRelations(ctx context.Context, categoryID string, rels []domain.CategoryRelation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM tenant_category_relation WHERE category_id = $1`, categoryID); err != nil {
		return err
	}
	for _, rel := range rels {
		weight := rel.WeightPct
		if weight <= 0 {
			weight = 100
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tenant_category_relation (category_id, related_category_id, relation_type, weight_pct)
			VALUES ($1,$2,$3,$4)`, categoryID, rel.RelatedCategoryID, rel.RelationType, weight); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type ProjectSQL struct {
	db *sql.DB
}

func NewProjectSQL(db *sql.DB) *ProjectSQL {
	return &ProjectSQL{db: db}
}

func (r *ProjectSQL) List(ctx context.Context, tenantID string, limit, offset int) ([]domain.Project, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `SELECT id, tenant_id, code, name, is_active FROM project WHERE 1=1`
	args := make([]any, 0, 3)
	if tenantID != "" {
		args = append(args, tenantID)
		query += ` AND tenant_id = $1`
	}
	args = append(args, limit, offset)
	query += ` ORDER BY code LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.Project, 0)
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Code, &p.Name, &p.IsActive); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProjectSQL) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	var p domain.Project
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, is_active FROM project WHERE id = $1`, id).
		Scan(&p.ID, &p.TenantID, &p.Code, &p.Name, &p.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *ProjectSQL) Insert(ctx context.Context, p *domain.Project) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO project (id, tenant_id, code, name, is_active) VALUES ($1,$2,$3,$4,$5)`,
		p.ID, p.TenantID, p.Code, p.Name, p.IsActive,
	)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return domain.ErrConflict
	}
	return err
}

func (r *ProjectSQL) Update(ctx context.Context, p *domain.Project) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE project SET name = $2, is_active = $3 WHERE id = $1`, p.ID, p.Name, p.IsActive)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProjectSQL) Delete(ctx context.Context, id string) error {
	var events int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM event WHERE project_id = $1`, id).Scan(&events); err != nil {
		return err
	}
	if events > 0 {
		return domain.ErrConflict
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM project WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type EventTypeSQL struct {
	db *sql.DB
}

func NewEventTypeSQL(db *sql.DB) *EventTypeSQL {
	return &EventTypeSQL{db: db}
}

func (r *EventTypeSQL) List(ctx context.Context) ([]domain.EventType, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, has_amount, counts_as_presence, is_active FROM event_type ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.EventType, 0)
	for rows.Next() {
		var e domain.EventType
		if err := rows.Scan(&e.ID, &e.Code, &e.Name, &e.HasAmount, &e.CountsAsPresence, &e.IsActive); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *EventTypeSQL) GetByID(ctx context.Context, id string) (*domain.EventType, error) {
	var e domain.EventType
	err := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, has_amount, counts_as_presence, is_active FROM event_type WHERE id = $1`, id).
		Scan(&e.ID, &e.Code, &e.Name, &e.HasAmount, &e.CountsAsPresence, &e.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &e, err
}

func (r *EventTypeSQL) Insert(ctx context.Context, e *domain.EventType) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO event_type (id, code, name, has_amount, counts_as_presence, is_active)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		e.ID, e.Code, e.Name, e.HasAmount, e.CountsAsPresence, e.IsActive,
	)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return domain.ErrConflict
	}
	return err
}

func (r *EventTypeSQL) Update(ctx context.Context, e *domain.EventType) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE event_type SET name = $2, has_amount = $3, counts_as_presence = $4, is_active = $5 WHERE id = $1`,
		e.ID, e.Name, e.HasAmount, e.CountsAsPresence, e.IsActive,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *EventTypeSQL) Delete(ctx context.Context, id string) error {
	var events int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM event WHERE event_type_id = $1`, id).Scan(&events); err != nil {
		return err
	}
	if events > 0 {
		return domain.ErrConflict
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM event_type WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
