package domain

import (
	"context"
	"time"
)

type ImportMode string

const (
	ImportModeFile ImportMode = "file"
	ImportModePull ImportMode = "pull"
)

type ImportStatus string

const (
	ImportQueued              ImportStatus = "queued"
	ImportRunning             ImportStatus = "running"
	ImportCompleted           ImportStatus = "completed"
	ImportCompletedWithErrors ImportStatus = "completed_with_errors"
	ImportFailed              ImportStatus = "failed"
)

type ImportBatch struct {
	ID             string
	DataSourceID   string
	DataSourceName string
	Mode           ImportMode
	FileHash       string
	FileName       string
	FileSizeBytes  int64
	Status         ImportStatus
	TotalRows      int
	AcceptedRows   int
	RejectedRows   int
	DuplicateRows  int
	ErrorMessage   string
	CreatedBy      string
	CreatedAt      time.Time
	StartedAt      *time.Time
	FinishedAt     *time.Time
	StorageKey     string
	WindowFrom     *time.Time
	WindowTo       *time.Time
}

type RejectedRow struct {
	ID            string
	ImportBatchID string
	RowNumber     int
	ErrorCode     string
	ErrorMessage  string
	RawLine       string
	CreatedAt     time.Time
}

type DataSourceType struct {
	ID           string
	Code         string
	Name         string
	ConnectorKey string
	SupportsFile bool
	SupportsPull bool
	IsActive     bool
}

type DataSource struct {
	ID           string
	TenantID     string
	TypeID       string
	TypeCode     string
	Name         string
	Config       string
	FieldMapping string
	ScheduleCron string
	IsActive     bool
	SupportsFile bool
	SupportsPull bool
	LastImportAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DataSourceFilter struct {
	Query        string
	TenantID     string
	TypeID       string
	IsActive     *bool
	ImportStatus string
	Sort         string
}

type DataSourceLastImport struct {
	BatchID      string
	Status       ImportStatus
	CreatedAt    time.Time
	FinishedAt   *time.Time
	RejectedRows int
}

type DataSourceListEntry struct {
	DataSource
	TenantName string
	LastImport *DataSourceLastImport
}

type ImportBatchRepo interface {
	CreateQueued(ctx context.Context, batch *ImportBatch, jobType, dedupKey string) error
	FindByHash(ctx context.Context, dataSourceID, fileHash string) (*ImportBatch, error)
	GetByID(ctx context.Context, id string) (*ImportBatch, error)
	List(ctx context.Context, dataSourceID, status string, limit, offset int) ([]ImportBatch, error)
	MarkRunning(ctx context.Context, id string, startedAt time.Time) error
	MarkFinished(ctx context.Context, id string, status ImportStatus, total, accepted, rejected, duplicate int, errMsg string, finishedAt time.Time) error
	HasActivePull(ctx context.Context, dataSourceID string) (bool, error)
}

type RejectedRowRepo interface {
	ListByBatch(ctx context.Context, batchID string, limit, offset int) ([]RejectedRow, error)
	InsertBatch(ctx context.Context, rows []RejectedRow) error
}

type DataSourceRepo interface {
	GetByID(ctx context.Context, id string) (*DataSource, error)
	Search(ctx context.Context, filter DataSourceFilter, limit, offset int) ([]DataSourceListEntry, int, error)
	Insert(ctx context.Context, ds *DataSource) error
	Update(ctx context.Context, ds *DataSource) error
	Deactivate(ctx context.Context, id string, now time.Time) error
	TouchLastImport(ctx context.Context, id string, at time.Time) error
}

type DataSourceTypeRepo interface {
	List(ctx context.Context) ([]DataSourceType, error)
	GetByID(ctx context.Context, id string) (*DataSourceType, error)
	Insert(ctx context.Context, t *DataSourceType) error
	Update(ctx context.Context, t *DataSourceType) error
	EnsureDefaults(ctx context.Context, clock Clock, idGen IDGen) error
}

type ImportCatalog interface {
	EventTypeID(ctx context.Context, code string) (string, bool, error)
	TenantIDByCode(ctx context.Context, code string) (string, bool, error)
	ProjectID(ctx context.Context, tenantID, code string) (string, bool, error)
	EnsureEventTypes(ctx context.Context, clock Clock, idGen IDGen) error
}
