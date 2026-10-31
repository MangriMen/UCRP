package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type ReferenceService struct {
	tenants    domain.TenantRepo
	malls      domain.MallRepo
	categories domain.TenantCategoryRepo
	projects   domain.ProjectRepo
	eventTypes domain.EventTypeRepo
	clock      domain.Clock
	idGen      domain.IDGen
}

func NewReferenceService(
	tenants domain.TenantRepo,
	malls domain.MallRepo,
	categories domain.TenantCategoryRepo,
	projects domain.ProjectRepo,
	eventTypes domain.EventTypeRepo,
) *ReferenceService {
	return &ReferenceService{
		tenants:    tenants,
		malls:      malls,
		categories: categories,
		projects:   projects,
		eventTypes: eventTypes,
		clock:      domain.SystemClock{},
		idGen:      idgen.UUID{},
	}
}

type TenantDTO struct {
	ID         string    `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	MallID     string    `json:"mallId"`
	CategoryID string    `json:"categoryId"`
	Kind       string    `json:"kind"`
	Floor      *int      `json:"floor"`
	Priority   int       `json:"priority"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type OptionalFloor struct {
	Set   bool
	Value *int
}

func (o *OptionalFloor) UnmarshalJSON(data []byte) error {
	o.Set = true
	return json.Unmarshal(data, &o.Value)
}

type TenantCreateRequest struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	MallID     string `json:"mallId"`
	CategoryID string `json:"categoryId"`
	Kind       string `json:"kind"`
	Floor      *int   `json:"floor"`
	Priority   int    `json:"priority"`
}

type TenantUpdateRequest struct {
	Name       string        `json:"name"`
	MallID     string        `json:"mallId"`
	CategoryID string        `json:"categoryId"`
	Kind       string        `json:"kind"`
	Floor      OptionalFloor `json:"floor"`
	Priority   *int          `json:"priority"`
	IsActive   *bool         `json:"isActive"`
}

func (s *ReferenceService) ListTenants(ctx context.Context, limit, offset int) ([]TenantDTO, error) {
	items, err := s.tenants.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]TenantDTO, 0, len(items))
	for _, t := range items {
		out = append(out, toTenantDTO(t))
	}
	return out, nil
}

func (s *ReferenceService) GetTenant(ctx context.Context, id string) (*TenantDTO, error) {
	t, err := s.tenants.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	d := toTenantDTO(*t)
	return &d, nil
}

func (s *ReferenceService) CreateTenant(ctx context.Context, req TenantCreateRequest) (*TenantDTO, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" || req.CategoryID == "" || req.MallID == "" {
		return nil, domain.ErrValidation
	}
	mall, err := s.tenantMall(ctx, req.MallID)
	if err != nil {
		return nil, err
	}
	if req.Floor != nil && !mall.HasFloor(*req.Floor) {
		return nil, domain.ErrValidation
	}
	kind := req.Kind
	if kind == "" {
		kind = "mall"
	}
	priority := req.Priority
	if priority == 0 {
		priority = 50
	}
	now := s.clock.Now()
	t := &domain.Tenant{
		ID:         s.idGen.NewID(ctx),
		Code:       code,
		Name:       name,
		MallID:     req.MallID,
		CategoryID: req.CategoryID,
		Kind:       kind,
		Floor:      req.Floor,
		Priority:   priority,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.tenants.Insert(ctx, t); err != nil {
		return nil, err
	}
	d := toTenantDTO(*t)
	return &d, nil
}

func (s *ReferenceService) UpdateTenant(ctx context.Context, id string, req TenantUpdateRequest) (*TenantDTO, error) {
	t, err := s.tenants.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		t.Name = req.Name
	}
	if req.CategoryID != "" {
		t.CategoryID = req.CategoryID
	}
	if req.MallID != "" {
		t.MallID = req.MallID
	}
	if req.Kind != "" {
		t.Kind = req.Kind
	}
	if req.Floor.Set {
		t.Floor = req.Floor.Value
	}
	if req.MallID != "" || req.Floor.Set {
		mall, err := s.tenantMall(ctx, t.MallID)
		if err != nil {
			return nil, err
		}
		if t.Floor != nil && !mall.HasFloor(*t.Floor) {
			return nil, domain.ErrValidation
		}
	}
	if req.Priority != nil {
		t.Priority = *req.Priority
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}
	t.UpdatedAt = s.clock.Now()
	if err := s.tenants.Update(ctx, t); err != nil {
		return nil, err
	}
	d := toTenantDTO(*t)
	return &d, nil
}

func (s *ReferenceService) tenantMall(ctx context.Context, mallID string) (*domain.Mall, error) {
	mall, err := s.malls.GetByID(ctx, mallID)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrValidation
	}
	return mall, err
}

func (s *ReferenceService) DeleteTenant(ctx context.Context, id string) error {
	return s.tenants.SoftDelete(ctx, id, s.clock.Now())
}

type CategoryDTO struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}

type CategoryRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive *bool  `json:"isActive"`
}

func (s *ReferenceService) ListCategories(ctx context.Context) ([]CategoryDTO, error) {
	items, err := s.categories.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryDTO, 0, len(items))
	for _, c := range items {
		out = append(out, CategoryDTO{ID: c.ID, Code: c.Code, Name: c.Name, IsActive: c.IsActive})
	}
	return out, nil
}

func (s *ReferenceService) CreateCategory(ctx context.Context, req CategoryRequest) (*CategoryDTO, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return nil, domain.ErrValidation
	}
	c := &domain.TenantCategory{
		ID:        s.idGen.NewID(ctx),
		Code:      code,
		Name:      name,
		IsActive:  true,
		CreatedAt: s.clock.Now(),
	}
	if err := s.categories.Insert(ctx, c); err != nil {
		return nil, err
	}
	return &CategoryDTO{ID: c.ID, Code: c.Code, Name: c.Name, IsActive: c.IsActive}, nil
}

func (s *ReferenceService) UpdateCategory(ctx context.Context, id string, req CategoryRequest) (*CategoryDTO, error) {
	c, err := s.categories.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		c.Name = req.Name
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}
	if err := s.categories.Update(ctx, c); err != nil {
		return nil, err
	}
	return &CategoryDTO{ID: c.ID, Code: c.Code, Name: c.Name, IsActive: c.IsActive}, nil
}

func (s *ReferenceService) DeleteCategory(ctx context.Context, id string) error {
	return s.categories.Delete(ctx, id)
}

type RelationDTO struct {
	CategoryID        string `json:"categoryId"`
	RelatedCategoryID string `json:"relatedCategoryId"`
	RelationType      string `json:"relationType"`
	WeightPct         int    `json:"weightPct"`
}

func (s *ReferenceService) ListRelations(ctx context.Context, categoryID string) ([]RelationDTO, error) {
	if _, err := s.categories.GetByID(ctx, categoryID); err != nil {
		return nil, err
	}
	items, err := s.categories.ListRelations(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	out := make([]RelationDTO, 0, len(items))
	for _, r := range items {
		out = append(out, RelationDTO{CategoryID: categoryID, RelatedCategoryID: r.RelatedCategoryID, RelationType: r.RelationType, WeightPct: r.WeightPct})
	}
	return out, nil
}

func (s *ReferenceService) ReplaceRelations(ctx context.Context, categoryID string, rels []RelationDTO) ([]RelationDTO, error) {
	if _, err := s.categories.GetByID(ctx, categoryID); err != nil {
		return nil, err
	}
	converted := make([]domain.CategoryRelation, 0, len(rels))
	for _, r := range rels {
		if r.RelationType != "adjacent" && r.RelationType != "competitor" {
			return nil, domain.ErrValidation
		}
		if strings.TrimSpace(r.CategoryID) == "" || strings.TrimSpace(r.RelatedCategoryID) == "" {
			return nil, domain.ErrValidation
		}
		if r.CategoryID != categoryID {
			return nil, domain.ErrValidation
		}
		converted = append(converted, domain.CategoryRelation{
			CategoryID:        categoryID,
			RelatedCategoryID: r.RelatedCategoryID,
			RelationType:      r.RelationType,
			WeightPct:         r.WeightPct,
		})
	}
	if err := s.categories.ReplaceRelations(ctx, categoryID, converted); err != nil {
		return nil, err
	}
	return s.ListRelations(ctx, categoryID)
}

type ProjectDTO struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}

type ProjectRequest struct {
	TenantID string `json:"tenantId"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive *bool  `json:"isActive"`
}

func (s *ReferenceService) ListProjects(ctx context.Context, tenantID string, limit, offset int) ([]ProjectDTO, error) {
	items, err := s.projects.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]ProjectDTO, 0, len(items))
	for _, p := range items {
		out = append(out, ProjectDTO{ID: p.ID, TenantID: p.TenantID, Code: p.Code, Name: p.Name, IsActive: p.IsActive})
	}
	return out, nil
}

func (s *ReferenceService) CreateProject(ctx context.Context, req ProjectRequest) (*ProjectDTO, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, domain.ErrValidation
	}
	p := &domain.Project{
		ID:       s.idGen.NewID(ctx),
		TenantID: req.TenantID,
		Code:     strings.TrimSpace(req.Code),
		Name:     strings.TrimSpace(req.Name),
		IsActive: true,
	}
	if err := s.projects.Insert(ctx, p); err != nil {
		return nil, err
	}
	return &ProjectDTO{ID: p.ID, TenantID: p.TenantID, Code: p.Code, Name: p.Name, IsActive: p.IsActive}, nil
}

func (s *ReferenceService) UpdateProject(ctx context.Context, id string, req ProjectRequest) (*ProjectDTO, error) {
	p, err := s.projects.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	if err := s.projects.Update(ctx, p); err != nil {
		return nil, err
	}
	return &ProjectDTO{ID: p.ID, TenantID: p.TenantID, Code: p.Code, Name: p.Name, IsActive: p.IsActive}, nil
}

func (s *ReferenceService) DeleteProject(ctx context.Context, id string) error {
	return s.projects.Delete(ctx, id)
}

type EventTypeDTO struct {
	ID               string `json:"id"`
	Code             string `json:"code"`
	Name             string `json:"name"`
	HasAmount        bool   `json:"hasAmount"`
	CountsAsPresence bool   `json:"countsAsPresence"`
	IsActive         bool   `json:"isActive"`
}

type EventTypeRequest struct {
	Code             string `json:"code"`
	Name             string `json:"name"`
	HasAmount        bool   `json:"hasAmount"`
	CountsAsPresence bool   `json:"countsAsPresence"`
	IsActive         *bool  `json:"isActive"`
}

func (s *ReferenceService) ListEventTypes(ctx context.Context) ([]EventTypeDTO, error) {
	items, err := s.eventTypes.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]EventTypeDTO, 0, len(items))
	for _, e := range items {
		out = append(out, EventTypeDTO{
			ID: e.ID, Code: e.Code, Name: e.Name,
			HasAmount: e.HasAmount, CountsAsPresence: e.CountsAsPresence, IsActive: e.IsActive,
		})
	}
	return out, nil
}

func (s *ReferenceService) CreateEventType(ctx context.Context, req EventTypeRequest) (*EventTypeDTO, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" || strings.TrimSpace(req.Name) == "" {
		return nil, domain.ErrValidation
	}
	e := &domain.EventType{
		ID:               s.idGen.NewID(ctx),
		Code:             code,
		Name:             strings.TrimSpace(req.Name),
		HasAmount:        req.HasAmount,
		CountsAsPresence: req.CountsAsPresence,
		IsActive:         true,
	}
	if err := s.eventTypes.Insert(ctx, e); err != nil {
		return nil, err
	}
	return &EventTypeDTO{
		ID: e.ID, Code: e.Code, Name: e.Name,
		HasAmount: e.HasAmount, CountsAsPresence: e.CountsAsPresence, IsActive: e.IsActive,
	}, nil
}

func (s *ReferenceService) UpdateEventType(ctx context.Context, id string, req EventTypeRequest) (*EventTypeDTO, error) {
	e, err := s.eventTypes.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		e.Name = req.Name
	}
	e.HasAmount = req.HasAmount
	e.CountsAsPresence = req.CountsAsPresence
	if req.IsActive != nil {
		e.IsActive = *req.IsActive
	}
	if err := s.eventTypes.Update(ctx, e); err != nil {
		return nil, err
	}
	return &EventTypeDTO{
		ID: e.ID, Code: e.Code, Name: e.Name,
		HasAmount: e.HasAmount, CountsAsPresence: e.CountsAsPresence, IsActive: e.IsActive,
	}, nil
}

func (s *ReferenceService) DeleteEventType(ctx context.Context, id string) error {
	return s.eventTypes.Delete(ctx, id)
}

func toTenantDTO(t domain.Tenant) TenantDTO {
	return TenantDTO{
		ID: t.ID, Code: t.Code, Name: t.Name, MallID: t.MallID, CategoryID: t.CategoryID,
		Kind: t.Kind, Floor: t.Floor, Priority: t.Priority, IsActive: t.IsActive,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}
