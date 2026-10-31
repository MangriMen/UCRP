package domain

import (
	"context"
	"encoding/json"
	"time"
)

type SegmentStatus string

const (
	SegmentActive   SegmentStatus = "active"
	SegmentArchived SegmentStatus = "archived"
)

type SegmentVersionStatus string

const (
	VersionPending   SegmentVersionStatus = "pending"
	VersionComputing SegmentVersionStatus = "computing"
	VersionReady     SegmentVersionStatus = "ready"
	VersionFailed    SegmentVersionStatus = "failed"
)

type Segment struct {
	ID               string
	Name             string
	Description      string
	TemplateCode     string
	IsTemplate       bool
	CurrentVersionNo int
	Status           SegmentStatus
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SegmentVersion struct {
	ID          string
	SegmentID   string
	VersionNo   int
	Rule        string
	Status      SegmentVersionStatus
	MemberCount *int
	ComputedAt  *time.Time
	CreatedBy   string
	CreatedAt   time.Time
}

type SegmentRule struct {
	VisitedTenantIDs  []string `json:"visitedTenantIds"`
	ExcludedTenantIDs []string `json:"excludedTenantIds"`
	PeriodDays        int      `json:"periodDays"`
	RequireContact    string   `json:"requireContact,omitempty"`
}

func (r *SegmentRule) UnmarshalJSON(data []byte) error {
	var raw struct {
		VisitedTenantIDs       []string `json:"visitedTenantIds"`
		ExcludedTenantIDs      []string `json:"excludedTenantIds"`
		ExcludedTenantIDsSnake []string `json:"excluded_tenant_ids"`
		PeriodDays             int      `json:"periodDays"`
		PeriodDaysSNAK         int      `json:"period_days"`
		RequireContact         string   `json:"requireContact"`
		RequireSnake           string   `json:"require_contact"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.VisitedTenantIDs = raw.VisitedTenantIDs
	r.ExcludedTenantIDs = raw.ExcludedTenantIDs
	if len(r.ExcludedTenantIDs) == 0 {
		r.ExcludedTenantIDs = raw.ExcludedTenantIDsSnake
	}
	r.PeriodDays = raw.PeriodDays
	if r.PeriodDays == 0 {
		r.PeriodDays = raw.PeriodDaysSNAK
	}
	r.RequireContact = raw.RequireContact
	if r.RequireContact == "" {
		r.RequireContact = raw.RequireSnake
	}
	return nil
}

type GuestContact struct {
	GuestID      string
	IdentifierID string
	Contact      string
	IDType       string
}

type SegmentRepo interface {
	Insert(ctx context.Context, s *Segment, v *SegmentVersion) error
	GetByID(ctx context.Context, id string) (*Segment, *SegmentVersion, error)
	List(ctx context.Context, limit, offset int) ([]Segment, error)
	ListTemplates(ctx context.Context) ([]Segment, error)
	ListVersions(ctx context.Context, segmentID string) ([]SegmentVersion, error)
	GetVersion(ctx context.Context, segmentID string, versionNo int) (*Segment, *SegmentVersion, error)
	InsertVersion(ctx context.Context, v *SegmentVersion) error
	Archive(ctx context.Context, id string, now time.Time) error
	MarkComputing(ctx context.Context, versionID string) error
	ReplaceMembers(ctx context.Context, versionID string, guestIDs []string, computedAt time.Time) error
	ListMembers(ctx context.Context, versionID string, limit, offset int) ([]string, error)
	CountMembers(ctx context.Context, versionID string) (int, error)
	CurrentVersions(ctx context.Context, segmentIDs []string) (map[string]*SegmentVersion, error)
	ListMemberContacts(ctx context.Context, versionID string) ([]GuestContact, error)
}

type SegmentEvaluator interface {
	Preview(ctx context.Context, rule SegmentRule) (size int, examples []string, err error)
	Evaluate(ctx context.Context, rule SegmentRule) ([]string, error)
}
