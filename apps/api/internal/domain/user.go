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
	LastLoginAt  *time.Time
	CreatedAt    time.Time
}

type UserRepo interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Insert(ctx context.Context, u *User) error
	TouchLastLogin(ctx context.Context, id string, at time.Time) error
}
