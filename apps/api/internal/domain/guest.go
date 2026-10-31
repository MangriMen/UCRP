package domain

import (
	"context"
	"fmt"
	"time"
)

type GuestStatus string

const (
	GuestActive  GuestStatus = "active"
	GuestMerged  GuestStatus = "merged"
	GuestDeleted GuestStatus = "deleted"
)

func (s *GuestStatus) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*s = GuestStatus(v)
		return nil
	case []byte:
		*s = GuestStatus(v)
		return nil
	default:
		return fmt.Errorf("invalid guest status %T", src)
	}
}

type Guest struct {
	ID           string
	DisplayName  string
	Status       GuestStatus
	MergedIntoID string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type GuestIdentifier struct {
	ID              string
	GuestID         string
	OriginGuestID   string
	IDType          string
	ValueNormalized string
	DataSourceID    string
	FirstSeenAt     time.Time
}

type Event struct {
	ID            string
	GuestID       string
	OriginGuestID string
	TenantID      string
	ProjectID     string
	EventTypeID   string
	DataSourceID  string
	ImportBatchID string
	ExternalID    string
	OccurredAt    time.Time
	AmountMinor   *int64
	Currency      string
	MatchKeys     string
	RawPayload    string
	CreatedAt     time.Time
}

type GuestRepo interface {
	GetByID(ctx context.Context, id string) (*Guest, error)
	Search(ctx context.Context, q string, limit, offset int) ([]Guest, error)
	Insert(ctx context.Context, g *Guest) error
	TenantCount(ctx context.Context, guestID string) (int, error)
	TenantNames(ctx context.Context, guestID string) ([]string, error)
	MarkMerged(ctx context.Context, loserID, winnerID string) error
	MarkActive(ctx context.Context, guestID string) error
}

type GuestIdentifierRepo interface {
	ListByGuest(ctx context.Context, guestID string) ([]GuestIdentifier, error)
	FindByValue(ctx context.Context, idType, value string) ([]GuestIdentifier, error)
	Insert(ctx context.Context, gi *GuestIdentifier) error
	MoveToGuest(ctx context.Context, fromGuestID, toGuestID string) error
	RestoreOrigin(ctx context.Context, originGuestID string) error
}

type EventRepo interface {
	ListByGuest(ctx context.Context, guestID string, limit, offset int) ([]Event, error)
	ListUnmatchedByBatch(ctx context.Context, batchID string, limit int) ([]Event, error)
	Insert(ctx context.Context, e *Event) error
	InsertIgnoreDuplicate(ctx context.Context, e *Event) (bool, error)
	AttachGuest(ctx context.Context, eventID, guestID string) error
	MoveToGuest(ctx context.Context, fromGuestID, toGuestID string) error
	RestoreOrigin(ctx context.Context, originGuestID string) error
}
