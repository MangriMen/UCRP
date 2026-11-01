package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (LoginResult, error) {
	p, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return LoginResult{}, domain.ErrUnauthorized
	}
	u, err := s.users.GetByID(ctx, p.UserID)
	if err != nil || !u.IsActive || u.AuthVersion != p.AuthVersion {
		return LoginResult{}, domain.ErrUnauthorized
	}
	pair, err := s.tokens.Issue(domain.Principal{UserID: u.ID, Role: u.Role, TenantID: u.TenantID, AuthVersion: u.AuthVersion})
	if err != nil {
		return LoginResult{}, err
	}
	me, err := s.meWithEffective(ctx, u)
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

func (s *AuthService) ListUsers(ctx context.Context, limit, offset int) ([]UserDTO, error) {
	items, err := s.users.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]UserDTO, 0, len(items))
	for _, u := range items {
		out = append(out, UserDTO{
			ID: u.ID, Email: u.Email, DisplayName: u.DisplayName,
			Role: string(u.Role), TenantID: u.TenantID, IsActive: u.IsActive,
			CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		})
	}
	return out, nil
}

func (s *AuthService) GetUser(ctx context.Context, id string) (*UserDTO, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UserDTO{
		ID: u.ID, Email: u.Email, DisplayName: u.DisplayName,
		Role: string(u.Role), TenantID: u.TenantID, IsActive: u.IsActive,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}, nil
}

type UserDTO struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role"`
	TenantID    string    `json:"tenantId,omitempty"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UserCreateRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	TenantID    string `json:"tenantId"`
}

type UserUpdateRequest struct {
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	TenantID    string `json:"tenantId"`
	Password    string `json:"password"`
	IsActive    *bool  `json:"isActive"`
}

func (s *AuthService) CreateUserService(ctx context.Context, req UserCreateRequest) (*UserDTO, error) {
	email := strings.TrimSpace(req.Email)
	if email == "" || len(req.Password) < 6 || strings.TrimSpace(req.DisplayName) == "" {
		return nil, domain.ErrValidation
	}
	role := domain.Role(req.Role)
	if role != domain.RoleMarketer && role != domain.RoleManager && role != domain.RoleTenant && role != domain.RoleAnalyst {
		return nil, domain.ErrValidation
	}
	if role == domain.RoleTenant && req.TenantID == "" {
		return nil, domain.ErrValidation
	}
	if role != domain.RoleTenant && req.TenantID != "" {
		return nil, domain.ErrValidation
	}
	pub, err := s.CreateUser(ctx, email, req.Password, req.DisplayName, role, req.TenantID)
	if err != nil {
		return nil, err
	}
	created, err := s.users.GetByID(ctx, pub.UserID)
	if err != nil {
		return nil, err
	}
	return &UserDTO{
		ID: created.ID, Email: created.Email, DisplayName: created.DisplayName,
		Role: string(created.Role), TenantID: created.TenantID, IsActive: created.IsActive,
		CreatedAt: created.CreatedAt, UpdatedAt: created.UpdatedAt,
	}, nil
}

func (s *AuthService) UpdateUser(ctx context.Context, id string, req UserUpdateRequest) (*UserDTO, error) {
	if req.Role != "" || req.TenantID != "" || req.Password != "" || req.IsActive != nil {
		return nil, domain.ErrForbidden
	}
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.DisplayName != "" {
		u.DisplayName = req.DisplayName
	}
	u.UpdatedAt = s.clock.Now().UTC().Truncate(time.Second)
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return &UserDTO{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName, Role: string(u.Role), TenantID: u.TenantID, IsActive: u.IsActive}, nil
}

func (s *AuthService) DeactivateUser(ctx context.Context, id string, actorRole domain.Role) error {
	if actorRole == domain.RoleManager {
		n, err := s.users.CountActiveByRole(ctx, domain.RoleManager)
		if err != nil {
			return err
		}
		u, err := s.users.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if u.Role == domain.RoleManager && n <= 1 {
			return domain.ErrConflict
		}
	}
	return s.users.Deactivate(ctx, id, s.clock.Now())
}

type AuditService struct {
	audit domain.AuditRepo
}

func NewAuditService(audit domain.AuditRepo) *AuditService {
	return &AuditService{audit: audit}
}

type AuditDTO struct {
	ID         string `json:"id"`
	Actor      string `json:"actorUserId,omitempty"`
	Role       string `json:"actorRole"`
	Action     string `json:"action"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	CreatedAt  string `json:"createdAt"`
}

func (s *AuditService) List(ctx context.Context, entityType, entityID, actor string, limit, offset int) ([]AuditDTO, error) {
	items, err := s.audit.List(ctx, entityType, entityID, actor, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]AuditDTO, 0, len(items))
	for _, a := range items {
		out = append(out, AuditDTO{
			ID: a.ID, Actor: a.ActorUserID, Role: a.ActorRole, Action: a.Action,
			EntityType: a.EntityType, EntityID: a.EntityID,
			CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

type CampaignService struct {
	campaigns domain.CampaignRepo
	clock     domain.Clock
	idGen     domain.IDGen
}

func NewCampaignService(campaigns domain.CampaignRepo) *CampaignService {
	return &CampaignService{campaigns: campaigns, clock: domain.SystemClock{}, idGen: idgen.UUID{}}
}

type CampaignDTO struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	TargetTenantID   string     `json:"targetTenantId"`
	TargetTenantName string     `json:"targetTenantName,omitempty"`
	SegmentID        string     `json:"segmentId"`
	SegmentName      string     `json:"segmentName,omitempty"`
	AttributionDays  int        `json:"attributionDays"`
	Status           string     `json:"status"`
	StartsAt         *time.Time `json:"startsAt"`
	EndsAt           *time.Time `json:"endsAt"`
	CreatedBy        string     `json:"createdBy,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type CampaignRequest struct {
	Name            string `json:"name"`
	TargetTenantID  string `json:"targetTenantId"`
	SegmentID       string `json:"segmentId"`
	AttributionDays int    `json:"attributionDays"`
	Status          string `json:"status"`
}

func (s *CampaignService) List(ctx context.Context, limit, offset int) ([]CampaignDTO, error) {
	items, err := s.campaigns.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]CampaignDTO, 0, len(items))
	for _, c := range items {
		out = append(out, toCampaignDTO(c))
	}
	return out, nil
}

func (s *CampaignService) Get(ctx context.Context, id string) (*CampaignDTO, error) {
	c, err := s.campaigns.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	d := toCampaignDTO(*c)
	return &d, nil
}

func (s *CampaignService) Create(ctx context.Context, req CampaignRequest, actor string) (*CampaignDTO, error) {
	if strings.TrimSpace(req.Name) == "" || req.TargetTenantID == "" || req.SegmentID == "" {
		return nil, domain.ErrValidation
	}
	days := req.AttributionDays
	if days <= 0 {
		days = 14
	}
	now := s.clock.Now()
	c := &domain.Campaign{
		ID:              s.idGen.NewID(ctx),
		Name:            strings.TrimSpace(req.Name),
		TenantID:        req.TargetTenantID,
		SegmentID:       req.SegmentID,
		AttributionDays: days,
		Status:          "active",
		CreatedBy:       actor,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.campaigns.Insert(ctx, c); err != nil {
		return nil, err
	}
	d := toCampaignDTO(*c)
	return &d, nil
}

func (s *CampaignService) Update(ctx context.Context, id string, req CampaignRequest) (*CampaignDTO, error) {
	c, err := s.campaigns.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if hasExports, _ := s.campaigns.HasExports(ctx, id); hasExports {
		if req.TargetTenantID != "" && req.TargetTenantID != c.TenantID {
			return nil, domain.ErrConflict
		}
	}
	if req.Name != "" {
		c.Name = req.Name
	}
	if req.SegmentID != "" {
		c.SegmentID = req.SegmentID
	}
	if req.AttributionDays > 0 {
		c.AttributionDays = req.AttributionDays
	}
	if req.Status != "" {
		c.Status = req.Status
	}
	c.UpdatedAt = s.clock.Now()
	if err := s.campaigns.Update(ctx, c); err != nil {
		return nil, err
	}
	d := toCampaignDTO(*c)
	return &d, nil
}

func (s *CampaignService) Delete(ctx context.Context, id string) error {
	if has, err := s.campaigns.HasExports(ctx, id); err != nil {
		return err
	} else if has {
		return domain.ErrConflict
	}
	return s.campaigns.Delete(ctx, id)
}

type CampaignResultsDTO struct {
	CampaignID            string   `json:"campaignId"`
	SegmentID             string   `json:"segmentId,omitempty"`
	AttributionDays       int      `json:"attributionDays"`
	Recipients            int      `json:"recipients"`
	ControlGroup          int      `json:"controlGroup"`
	Conversions           int      `json:"conversions"`
	ConversionRate        float64  `json:"conversionRate"`
	ControlConversions    int      `json:"controlConversions"`
	ControlConversionRate float64  `json:"controlConversionRate"`
	ConversionLift        float64  `json:"conversionLift"`
	IncrementalVisits     int      `json:"incrementalVisits"`
	PValue                *float64 `json:"pValue"`
	IsSignificant         *bool    `json:"isSignificant"`
	ComputedAt            string   `json:"computedAt"`
}

func (s *CampaignService) Results(ctx context.Context, id string) (*CampaignResultsDTO, error) {
	c, err := s.campaigns.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	conversion, err := s.campaigns.Conversion(ctx, domain.ConversionFilter{CampaignID: id})
	if err != nil {
		return nil, err
	}
	out := &CampaignResultsDTO{
		CampaignID:            c.ID,
		SegmentID:             c.SegmentID,
		AttributionDays:       c.AttributionDays,
		Recipients:            conversion.Recipients,
		ControlGroup:          conversion.ControlGroup,
		Conversions:           conversion.Conversions,
		ConversionRate:        conversion.Rate(),
		ControlConversions:    conversion.ControlConversions,
		ControlConversionRate: conversion.ControlRate(),
		ConversionLift:        conversion.Lift(),
		IncrementalVisits:     int(math.Round(conversion.Lift() * float64(conversion.Recipients))),
		ComputedAt:            s.clock.Now().UTC().Format(time.RFC3339),
	}
	if pValue, ok := conversion.PValue(); ok {
		significant := pValue < 0.05
		out.PValue = &pValue
		out.IsSignificant = &significant
	}
	return out, nil
}

func toCampaignDTO(c domain.Campaign) CampaignDTO {
	return CampaignDTO{
		ID: c.ID, Name: c.Name, TargetTenantID: c.TenantID, TargetTenantName: c.TenantName,
		SegmentID: c.SegmentID, SegmentName: c.SegmentName,
		AttributionDays: c.AttributionDays, Status: c.Status, StartsAt: c.StartsAt, EndsAt: c.EndsAt,
		CreatedBy: c.CreatedBy, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}
