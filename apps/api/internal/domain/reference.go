package domain

import (
	"context"
	"time"
)

type Mall struct {
	ID        string
	Code      string
	Name      string
	IsActive  bool
	FloorMin  int
	FloorMax  int
	HasZero   bool
	CreatedAt time.Time
}

const maxFloorSpan = 200

func (m Mall) Floors() []int {
	out := make([]int, 0, m.FloorMax-m.FloorMin+1)
	for floor := m.FloorMin; floor <= m.FloorMax; floor++ {
		if floor == 0 && !m.HasZero {
			continue
		}
		out = append(out, floor)
	}
	return out
}

func (m Mall) ValidFloorRange() bool {
	return m.FloorMin <= m.FloorMax && m.FloorMax-m.FloorMin < maxFloorSpan
}

func (m Mall) HasFloor(floor int) bool {
	if floor == 0 && !m.HasZero {
		return false
	}
	return floor >= m.FloorMin && floor <= m.FloorMax
}

type Tenant struct {
	ID         string
	Code       string
	Name       string
	MallID     string
	CategoryID string
	Kind       string
	Floor      *int
	Priority   int
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type TenantCategory struct {
	ID          string
	Code        string
	Name        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type CategoryRelation struct {
	CategoryID          string
	RelatedCategoryID   string
	RelatedCategoryName string
	RelationType        string
	WeightPct           int
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

type TenantFilter struct {
	Query      string
	MallID     string
	CategoryID string
	Floor      *int
	IsActive   *bool
	PeriodFrom *time.Time
	PeriodTo   *time.Time
	Sort       string
	KThreshold int
}

type TenantStats struct {
	GuestsCount         *int
	VisitsCount         *int
	NewGuestsCount      *int
	SourcesCount        int
	ProblemSourcesCount int
}

type TenantListEntry struct {
	Tenant
	CategoryName string
	Stats        TenantStats
}

type TenantRepo interface {
	Search(ctx context.Context, filter TenantFilter, limit, offset int) ([]TenantListEntry, int, error)
	GetByID(ctx context.Context, id string) (*Tenant, error)
	Names(ctx context.Context, ids []string) (map[string]string, error)
	Insert(ctx context.Context, t *Tenant) error
	Update(ctx context.Context, t *Tenant) error
	SoftDelete(ctx context.Context, id string, now time.Time) error
}

type MallRepo interface {
	List(ctx context.Context) ([]Mall, error)
	GetByID(ctx context.Context, id string) (*Mall, error)
	Insert(ctx context.Context, m *Mall) error
	Update(ctx context.Context, m *Mall) error
	Delete(ctx context.Context, id string) error
}

type TenantCategoryRepo interface {
	List(ctx context.Context, includeInactive bool) ([]TenantCategory, error)
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
