package service

import (
	"context"
	"errors"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
)

type AuthService struct {
	users  domain.UserRepo
	tokens *auth.Manager
	clock  domain.Clock
	idGen  domain.IDGen
}

func NewAuthService(users domain.UserRepo, tokens *auth.Manager, clock domain.Clock, idGen domain.IDGen) *AuthService {
	return &AuthService{users: users, tokens: tokens, clock: clock, idGen: idGen}
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         PublicUser
}

type PublicUser struct {
	ID       string      `json:"id"`
	Role     domain.Role `json:"role"`
	TenantID string      `json:"tenant_id"`
}

func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return LoginResult{}, domain.ErrUnauthorized
	}
	if !u.IsActive || !auth.CheckPassword(u.PasswordHash, password) {
		return LoginResult{}, domain.ErrUnauthorized
	}
	pair, err := s.tokens.Issue(domain.Principal{UserID: u.ID, Role: u.Role, TenantID: u.TenantID})
	if err != nil {
		return LoginResult{}, err
	}
	if err := s.users.TouchLastLogin(ctx, u.ID, s.clock.Now()); err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		User:         PublicUser{ID: u.ID, Role: u.Role, TenantID: u.TenantID},
	}, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (PublicUser, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return PublicUser{}, domain.ErrNotFound
		}
		return PublicUser{}, err
	}
	return PublicUser{ID: u.ID, Role: u.Role, TenantID: u.TenantID}, nil
}

func (s *AuthService) CreateUser(ctx context.Context, email, password, displayName string, role domain.Role, tenantID string) (PublicUser, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return PublicUser{}, err
	}
	u := &domain.User{
		ID:           s.idGen.NewID(ctx),
		Email:        email,
		PasswordHash: hash,
		DisplayName:  displayName,
		Role:         role,
		TenantID:     tenantID,
		IsActive:     true,
		CreatedAt:    s.clock.Now().UTC().Truncate(time.Second),
	}
	if err := s.users.Insert(ctx, u); err != nil {
		return PublicUser{}, err
	}
	return PublicUser{ID: u.ID, Role: u.Role, TenantID: u.TenantID}, nil
}
