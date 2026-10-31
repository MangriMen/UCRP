package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type expiryTestRepository struct {
	now   time.Time
	limit int
	count int
	err   error
}

func (r *expiryTestRepository) ExpireDue(_ context.Context, now time.Time, limit int) (int, error) {
	r.now = now
	r.limit = limit
	return r.count, r.err
}

func TestGrantExpiryRunOnce(t *testing.T) {
	instant := time.Date(2026, 11, 1, 12, 0, 0, 0, time.FixedZone("test", 7*60*60))
	repo := &expiryTestRepository{count: 2}
	svc := NewGrantExpiry(repo, grantTestClock{now: instant})
	count, err := svc.RunOnce(context.Background())
	if err != nil || count != 2 {
		t.Fatalf("RunOnce: count = %d, error = %v", count, err)
	}
	if !repo.now.Equal(instant) || repo.now.Location() != time.UTC || repo.limit != 100 {
		t.Fatalf("RunOnce: now = %v, limit = %d", repo.now, repo.limit)
	}
	failure := errors.New("storage unavailable")
	repo.err = failure
	_, err = svc.RunOnce(context.Background())
	if !errors.Is(err, failure) {
		t.Fatalf("RunOnce: error = %v, want %v", err, failure)
	}
}
