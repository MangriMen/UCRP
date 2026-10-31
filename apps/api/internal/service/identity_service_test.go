package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type fakeQueue struct {
	jobs []domain.Job
}

func (f *fakeQueue) Enqueue(ctx context.Context, job domain.Job) error {
	f.jobs = append(f.jobs, job)
	return nil
}

func (f *fakeQueue) Take(ctx context.Context, jobTypes []string, workerID string) (*domain.Job, error) {
	return nil, nil
}

func (f *fakeQueue) Complete(ctx context.Context, id string) error { return nil }

func (f *fakeQueue) Fail(ctx context.Context, id string, errMsg string, retryAfter time.Time) error {
	return nil
}

type fakeMergeRepo struct {
	confirm *domain.MergeCandidate
	revert  *domain.MergeCandidate
}

func (f *fakeMergeRepo) List(ctx context.Context, resolution domain.MergeResolution, limit, offset int) ([]domain.MergeCandidate, error) {
	return nil, nil
}

func (f *fakeMergeRepo) GetByID(ctx context.Context, id string) (*domain.MergeCandidate, error) {
	return nil, domain.ErrNotFound
}

func (f *fakeMergeRepo) CreatePending(ctx context.Context, guestA, guestB, matchType, reasonCode, reasonText string, score int, now time.Time) (*domain.MergeCandidate, error) {
	return nil, nil
}

func (f *fakeMergeRepo) AutoMerge(ctx context.Context, guestA, guestB, matchType, reasonCode, reasonText string, score int, now time.Time) (*domain.MergeCandidate, error) {
	return nil, nil
}

func (f *fakeMergeRepo) Confirm(ctx context.Context, id, actorUserID string, now time.Time) (*domain.MergeCandidate, error) {
	return f.confirm, nil
}

func (f *fakeMergeRepo) Reject(ctx context.Context, id, actorUserID string, now time.Time) (*domain.MergeCandidate, error) {
	return f.confirm, nil
}

func (f *fakeMergeRepo) Revert(ctx context.Context, id, actorUserID string, now time.Time) (*domain.MergeCandidate, error) {
	return f.revert, nil
}

type fakeManualRepo struct {
	result *domain.MergeCandidate
}

func (f *fakeManualRepo) ManualMerge(ctx context.Context, guestA, guestB, actorUserID string, now time.Time) (*domain.MergeCandidate, error) {
	return f.result, nil
}

type fakeBlocks struct{}

func (f *fakeBlocks) HasBlock(ctx context.Context, guestA, guestB string) (bool, error) {
	return false, nil
}

func (f *fakeBlocks) Insert(ctx context.Context, b *domain.UnmergeBlock) error { return nil }

func (f *fakeBlocks) List(ctx context.Context, limit, offset int) ([]domain.UnmergeBlock, error) {
	return nil, nil
}

func (f *fakeBlocks) GetByID(ctx context.Context, id string) (*domain.UnmergeBlock, error) {
	return nil, domain.ErrNotFound
}

func (f *fakeBlocks) Delete(ctx context.Context, id string) error { return nil }

func TestConfirmEnqueuesRecommendRefresh(t *testing.T) {
	q := &fakeQueue{}
	svc := NewIdentityService(&fakeMergeRepo{confirm: &domain.MergeCandidate{
		GuestAID: "win",
		GuestBID: "lose",
	}}, &fakeBlocks{}, &fakeManualRepo{}, q)
	if _, err := svc.Confirm(context.Background(), "c1", "u1"); err != nil {
		t.Fatal(err)
	}
	if len(q.jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(q.jobs))
	}
	for i, job := range q.jobs {
		if job.JobType != "recommend.refresh" {
			t.Fatalf("job %d type: %s", i, job.JobType)
		}
		var payload map[string]string
		if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
			t.Fatal(err)
		}
		if payload["guest_id"] == "" {
			t.Fatalf("job %d missing guest_id", i)
		}
	}
}

func TestManualMergeEnqueuesRecommendRefresh(t *testing.T) {
	q := &fakeQueue{}
	svc := NewIdentityService(&fakeMergeRepo{}, &fakeBlocks{}, &fakeManualRepo{result: &domain.MergeCandidate{
		WinnerGuestID: "target",
		MergedGuestID: "source",
	}}, q)
	if _, err := svc.Merge(context.Background(), "source", "target", "u1"); err != nil {
		t.Fatal(err)
	}
	if len(q.jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(q.jobs))
	}
}
