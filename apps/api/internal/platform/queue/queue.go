package queue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type DBQueue struct {
	db    *sql.DB
	clock domain.Clock
	idGen domain.IDGen
}

func NewDBQueue(db *sql.DB, clock domain.Clock, idGen domain.IDGen) *DBQueue {
	return &DBQueue{db: db, clock: clock, idGen: idGen}
}

func (q *DBQueue) Enqueue(ctx context.Context, job domain.Job) error {
	if job.ID == "" {
		job.ID = q.idGen.NewID(ctx)
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = q.clock.Now()
	}
	if job.RunAfter.IsZero() {
		job.RunAfter = job.CreatedAt
	}
	if job.Status == "" {
		job.Status = domain.JobQueued
	}
	if job.DedupKey != "" {
		var exists int
		err := q.db.QueryRowContext(ctx, `
			SELECT 1 FROM job_queue
			WHERE dedup_key = $1 AND status IN ('queued','running')
			LIMIT 1`, job.DedupKey).Scan(&exists)
		if err == nil {
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
	}
	var dedup any
	if job.DedupKey != "" {
		dedup = job.DedupKey
	}
	_, err := q.db.ExecContext(ctx, `
		INSERT INTO job_queue (id, job_type, payload, dedup_key, status, attempts, run_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		job.ID, job.JobType, job.Payload, dedup, string(job.Status), job.Attempts, job.RunAfter, job.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("enqueue: %w", err)
	}
	return nil
}

func (q *DBQueue) Take(ctx context.Context, jobTypes []string, workerID string) (*domain.Job, error) {
	if len(jobTypes) == 0 {
		return nil, nil
	}
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	now := q.clock.Now()
	args := make([]any, 0, len(jobTypes)+3)
	args = append(args, now, string(domain.JobRunning), workerID)
	for _, t := range jobTypes {
		args = append(args, t)
	}
	query := `
		WITH next AS (
			SELECT id FROM job_queue
			WHERE status = 'queued' AND run_after <= $1 AND job_type IN (` + numberedPlaceholders(4, len(jobTypes)) + `)
			ORDER BY run_after
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE job_queue j
		SET status = $2, locked_by = $3, locked_at = $1, attempts = j.attempts + 1
		FROM next
		WHERE j.id = next.id
		RETURNING j.id, j.job_type, j.payload, j.dedup_key, j.status, j.attempts, j.run_after, j.created_at`

	var (
		job    domain.Job
		status string
		dedup  sql.NullString
	)
	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&job.ID, &job.JobType, &job.Payload, &dedup, &status, &job.Attempts, &job.RunAfter, &job.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	job.DedupKey = dedup.String
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatus(status)
	job.LockedBy = workerID
	job.Attempts++
	return &job, nil
}

func (q *DBQueue) Complete(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, `
		UPDATE job_queue SET status = 'done', locked_by = NULL, locked_at = NULL, last_error = NULL
		WHERE id = $1`, id)
	return err
}

func (q *DBQueue) Fail(ctx context.Context, id string, errMsg string, retryAfter time.Time) error {
	if retryAfter.IsZero() {
		_, err := q.db.ExecContext(ctx, `
			UPDATE job_queue SET status = 'failed', locked_by = NULL, locked_at = NULL, last_error = $2
			WHERE id = $1`, id, errMsg)
		return err
	}
	_, err := q.db.ExecContext(ctx, `
		UPDATE job_queue SET status = 'queued', run_after = $2, locked_by = NULL, locked_at = NULL, last_error = $3
		WHERE id = $1`, id, retryAfter, errMsg)
	return err
}

func numberedPlaceholders(start, n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(parts, ",")
}
