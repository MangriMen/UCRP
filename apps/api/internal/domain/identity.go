package domain

import (
	"context"
	"time"
)

type MergeResolution string

const (
	MergePending    MergeResolution = "pending"
	MergeAutoMerged MergeResolution = "auto_merged"
	MergeConfirmed  MergeResolution = "confirmed"
	MergeRejected   MergeResolution = "rejected"
	MergeReverted   MergeResolution = "reverted"
)

type MergeCandidate struct {
	ID            string
	GuestAID      string
	GuestBID      string
	MatchType     string
	ReasonCode    string
	ReasonText    string
	Score         int
	Resolution    MergeResolution
	WinnerGuestID string
	MergedGuestID string
	DecidedBy     string
	DecidedAt     *time.Time
	CreatedAt     time.Time
}

type UnmergeBlock struct {
	ID          string
	GuestAID    string
	GuestBID    string
	Reason      string
	CandidateID string
	CreatedBy   string
	CreatedAt   time.Time
}

type MergeCandidateRepo interface {
	GetByID(ctx context.Context, id string) (*MergeCandidate, error)
	List(ctx context.Context, resolution MergeResolution, limit, offset int) ([]MergeCandidate, error)
	Confirm(ctx context.Context, id, actorUserID string, now time.Time) (*MergeCandidate, error)
	Reject(ctx context.Context, id, actorUserID string, now time.Time) (*MergeCandidate, error)
	Revert(ctx context.Context, id, actorUserID string, now time.Time) (*MergeCandidate, error)
}

type UnmergeBlockRepo interface {
	HasBlock(ctx context.Context, guestA, guestB string) (bool, error)
	Insert(ctx context.Context, b *UnmergeBlock) error
}
