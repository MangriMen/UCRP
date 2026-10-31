package domain

import (
	"context"
	"time"
)

type Tenant struct {
	ID         string
	Code       string
	Name       string
	CategoryID string
	Kind       string
	Floor      int
	Priority   int
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type TenantCategory struct {
	ID        string
	Code      string
	Name      string
	IsActive  bool
	CreatedAt time.Time
}

type CategoryRelation struct {
	CategoryID        string
	RelatedCategoryID string
	RelationType      string
	WeightPct         int
}

type Project struct {
	ID       string
	TenantID string
	Code     string
	Name     string
	IsActive bool
}

type EventType struct {
	ID               string
	Code             string
	Name             string
	HasAmount        bool
	CountsAsPresence bool
	IsActive         bool
}

type TenantRepo interface {
	List(ctx context.Context, limit, offset int) ([]Tenant, error)
	GetByID(ctx context.Context, id string) (*Tenant, error)
	Insert(ctx context.Context, t *Tenant) error
	Update(ctx context.Context, t *Tenant) error
	SoftDelete(ctx context.Context, id string, now time.Time) error
}

type TenantCategoryRepo interface {
	List(ctx context.Context) ([]TenantCategory, error)
	GetByID(ctx context.Context, id string) (*TenantCategory, error)
	Insert(ctx context.Context, c *TenantCategory) error
	Update(ctx context.Context, c *TenantCategory) error
	Delete(ctx context.Context, id string) error
	ListRelations(ctx context.Context, categoryID string) ([]CategoryRelation, error)
	ReplaceRelations(ctx context.Context, categoryID string, rels []CategoryRelation) error
}

type ProjectRepo interface {
	List(ctx context.Context, tenantID string, limit, offset int) ([]Project, error)
	GetByID(ctx context.Context, id string) (*Project, error)
	Insert(ctx context.Context, p *Project) error
	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id string) error
}

type EventTypeRepo interface {
	List(ctx context.Context) ([]EventType, error)
	GetByID(ctx context.Context, id string) (*EventType, error)
	Insert(ctx context.Context, e *EventType) error
	Update(ctx context.Context, e *EventType) error
	Delete(ctx context.Context, id string) error
}
