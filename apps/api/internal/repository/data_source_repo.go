package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type DataSourceSQL struct {
	db *sql.DB
}

func NewDataSourceSQL(db *sql.DB) *DataSourceSQL {
	return &DataSourceSQL{db: db}
}

func (r *DataSourceSQL) GetByID(ctx context.Context, id string) (*domain.DataSource, error) {
	var (
		ds       domain.DataSource
		tenantID sql.NullString
		lastImp  sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT d.id, d.tenant_id, d.type_id, t.code, d.name, COALESCE(d.config, ''), COALESCE(d.field_mapping, ''),
			COALESCE(d.schedule_cron, ''), d.is_active, t.supports_file, t.supports_pull, d.last_import_at, d.created_at, d.updated_at
		FROM data_source d
		JOIN data_source_type t ON t.id = d.type_id
		WHERE d.id = $1`, id).Scan(
		&ds.ID, &tenantID, &ds.TypeID, &ds.TypeCode, &ds.Name, &ds.Config, &ds.FieldMapping,
		&ds.ScheduleCron, &ds.IsActive, &ds.SupportsFile, &ds.SupportsPull, &lastImp, &ds.CreatedAt, &ds.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if tenantID.Valid {
		ds.TenantID = tenantID.String
	}
	if lastImp.Valid {
		ds.LastImportAt = &lastImp.Time
	}
	return &ds, nil
}

func (r *DataSourceSQL) List(ctx context.Context, limit, offset int) ([]domain.DataSource, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.tenant_id, d.type_id, t.code, d.name, COALESCE(d.config, ''), COALESCE(d.field_mapping, ''),
			COALESCE(d.schedule_cron, ''), d.is_active, t.supports_file, t.supports_pull, d.last_import_at, d.created_at, d.updated_at
		FROM data_source d
		JOIN data_source_type t ON t.id = d.type_id
		ORDER BY d.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.DataSource, 0)
	for rows.Next() {
		var (
			ds       domain.DataSource
			tenantID sql.NullString
			lastImp  sql.NullTime
		)
		if err := rows.Scan(
			&ds.ID, &tenantID, &ds.TypeID, &ds.TypeCode, &ds.Name, &ds.Config, &ds.FieldMapping,
			&ds.ScheduleCron, &ds.IsActive, &ds.SupportsFile, &ds.SupportsPull, &lastImp, &ds.CreatedAt, &ds.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if tenantID.Valid {
			ds.TenantID = tenantID.String
		}
		if lastImp.Valid {
			ds.LastImportAt = &lastImp.Time
		}
		out = append(out, ds)
	}
	return out, rows.Err()
}

func (r *DataSourceSQL) Insert(ctx context.Context, ds *domain.DataSource) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO data_source (id, tenant_id, type_id, name, config, field_mapping, schedule_cron, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		ds.ID, nullIfEmpty(ds.TenantID), ds.TypeID, ds.Name, nullIfEmpty(ds.Config), nullIfEmpty(ds.FieldMapping),
		nullIfEmpty(ds.ScheduleCron), ds.IsActive, ds.CreatedAt, ds.UpdatedAt,
	)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return domain.ErrConflict
	}
	return err
}

func (r *DataSourceSQL) Update(ctx context.Context, ds *domain.DataSource) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE data_source SET name = $2, config = $3, field_mapping = $4, schedule_cron = $5, is_active = $6, updated_at = $7
		WHERE id = $1`,
		ds.ID, ds.Name, nullIfEmpty(ds.Config), nullIfEmpty(ds.FieldMapping), nullIfEmpty(ds.ScheduleCron), ds.IsActive, ds.UpdatedAt,
	)
	return err
}

func (r *DataSourceSQL) Deactivate(ctx context.Context, id string, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE data_source SET is_active = FALSE, updated_at = $2 WHERE id = $1`, id, now)
	return err
}

func (r *DataSourceSQL) TouchLastImport(ctx context.Context, id string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE data_source SET last_import_at = $2, updated_at = $2 WHERE id = $1`, id, at)
	return err
}

type DataSourceTypeSQL struct {
	db *sql.DB
}

func NewDataSourceTypeSQL(db *sql.DB) *DataSourceTypeSQL {
	return &DataSourceTypeSQL{db: db}
}

func (r *DataSourceTypeSQL) List(ctx context.Context) ([]domain.DataSourceType, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, connector_key, supports_file, supports_pull, is_active
		FROM data_source_type WHERE is_active ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.DataSourceType, 0)
	for rows.Next() {
		var t domain.DataSourceType
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.ConnectorKey, &t.SupportsFile, &t.SupportsPull, &t.IsActive); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *DataSourceTypeSQL) GetByID(ctx context.Context, id string) (*domain.DataSourceType, error) {
	var t domain.DataSourceType
	err := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, connector_key, supports_file, supports_pull, is_active
		FROM data_source_type WHERE id = $1`, id).
		Scan(&t.ID, &t.Code, &t.Name, &t.ConnectorKey, &t.SupportsFile, &t.SupportsPull, &t.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &t, err
}

func (r *DataSourceTypeSQL) Insert(ctx context.Context, t *domain.DataSourceType) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO data_source_type (id, code, name, connector_key, supports_file, supports_pull, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		t.ID, t.Code, t.Name, t.ConnectorKey, t.SupportsFile, t.SupportsPull, t.IsActive,
	)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return domain.ErrConflict
	}
	return err
}

func (r *DataSourceTypeSQL) Update(ctx context.Context, t *domain.DataSourceType) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE data_source_type SET name = $2, supports_file = $3, supports_pull = $4, is_active = $5
		WHERE id = $1`,
		t.ID, t.Name, t.SupportsFile, t.SupportsPull, t.IsActive,
	)
	return err
}

func (r *DataSourceTypeSQL) EnsureDefaults(ctx context.Context, clock domain.Clock, idGen domain.IDGen) error {
	defaults := []struct {
		code         string
		name         string
		connectorKey string
		supportsFile bool
		supportsPull bool
	}{
		{"csv_excel", "CSV/Excel", "csv", true, false},
		{"crm", "CRM", "csv", true, false},
		{"pos", "Касса/POS", "csv", true, false},
		{"api", "API", "api", false, true},
	}
	for _, d := range defaults {
		var exists int
		err := r.db.QueryRowContext(ctx, `SELECT 1 FROM data_source_type WHERE code = $1`, d.code).Scan(&exists)
		if err == nil {
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO data_source_type (id, code, name, connector_key, supports_file, supports_pull, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, TRUE)`,
			idGen.NewID(ctx), d.code, d.name, d.connectorKey, d.supportsFile, d.supportsPull,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

type ImportCatalogSQL struct {
	db *sql.DB
}

func NewImportCatalogSQL(db *sql.DB) *ImportCatalogSQL {
	return &ImportCatalogSQL{db: db}
}

func (r *ImportCatalogSQL) EventTypeID(ctx context.Context, code string) (string, bool, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM event_type WHERE code = $1 AND is_active`, strings.ToLower(strings.TrimSpace(code))).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (r *ImportCatalogSQL) TenantIDByCode(ctx context.Context, code string) (string, bool, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM tenant WHERE code = $1 AND is_active AND deleted_at IS NULL`, strings.TrimSpace(code)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (r *ImportCatalogSQL) ProjectID(ctx context.Context, tenantID, code string) (string, bool, error) {
	if strings.TrimSpace(code) == "" {
		return "", false, nil
	}
	var id string
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM project WHERE tenant_id = $1 AND code = $2 AND is_active`, tenantID, strings.TrimSpace(code)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (r *ImportCatalogSQL) EnsureEventTypes(ctx context.Context, clock domain.Clock, idGen domain.IDGen) error {
	defaults := []struct {
		code           string
		name           string
		hasAmount      bool
		countsPresence bool
	}{
		{"visit", "Визит", false, true},
		{"purchase", "Покупка", true, true},
		{"booking", "Бронь", true, false},
		{"top_up", "Пополнение", true, false},
		{"usage", "Списание", true, false},
		{"reversal", "Возврат/отмена", true, false},
	}
	for _, d := range defaults {
		var exists int
		err := r.db.QueryRowContext(ctx, `SELECT 1 FROM event_type WHERE code = $1`, d.code).Scan(&exists)
		if err == nil {
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO event_type (id, code, name, has_amount, counts_as_presence, is_active)
			VALUES ($1, $2, $3, $4, $5, TRUE)`,
			idGen.NewID(ctx), d.code, d.name, d.hasAmount, d.countsPresence,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
