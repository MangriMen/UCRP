package domain

import (
	"context"
	"time"
)

type AuditLog struct {
	ID          string
	ActorUserID string
	ActorRole   string
	Action      string
	EntityType  string
	EntityID    string
	Details     string
	RequestID   string
	CreatedAt   time.Time
}

type AuditRepo interface {
	List(ctx context.Context, entityType, entityID, actor string, limit, offset int) ([]AuditLog, error)
	Insert(ctx context.Context, a *AuditLog) error
}

type Campaign struct {
	ID              string
	Name            string
	TenantID        string
	SegmentID       string
	AttributionDays int
	Status          string
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CampaignRepo interface {
	List(ctx context.Context, limit, offset int) ([]Campaign, error)
	GetByID(ctx context.Context, id string) (*Campaign, error)
	Insert(ctx context.Context, c *Campaign) error
	Update(ctx context.Context, c *Campaign) error
	Delete(ctx context.Context, id string) error
	HasExports(ctx context.Context, campaignID string) (bool, error)
}
