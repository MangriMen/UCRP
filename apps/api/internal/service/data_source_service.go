package service

import (
	"context"
	"encoding/json"
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
	ID           string          `json:"id"`
	TenantID     string          `json:"tenantId,omitempty"`
	TypeID       string          `json:"typeId"`
	TypeCode     string          `json:"typeCode"`
	Name         string          `json:"name"`
	Config       json.RawMessage `json:"config,omitempty"`
	FieldMapping json.RawMessage `json:"fieldMapping,omitempty"`
	ScheduleCron string          `json:"scheduleCron,omitempty"`
	IsActive     bool            `json:"isActive"`
	LastImportAt *time.Time      `json:"lastImportAt,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

type CreateDataSourceRequest struct {
	TenantID     string          `json:"tenantId"`
	TypeID       string          `json:"typeId"`
	Name         string          `json:"name"`
	Config       json.RawMessage `json:"config,omitempty"`
	FieldMapping json.RawMessage `json:"fieldMapping,omitempty"`
	ScheduleCron string          `json:"scheduleCron"`
}

type DataSourceTypeDTO struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	ConnectorKey string `json:"connectorKey"`
	SupportsFile bool   `json:"supportsFile"`
	SupportsPull bool   `json:"supportsPull"`
}

func (s *DataSourceService) Create(ctx context.Context, req CreateDataSourceRequest) (*DataSourceDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || req.TypeID == "" {
		return nil, domain.ErrValidation
	}
	if len(req.FieldMapping) == 0 {
		return nil, domain.ErrValidation
	}
	if _, err := ingestion.ParseFieldMapping(string(req.FieldMapping)); err != nil {
		return nil, domain.ErrValidation
	}
	now := s.clock.Now()
	ds := &domain.DataSource{
		ID:           s.idGen.NewID(ctx),
		TenantID:     req.TenantID,
		TypeID:       req.TypeID,
		Name:         name,
		Config:       string(req.Config),
		FieldMapping: string(req.FieldMapping),
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

type DataSourceUpdateRequest struct {
	Name         string          `json:"name"`
	Config       json.RawMessage `json:"config,omitempty"`
	FieldMapping json.RawMessage `json:"fieldMapping,omitempty"`
	ScheduleCron string          `json:"scheduleCron"`
	IsActive     *bool           `json:"isActive"`
}

func (s *DataSourceService) Update(ctx context.Context, id string, req DataSourceUpdateRequest) (*DataSourceDTO, error) {
	ds, err := s.sources.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		ds.Name = req.Name
	}
	if len(req.Config) > 0 {
		ds.Config = string(req.Config)
	}
	if len(req.FieldMapping) > 0 {
		if _, err := ingestion.ParseFieldMapping(string(req.FieldMapping)); err != nil {
			return nil, domain.ErrValidation
		}
		ds.FieldMapping = string(req.FieldMapping)
	}
	if req.ScheduleCron != "" {
		ds.ScheduleCron = req.ScheduleCron
	}
	if req.IsActive != nil {
		ds.IsActive = *req.IsActive
	}
	ds.UpdatedAt = s.clock.Now()
	if err := s.sources.Update(ctx, ds); err != nil {
		return nil, err
	}
	d := toDataSourceDTO(*ds)
	return &d, nil
}

func (s *DataSourceService) Deactivate(ctx context.Context, id string) error {
	return s.sources.Deactivate(ctx, id, s.clock.Now())
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

type DataSourceTypeRequest struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	ConnectorKey string `json:"connectorKey"`
	SupportsFile bool   `json:"supportsFile"`
	SupportsPull bool   `json:"supportsPull"`
}

func (s *DataSourceService) CreateType(ctx context.Context, req DataSourceTypeRequest) (*DataSourceTypeDTO, error) {
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.ConnectorKey) == "" {
		return nil, domain.ErrValidation
	}
	t := &domain.DataSourceType{
		ID:           s.idGen.NewID(ctx),
		Code:         strings.TrimSpace(req.Code),
		Name:         strings.TrimSpace(req.Name),
		ConnectorKey: strings.TrimSpace(req.ConnectorKey),
		SupportsFile: req.SupportsFile,
		SupportsPull: req.SupportsPull,
		IsActive:     true,
	}
	if err := s.types.Insert(ctx, t); err != nil {
		return nil, err
	}
	return &DataSourceTypeDTO{
		ID: t.ID, Code: t.Code, Name: t.Name, ConnectorKey: t.ConnectorKey,
		SupportsFile: t.SupportsFile, SupportsPull: t.SupportsPull,
	}, nil
}

func (s *DataSourceService) UpdateType(ctx context.Context, id string, req DataSourceTypeRequest) (*DataSourceTypeDTO, error) {
	t, err := s.types.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		t.Name = req.Name
	}
	t.SupportsFile = req.SupportsFile
	t.SupportsPull = req.SupportsPull
	if err := s.types.Update(ctx, t); err != nil {
		return nil, err
	}
	return &DataSourceTypeDTO{
		ID: t.ID, Code: t.Code, Name: t.Name, ConnectorKey: t.ConnectorKey,
		SupportsFile: t.SupportsFile, SupportsPull: t.SupportsPull,
	}, nil
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

func rawJSONObject(s string) json.RawMessage {
	if strings.TrimSpace(s) == "" || s == "null" || !json.Valid([]byte(s)) {
		return nil
	}
	return json.RawMessage(s)
}

func toDataSourceDTO(ds domain.DataSource) DataSourceDTO {
	return DataSourceDTO{
		ID:           ds.ID,
		TenantID:     ds.TenantID,
		TypeID:       ds.TypeID,
		TypeCode:     ds.TypeCode,
		Name:         ds.Name,
		Config:       rawJSONObject(ds.Config),
		FieldMapping: rawJSONObject(ds.FieldMapping),
		ScheduleCron: ds.ScheduleCron,
		IsActive:     ds.IsActive,
		LastImportAt: ds.LastImportAt,
		CreatedAt:    ds.CreatedAt,
		UpdatedAt:    ds.UpdatedAt,
	}
}
