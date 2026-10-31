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
	access *AccessGrantService
}

func (s *AuthService) SetAccess(access *AccessGrantService) {
	s.access = access
}

func NewAuthService(users domain.UserRepo, tokens *auth.Manager, clock domain.Clock, idGen domain.IDGen) *AuthService {
	return &AuthService{users: users, tokens: tokens, clock: clock, idGen: idGen}
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         MeResponse
}

type MeResponse struct {
	UserID        string   `json:"userId"`
	Email         string   `json:"email"`
	DisplayName   string   `json:"displayName"`
	Role          string   `json:"role"`
	TenantID      string   `json:"tenantId,omitempty"`
	Permissions   []string `json:"permissions"`
	PolicyVersion int64    `json:"policyVersion"`
}

func permissionsFor(role domain.Role) []string {
	switch role {
	case domain.RoleMarketer:
		return []string{"guests:read", "segments:write", "exports:write", "contact-limits:write"}
	case domain.RoleManager:
		return []string{"guests:read", "imports:read", "data-sources:read", "users:write", "audit:read"}
	case domain.RoleAnalyst:
		return []string{"guests:read", "guests:merge", "imports:write", "data-sources:write", "reference:write", "segments:write", "dashboard:read", "dq:read"}
	case domain.RoleTenant:
		return []string{"dashboard:read", "campaigns:write"}
	default:
		return []string{}
	}
}

func toMe(u *domain.User) MeResponse {
	m := MeResponse{
		UserID:      u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        string(u.Role),
		TenantID:    u.TenantID,
		Permissions: permissionsFor(u.Role),
	}
	if m.Permissions == nil {
		m.Permissions = []string{}
	}
	return m
}

func (s *AuthService) effectiveMe(ctx context.Context, u *domain.User) (MeResponse, error) {
	me := toMe(u)
	if s.access == nil {
		return me, nil
	}
	permissions, version, err := s.access.Effective(ctx, domain.Principal{UserID: u.ID, Role: u.Role, TenantID: u.TenantID})
	if err != nil {
		return MeResponse{}, err
	}
	me.Permissions = permissions
	me.PolicyVersion = version
	return me, nil
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
	me, err := s.effectiveMe(ctx, u)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		User:         me,
	}, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (MeResponse, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return MeResponse{}, domain.ErrNotFound
		}
		return MeResponse{}, err
	}
	return s.effectiveMe(ctx, u)
}

func (s *AuthService) CreateUser(ctx context.Context, email, password, displayName string, role domain.Role, tenantID string) (MeResponse, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return MeResponse{}, err
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
		return MeResponse{}, err
	}
	return toMe(u), nil
}
