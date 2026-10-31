package domain

import (
	"context"
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
	Was        []string `json:"was"`
	WasNot     []string `json:"was_not"`
	PeriodDays int      `json:"period_days"`
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
	ListMemberContacts(ctx context.Context, versionID string) ([]GuestContact, error)
}

type SegmentEvaluator interface {
	Preview(ctx context.Context, rule SegmentRule) (size int, examples []string, err error)
	Evaluate(ctx context.Context, rule SegmentRule) ([]string, error)
}
