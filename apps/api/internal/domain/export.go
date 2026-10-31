package domain

import (
	"context"
	"time"
)

type ExportChannel string

const (
	ChannelSMS   ExportChannel = "sms"
	ChannelPush  ExportChannel = "push"
	ChannelEmail ExportChannel = "email"
)

type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatAPI  ExportFormat = "api"
	FormatXLSX ExportFormat = "xlsx"
)

type ExportStatus string

const (
	ExportQueued   ExportStatus = "queued"
	ExportBuilding ExportStatus = "building"
	ExportReady    ExportStatus = "ready"
	ExportFailed   ExportStatus = "failed"
)

type ExportJob struct {
	ID                  string
	SegmentID           string
	SegmentVersionID    string
	CampaignID          string
	Channel             ExportChannel
	Format              ExportFormat
	ControlPercent      int
	ControlSeed         int64
	IdempotencyKey      string
	Status              ExportStatus
	TotalInSegment      int
	NoContactCount      int
	BlockedByLimitCount int
	ControlCount        int
	ExportedCount       int
	CreatedBy           string
	CreatedAt           time.Time
	FinishedAt          *time.Time
}

type ExportMember struct {
	ExportID     string
	GuestID      string
	IdentifierID string
	IsControl    bool
	Contact      string
}

type ContactLimit struct {
	ID          string
	Channel     string
	MaxContacts int
	PeriodDays  int
	IsActive    bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ContactLog struct {
	ID          string
	GuestID     string
	ExportID    string
	Channel     string
	ContactedAt time.Time
}

type ExportRepo interface {
	FindByKey(ctx context.Context, userID, key string) (*ExportJob, error)
	GetByID(ctx context.Context, id string) (*ExportJob, error)
	ListMembers(ctx context.Context, exportID string) ([]ExportMember, error)
	Insert(ctx context.Context, job *ExportJob, members []ExportMember) error
	List(ctx context.Context, limit, offset int) ([]ExportJob, error)
}

type ContactLimitRepo interface {
	List(ctx context.Context, limit, offset int) ([]ContactLimit, error)
	ListActive(ctx context.Context) ([]ContactLimit, error)
	GetByID(ctx context.Context, id string) (*ContactLimit, error)
	Insert(ctx context.Context, c *ContactLimit) error
	Update(ctx context.Context, c *ContactLimit) error
	Delete(ctx context.Context, id string) error
}

type ContactLogRepo interface {
	ListByGuest(ctx context.Context, guestID string, limit, offset int) ([]ContactLog, error)
	CountSince(ctx context.Context, guestIDs []string, since time.Time) (map[string]int, error)
	CountSinceChannel(ctx context.Context, guestIDs []string, since time.Time, channel string) (map[string]int, error)
	InsertBatch(ctx context.Context, logs []ContactLog) error
}
