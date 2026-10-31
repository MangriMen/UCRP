package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	Role         Role
	TenantID     string
	IsActive     bool
	AuthVersion  int64
	LastLoginAt  *time.Time
	CreatedAt    time.Time
}

type UserRepo interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	Insert(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	Deactivate(ctx context.Context, id string, now time.Time) error
	TouchLastLogin(ctx context.Context, id string, at time.Time) error
	CountActiveByRole(ctx context.Context, role Role) (int, error)
}
