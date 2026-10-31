package service

import (
	"context"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/ingestion"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type DataSourceService struct {
	sources domain.DataSourceRepo
	types   domain.DataSourceTypeRepo
	clock   domain.Clock
	idGen   domain.IDGen
}

func NewDataSourceService(sources domain.DataSourceRepo, types domain.DataSourceTypeRepo) *DataSourceService {
	return &DataSourceService{
		sources: sources,
		types:   types,
		clock:   domain.SystemClock{},
		idGen:   idgen.UUID{},
	}
}

type DataSourceDTO struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id,omitempty"`
	TypeID       string     `json:"type_id"`
	TypeCode     string     `json:"type_code"`
	Name         string     `json:"name"`
	Config       string     `json:"config,omitempty"`
	FieldMapping string     `json:"field_mapping,omitempty"`
	ScheduleCron string     `json:"schedule_cron,omitempty"`
	IsActive     bool       `json:"is_active"`
	LastImportAt *time.Time `json:"last_import_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateDataSourceRequest struct {
	TenantID     string `json:"tenant_id"`
	TypeID       string `json:"type_id"`
	Name         string `json:"name"`
	Config       string `json:"config"`
	FieldMapping string `json:"field_mapping"`
	ScheduleCron string `json:"schedule_cron"`
}

type DataSourceTypeDTO struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	ConnectorKey string `json:"connector_key"`
	SupportsFile bool   `json:"supports_file"`
	SupportsPull bool   `json:"supports_pull"`
}

func (s *DataSourceService) Create(ctx context.Context, req CreateDataSourceRequest) (*DataSourceDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || req.TypeID == "" {
		return nil, domain.ErrValidation
	}
	if req.FieldMapping != "" {
		if _, err := ingestion.ParseFieldMapping(req.FieldMapping); err != nil {
			return nil, domain.ErrValidation
		}
	}
	now := s.clock.Now()
	ds := &domain.DataSource{
		ID:           s.idGen.NewID(ctx),
		TenantID:     req.TenantID,
		TypeID:       req.TypeID,
		Name:         name,
		Config:       req.Config,
		FieldMapping: req.FieldMapping,
		ScheduleCron: req.ScheduleCron,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.sources.Insert(ctx, ds); err != nil {
		return nil, err
	}
	created, err := s.sources.GetByID(ctx, ds.ID)
	if err != nil {
		return nil, err
	}
	d := toDataSourceDTO(*created)
	return &d, nil
}

func (s *DataSourceService) Get(ctx context.Context, id string) (*DataSourceDTO, error) {
	ds, err := s.sources.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	d := toDataSourceDTO(*ds)
	return &d, nil
}

func (s *DataSourceService) List(ctx context.Context, limit, offset int) ([]DataSourceDTO, error) {
	items, err := s.sources.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]DataSourceDTO, 0, len(items))
	for _, ds := range items {
		out = append(out, toDataSourceDTO(ds))
	}
	return out, nil
}

func (s *DataSourceService) ListTypes(ctx context.Context) ([]DataSourceTypeDTO, error) {
	items, err := s.types.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]DataSourceTypeDTO, 0, len(items))
	for _, t := range items {
		out = append(out, DataSourceTypeDTO{
			ID:           t.ID,
			Code:         t.Code,
			Name:         t.Name,
			ConnectorKey: t.ConnectorKey,
			SupportsFile: t.SupportsFile,
			SupportsPull: t.SupportsPull,
		})
	}
	return out, nil
}

func toDataSourceDTO(ds domain.DataSource) DataSourceDTO {
	return DataSourceDTO{
		ID:           ds.ID,
		TenantID:     ds.TenantID,
		TypeID:       ds.TypeID,
		TypeCode:     ds.TypeCode,
		Name:         ds.Name,
		Config:       ds.Config,
		FieldMapping: ds.FieldMapping,
		ScheduleCron: ds.ScheduleCron,
		IsActive:     ds.IsActive,
		LastImportAt: ds.LastImportAt,
		CreatedAt:    ds.CreatedAt,
		UpdatedAt:    ds.UpdatedAt,
	}
}
