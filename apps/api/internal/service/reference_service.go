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

const (
	maxTenantCodeLength   = 64
	maxTenantNameLength   = 200
	defaultTenantPriority = 50
)

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
	Priority   *int   `json:"priority"`
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

type TenantStatsAccess struct {
	Visits    bool
	NewGuests bool
}

type TenantListItemDTO struct {
	TenantDTO
	CategoryName        string `json:"categoryName"`
	GuestsCount         *int   `json:"guestsCount"`
	VisitsCount         *int   `json:"visitsCount"`
	NewGuestsCount      *int   `json:"newGuestsCount"`
	Suppressed          bool   `json:"suppressed"`
	SourcesCount        int    `json:"sourcesCount"`
	ProblemSourcesCount int    `json:"problemSourcesCount"`
}

func (s *ReferenceService) SearchTenants(ctx context.Context, filter domain.TenantFilter, access TenantStatsAccess, limit, offset int) ([]TenantListItemDTO, int, error) {
	entries, total, err := s.tenants.Search(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]TenantListItemDTO, 0, len(entries))
	for _, e := range entries {
		item := TenantListItemDTO{
			TenantDTO:           toTenantDTO(e.Tenant),
			CategoryName:        e.CategoryName,
			SourcesCount:        e.Stats.SourcesCount,
			ProblemSourcesCount: e.Stats.ProblemSourcesCount,
		}
		if access.Visits {
			item.GuestsCount, item.VisitsCount = e.Stats.GuestsCount, e.Stats.VisitsCount
			item.Suppressed = e.Stats.VisitsCount == nil
		}
		if access.NewGuests {
			item.NewGuestsCount = e.Stats.NewGuestsCount
			item.Suppressed = item.Suppressed || e.Stats.NewGuestsCount == nil
		}
		out = append(out, item)
	}
	return out, total, nil
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
	if len([]rune(code)) > maxTenantCodeLength || len([]rune(name)) > maxTenantNameLength {
		return nil, domain.ErrValidation
	}
	if err := s.requireActiveCategory(ctx, req.CategoryID); err != nil {
		return nil, err
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
	if !validTenantKind(kind) {
		return nil, domain.ErrValidation
	}
	priority := defaultTenantPriority
	if req.Priority != nil {
		priority = *req.Priority
	}
	if !validTenantPriority(priority) {
		return nil, domain.ErrValidation
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
		name := strings.TrimSpace(req.Name)
		if name == "" || len([]rune(name)) > maxTenantNameLength {
			return nil, domain.ErrValidation
		}
		t.Name = name
	}
	if req.CategoryID != "" && req.CategoryID != t.CategoryID {
		if err := s.requireActiveCategory(ctx, req.CategoryID); err != nil {
			return nil, err
		}
		t.CategoryID = req.CategoryID
	}
	if req.MallID != "" {
		t.MallID = req.MallID
	}
	if req.Kind != "" {
		if !validTenantKind(req.Kind) {
			return nil, domain.ErrValidation
		}
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
		if !validTenantPriority(*req.Priority) {
			return nil, domain.ErrValidation
		}
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

func (s *ReferenceService) requireActiveCategory(ctx context.Context, categoryID string) error {
	category, err := s.categories.GetByID(ctx, categoryID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrValidation
	}
	if err != nil {
		return err
	}
	if !category.IsActive {
		return domain.ErrValidation
	}
	return nil
}

func validTenantKind(kind string) bool {
	return kind == "mall" || kind == "external"
}

func validTenantPriority(priority int) bool {
	return priority >= 0 && priority <= 100
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

const defaultRelationWeightPct = 100

type RelationDTO struct {
	CategoryID          string `json:"categoryId"`
	RelatedCategoryID   string `json:"relatedCategoryId"`
	RelatedCategoryName string `json:"relatedCategoryName,omitempty"`
	RelationType        string `json:"relationType"`
	WeightPct           *int   `json:"weightPct"`
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
		weight := r.WeightPct
		out = append(out, RelationDTO{
			CategoryID:          categoryID,
			RelatedCategoryID:   r.RelatedCategoryID,
			RelatedCategoryName: r.RelatedCategoryName,
			RelationType:        r.RelationType,
			WeightPct:           &weight,
		})
	}
	return out, nil
}

func (s *ReferenceService) ReplaceRelations(ctx context.Context, categoryID string, rels []RelationDTO) ([]RelationDTO, error) {
	if _, err := s.categories.GetByID(ctx, categoryID); err != nil {
		return nil, err
	}
	current, err := s.categories.ListRelations(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	alreadyRelated := make(map[string]bool, len(current))
	for _, r := range current {
		alreadyRelated[r.RelatedCategoryID] = true
	}
	seen := make(map[string]bool, len(rels))
	converted := make([]domain.CategoryRelation, 0, len(rels))
	for _, r := range rels {
		rel, err := s.validateRelation(ctx, categoryID, r, alreadyRelated)
		if err != nil {
			return nil, err
		}
		if seen[rel.RelatedCategoryID] {
			return nil, domain.ErrValidation
		}
		seen[rel.RelatedCategoryID] = true
		converted = append(converted, rel)
	}
	if err := s.categories.ReplaceRelations(ctx, categoryID, converted); err != nil {
		return nil, err
	}
	return s.ListRelations(ctx, categoryID)
}

func (s *ReferenceService) validateRelation(ctx context.Context, categoryID string, r RelationDTO, alreadyRelated map[string]bool) (domain.CategoryRelation, error) {
	if r.RelationType != "adjacent" && r.RelationType != "competitor" {
		return domain.CategoryRelation{}, domain.ErrValidation
	}
	relatedID := strings.TrimSpace(r.RelatedCategoryID)
	if r.CategoryID != categoryID || relatedID == "" || relatedID == categoryID {
		return domain.CategoryRelation{}, domain.ErrValidation
	}
	weight := defaultRelationWeightPct
	if r.WeightPct != nil {
		weight = *r.WeightPct
	}
	if weight < 0 || weight > 100 {
		return domain.CategoryRelation{}, domain.ErrValidation
	}
	related, err := s.categories.GetByID(ctx, relatedID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.CategoryRelation{}, domain.ErrValidation
	}
	if err != nil {
		return domain.CategoryRelation{}, err
	}
	if !related.IsActive && !alreadyRelated[relatedID] {
		return domain.CategoryRelation{}, domain.ErrValidation
	}
	return domain.CategoryRelation{
		CategoryID:        categoryID,
		RelatedCategoryID: relatedID,
		RelationType:      r.RelationType,
		WeightPct:         weight,
	}, nil
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
