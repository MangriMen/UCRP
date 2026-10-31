package domain

import (
	"context"
	"time"
)

type Permission struct {
	Code        string
	Domain      string
	Kind        string
	Title       string
	Description string
	IsSensitive bool
	MinK        *int
	IsActive    bool
}

type AccessRolePermission struct {
	PermissionCode string
	Effect         string
}

type AccessGrantItem struct {
	PermissionCode string
	Effect         string
}

type AccessGrantScope struct {
	Type  string
	Value string
}

type AccessGrant struct {
	ID            string
	SubjectType   string
	SubjectID     string
	PresetCode    string
	Status        string
	ValidFrom     time.Time
	ValidTo       *time.Time
	CreatedBy     string
	ApprovedBy    string
	Reason        string
	PolicyVersion int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Items         []AccessGrantItem
	Scopes        []AccessGrantScope
}

type AccessGrantFilter struct {
	SubjectType string
	SubjectID   string
	Status      string
	Limit       int
	Offset      int
}

type AccessRepository interface {
	ListPermissions(ctx context.Context) ([]Permission, error)
	RolePermissions(ctx context.Context, role Role) ([]AccessRolePermission, error)
	PolicyVersion(ctx context.Context) (int64, error)
	ActiveGrants(ctx context.Context, principal Principal, now time.Time) ([]AccessGrant, error)
	GetGrant(ctx context.Context, id string) (*AccessGrant, error)
	ListGrants(ctx context.Context, filter AccessGrantFilter) ([]AccessGrant, error)
	CreateGrant(ctx context.Context, grant *AccessGrant, expectedVersion int64, audit *AuditLog) error
	UpdateGrant(ctx context.Context, grant *AccessGrant, expectedVersion int64, audit *AuditLog) error
}
