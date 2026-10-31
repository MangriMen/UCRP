package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type UserSQL struct {
	db *sql.DB
}

func NewUserSQL(db *sql.DB) *UserSQL {
	return &UserSQL{db: db}
}

func (r *UserSQL) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return r.getOne(ctx, `
		SELECT id, email, password_hash, display_name, role, COALESCE(tenant_id, ''), is_active, last_login_at, created_at
		FROM app_user WHERE id = $1`, id)
}

func (r *UserSQL) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.getOne(ctx, `
		SELECT id, email, password_hash, display_name, role, COALESCE(tenant_id, ''), is_active, last_login_at, created_at
		FROM app_user WHERE email = $1`, email)
}

func (r *UserSQL) Insert(ctx context.Context, u *domain.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO app_user (id, email, password_hash, display_name, role, tenant_id, is_active, last_login_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		u.ID, u.Email, u.PasswordHash, u.DisplayName, string(u.Role), nullIfEmpty(u.TenantID), u.IsActive, u.LastLoginAt, u.CreatedAt,
	)
	return err
}

func (r *UserSQL) TouchLastLogin(ctx context.Context, id string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE app_user SET last_login_at = $1 WHERE id = $2`, at, id)
	return err
}

func (r *UserSQL) getOne(ctx context.Context, query string, arg any) (*domain.User, error) {
	var (
		u         domain.User
		role      string
		active    bool
		lastLogin sql.NullTime
		created   time.Time
	)
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &role, &u.TenantID, &active, &lastLogin, &created,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u.Role = domain.Role(role)
	u.IsActive = active
	u.CreatedAt = created
	if lastLogin.Valid {
		t := lastLogin.Time
		u.LastLoginAt = &t
	}
	return &u, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
