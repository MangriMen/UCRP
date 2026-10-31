package domain

import (
	"context"
	"time"
)

type TenantGuest struct {
	ID          string
	FirstSeenAt time.Time
}

type TenantGuestIdentifier struct {
	ID              string
	Type            string
	ValueNormalized string
	FirstSeenAt     time.Time
}

type TenantGuestEvent struct {
	ID          string
	EventTypeID string
	TenantID    string
	ProjectID   string
	ExternalID  string
	OccurredAt  time.Time
	AmountMinor *int64
	Currency    string
}

type TenantGuestRepo interface {
	Get(ctx context.Context, tenantID, guestID string, since time.Time) (*TenantGuest, error)
	Identifiers(ctx context.Context, tenantID, guestID string, since time.Time, limit int) ([]TenantGuestIdentifier, error)
	Events(ctx context.Context, tenantID, guestID string, since time.Time, limit, offset int) ([]TenantGuestEvent, error)
}
