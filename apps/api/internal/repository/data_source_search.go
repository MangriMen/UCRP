package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MangriMen/UCRP/internal/domain"
)

const (
	defaultDataSourcePageSize = 50
	maxDataSourcePageSize     = 200
	noImportStatus            = "none"
)

var dataSourceSorts = map[string]string{
	"name":          "d.name ASC",
	"tenantName":    "tenant_name ASC NULLS FIRST",
	"typeName":      "t.name ASC",
	"lastImportAt":  "d.last_import_at ASC NULLS LAST",
	"createdAt":     "d.created_at ASC",
	"-name":         "d.name DESC",
	"-tenantName":   "tenant_name DESC NULLS LAST",
	"-typeName":     "t.name DESC",
	"-lastImportAt": "d.last_import_at DESC NULLS LAST",
	"-createdAt":    "d.created_at DESC",
}

func (r *DataSourceSQL) Search(ctx context.Context, filter domain.DataSourceFilter, limit, offset int) ([]domain.DataSourceListEntry, int, error) {
	if limit <= 0 {
		limit = defaultDataSourcePageSize
	}
	if limit > maxDataSourcePageSize {
		limit = maxDataSourcePageSize
	}
	if offset < 0 {
		offset = 0
	}
	order, ok := dataSourceSorts[filter.Sort]
	if !ok {
		order = dataSourceSorts["-createdAt"]
	}

	var args []any
	where := []string{"TRUE"}
	add := func(condition string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(condition, len(args)))
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		add("(d.name ILIKE '%%' || $%[1]d || '%%' OR tn.name ILIKE '%%' || $%[1]d || '%%')", escapeLike(q))
	}
	if filter.TenantID != "" {
		add("d.tenant_id = $%d", filter.TenantID)
	}
	if filter.TypeID != "" {
		add("d.type_id = $%d", filter.TypeID)
	}
	if filter.IsActive != nil {
		add("d.is_active = $%d", *filter.IsActive)
	}
	switch filter.ImportStatus {
	case "":
	case noImportStatus:
		where = append(where, "lb.status IS NULL")
	default:
		add("lb.status = $%d", filter.ImportStatus)
	}
	whereSQL := strings.Join(where, " AND ")

	const from = `
		FROM data_source d
		JOIN data_source_type t ON t.id = d.type_id
		LEFT JOIN tenant tn ON tn.id = d.tenant_id
		LEFT JOIN LATERAL (
			SELECT b.id, b.status, b.created_at, b.finished_at, b.rejected_rows
			FROM import_batch b
			WHERE b.data_source_id = d.id
			ORDER BY b.created_at DESC
			LIMIT 1
		) lb ON TRUE`

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+from+` WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT d.id, d.tenant_id, d.type_id, t.code, d.name, COALESCE(d.config, ''), COALESCE(d.field_mapping, ''),
			COALESCE(d.schedule_cron, ''), d.is_active, t.supports_file, t.supports_pull, d.last_import_at,
			d.created_at, d.updated_at, tn.name AS tenant_name,
			lb.id, lb.status, lb.created_at, lb.finished_at, lb.rejected_rows`+from+`
		WHERE %s
		ORDER BY %s, d.id ASC
		LIMIT $%d OFFSET $%d`, whereSQL, order, len(args)-1, len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.DataSourceListEntry, 0)
	for rows.Next() {
		var (
			entry        domain.DataSourceListEntry
			tenantID     sql.NullString
			tenantName   sql.NullString
			lastImportAt sql.NullTime
			batchID      sql.NullString
			status       sql.NullString
			batchCreated sql.NullTime
			batchDone    sql.NullTime
			rejected     sql.NullInt64
		)
		if err := rows.Scan(
			&entry.ID, &tenantID, &entry.TypeID, &entry.TypeCode, &entry.Name, &entry.Config, &entry.FieldMapping,
			&entry.ScheduleCron, &entry.IsActive, &entry.SupportsFile, &entry.SupportsPull, &lastImportAt,
			&entry.CreatedAt, &entry.UpdatedAt, &tenantName,
			&batchID, &status, &batchCreated, &batchDone, &rejected,
		); err != nil {
			return nil, 0, err
		}
		entry.TenantID = tenantID.String
		entry.TenantName = tenantName.String
		if lastImportAt.Valid {
			entry.LastImportAt = &lastImportAt.Time
		}
		if batchID.Valid {
			last := &domain.DataSourceLastImport{
				BatchID:      batchID.String,
				Status:       domain.ImportStatus(status.String),
				CreatedAt:    batchCreated.Time,
				RejectedRows: int(rejected.Int64),
			}
			if batchDone.Valid {
				last.FinishedAt = &batchDone.Time
			}
			entry.LastImport = last
		}
		out = append(out, entry)
	}
	return out, total, rows.Err()
}
