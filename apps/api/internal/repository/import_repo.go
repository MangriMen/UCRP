package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/google/uuid"
)

type ImportBatchSQL struct {
	db *sql.DB
}

func NewImportBatchSQL(db *sql.DB) *ImportBatchSQL {
	return &ImportBatchSQL{db: db}
}

func (r *ImportBatchSQL) CreateQueued(ctx context.Context, batch *domain.ImportBatch, jobType, dedupKey string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO import_batch (
			id, data_source_id, mode, file_name, file_hash, file_size_bytes, status,
			total_rows, accepted_rows, rejected_rows, duplicate_rows, error_message,
			created_by, created_at, started_at, finished_at, window_from, window_to
		) VALUES ($1,$2,$3,$4,$5,$6,$7,0,0,0,0,NULL,$8,$9,NULL,NULL,$10,$11)`,
		batch.ID, batch.DataSourceID, string(batch.Mode), nullIfEmpty(batch.FileName),
		nullIfEmpty(batch.FileHash), nullInt(batch.FileSizeBytes), string(batch.Status),
		nullIfEmpty(batch.CreatedBy), batch.CreatedAt, batch.WindowFrom, batch.WindowTo,
	)
	if err != nil {
		if isDuplicate(err) {
			return domain.ErrConflict
		}
		return err
	}
	if jobType != "" {
		payload, err := json.Marshal(map[string]string{"batch_id": batch.ID, "data_source_id": batch.DataSourceID})
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO job_queue (id, job_type, payload, dedup_key, status, attempts, run_after, created_at)
			VALUES ($1, $2, $3, $4, 'queued', 0, $5, $5)
			ON CONFLICT DO NOTHING`,
			uuid.NewString(), jobType, string(payload), dedupKey, batch.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ImportBatchSQL) FindByHash(ctx context.Context, dataSourceID, fileHash string) (*domain.ImportBatch, error) {
	return r.one(ctx, batchSelect+` WHERE data_source_id = $1 AND file_hash = $2`, dataSourceID, fileHash)
}

func (r *ImportBatchSQL) GetByID(ctx context.Context, id string) (*domain.ImportBatch, error) {
	return r.one(ctx, batchSelect+` WHERE id = $1`, id)
}

func (r *ImportBatchSQL) List(ctx context.Context, dataSourceID, status string, limit, offset int) ([]domain.ImportBatch, int, error) {
	limit, offset = clampPage(limit, offset)
	where := ` WHERE TRUE`
	args := make([]any, 0, 4)
	if dataSourceID != "" {
		args = append(args, dataSourceID)
		where += ` AND data_source_id = $` + strconv.Itoa(len(args))
	}
	if status != "" {
		args = append(args, status)
		where += ` AND status = $` + strconv.Itoa(len(args))
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM import_batch`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	query := batchSelect + where + ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.ImportBatch, 0)
	for rows.Next() {
		b, err := scanBatch(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *b)
	}
	return out, total, rows.Err()
}

func (r *ImportBatchSQL) MarkRunning(ctx context.Context, id string, startedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_batch SET status = 'running', started_at = $2 WHERE id = $1`, id, startedAt)
	return err
}

func (r *ImportBatchSQL) MarkFinished(ctx context.Context, id string, status domain.ImportStatus, total, accepted, rejected, duplicate int, errMsg string, finishedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_batch
		SET status = $2, total_rows = $3, accepted_rows = $4, rejected_rows = $5, duplicate_rows = $6,
			error_message = $7, finished_at = $8
		WHERE id = $1`,
		id, string(status), total, accepted, rejected, duplicate, nullIfEmpty(errMsg), finishedAt,
	)
	return err
}

func (r *ImportBatchSQL) HasActivePull(ctx context.Context, dataSourceID string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1 FROM import_batch
		WHERE data_source_id = $1 AND mode = 'pull' AND status IN ('queued','running')
		LIMIT 1`, dataSourceID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ImportBatchSQL) one(ctx context.Context, query string, args ...any) (*domain.ImportBatch, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	b, err := scanBatch(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return b, err
}

const batchSelect = `
	SELECT id, data_source_id, COALESCE((SELECT d.name FROM data_source d WHERE d.id = import_batch.data_source_id), ''), mode, COALESCE(file_name, ''), COALESCE(file_hash, ''), COALESCE(file_size_bytes, 0),
		status, total_rows, accepted_rows, rejected_rows, duplicate_rows, COALESCE(error_message, ''),
		COALESCE(created_by, ''), created_at, started_at, finished_at, window_from, window_to
	FROM import_batch`

func scanBatch(s scanner) (*domain.ImportBatch, error) {
	var (
		b        domain.ImportBatch
		started  sql.NullTime
		finished sql.NullTime
		winFrom  sql.NullTime
		winTo    sql.NullTime
	)
	err := s.Scan(
		&b.ID, &b.DataSourceID, &b.DataSourceName, &b.Mode, &b.FileName, &b.FileHash, &b.FileSizeBytes,
		&b.Status, &b.TotalRows, &b.AcceptedRows, &b.RejectedRows, &b.DuplicateRows, &b.ErrorMessage,
		&b.CreatedBy, &b.CreatedAt, &started, &finished, &winFrom, &winTo,
	)
	if err != nil {
		return nil, err
	}
	if started.Valid {
		b.StartedAt = &started.Time
	}
	if finished.Valid {
		b.FinishedAt = &finished.Time
	}
	if winFrom.Valid {
		b.WindowFrom = &winFrom.Time
	}
	if winTo.Valid {
		b.WindowTo = &winTo.Time
	}
	if b.Mode == domain.ImportModeFile {
		b.StorageKey = "imports/" + b.ID + ".csv"
	}
	return &b, nil
}

type RejectedRowSQL struct {
	db *sql.DB
}

func NewRejectedRowSQL(db *sql.DB) *RejectedRowSQL {
	return &RejectedRowSQL{db: db}
}

func (r *RejectedRowSQL) ListByBatch(ctx context.Context, batchID string, limit, offset int) ([]domain.RejectedRow, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, import_batch_id, row_number, error_code, error_message, COALESCE(raw_line, ''), created_at
		FROM rejected_row WHERE import_batch_id = $1
		ORDER BY row_number
		LIMIT $2 OFFSET $3`, batchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.RejectedRow, 0)
	for rows.Next() {
		var rr domain.RejectedRow
		if err := rows.Scan(&rr.ID, &rr.ImportBatchID, &rr.RowNumber, &rr.ErrorCode, &rr.ErrorMessage, &rr.RawLine, &rr.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rr)
	}
	return out, rows.Err()
}

func (r *RejectedRowSQL) CountByBatch(ctx context.Context, batchID string) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM rejected_row WHERE import_batch_id = $1`, batchID).Scan(&total)
	return total, err
}

func (r *RejectedRowSQL) InsertBatch(ctx context.Context, rows []domain.RejectedRow) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO rejected_row (id, import_batch_id, row_number, error_code, error_message, raw_line, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (import_batch_id, row_number) DO NOTHING`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()
	for _, row := range rows {
		if _, err := stmt.ExecContext(ctx, row.ID, row.ImportBatchID, row.RowNumber, row.ErrorCode, row.ErrorMessage, nullIfEmpty(row.RawLine), row.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func nullInt(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}

func isDuplicate(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint"))
}
