package domain

import (
	"context"
	"time"
)

type JobStatus string

const (
	JobQueued  JobStatus = "queued"
	JobRunning JobStatus = "running"
	JobDone    JobStatus = "done"
	JobFailed  JobStatus = "failed"
)

type Job struct {
	ID        string
	JobType   string
	Payload   string
	DedupKey  string
	Status    JobStatus
	Attempts  int
	RunAfter  time.Time
	LockedBy  string
	LockedAt  *time.Time
	LastError string
	CreatedAt time.Time
}

type Queue interface {
	Enqueue(ctx context.Context, job Job) error
	Take(ctx context.Context, jobTypes []string, workerID string) (*Job, error)
	Complete(ctx context.Context, id string) error
	Fail(ctx context.Context, id string, errMsg string, retryAfter time.Time) error
}
