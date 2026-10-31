package service

import (
	"context"
	"strings"

	"github.com/MangriMen/UCRP/internal/domain"
)

type MallDTO struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
	FloorMin int    `json:"floorMin"`
	FloorMax int    `json:"floorMax"`
	HasZero  bool   `json:"hasZeroFloor"`
	Floors   []int  `json:"floors"`
}

type MallRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive *bool  `json:"isActive"`
	FloorMin *int   `json:"floorMin"`
	FloorMax *int   `json:"floorMax"`
	HasZero  *bool  `json:"hasZeroFloor"`
}

func (r MallRequest) applyFloors(m *domain.Mall) error {
	if r.FloorMin != nil {
		m.FloorMin = *r.FloorMin
	}
	if r.FloorMax != nil {
		m.FloorMax = *r.FloorMax
	}
	if r.HasZero != nil {
		m.HasZero = *r.HasZero
	}
	if !m.ValidFloorRange() {
		return domain.ErrValidation
	}
	return nil
}

func (s *ReferenceService) ListMalls(ctx context.Context) ([]MallDTO, error) {
	items, err := s.malls.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]MallDTO, 0, len(items))
	for _, m := range items {
		out = append(out, toMallDTO(m))
	}
	return out, nil
}

func (s *ReferenceService) CreateMall(ctx context.Context, req MallRequest) (*MallDTO, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return nil, domain.ErrValidation
	}
	m := &domain.Mall{
		ID:        s.idGen.NewID(ctx),
		Code:      code,
		Name:      name,
		IsActive:  true,
		FloorMin:  1,
		FloorMax:  1,
		CreatedAt: s.clock.Now(),
	}
	if err := req.applyFloors(m); err != nil {
		return nil, err
	}
	if err := s.malls.Insert(ctx, m); err != nil {
		return nil, err
	}
	d := toMallDTO(*m)
	return &d, nil
}

func (s *ReferenceService) UpdateMall(ctx context.Context, id string, req MallRequest) (*MallDTO, error) {
	m, err := s.malls.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		m.Name = name
	}
	if req.IsActive != nil {
		m.IsActive = *req.IsActive
	}
	if err := req.applyFloors(m); err != nil {
		return nil, err
	}
	if err := s.malls.Update(ctx, m); err != nil {
		return nil, err
	}
	d := toMallDTO(*m)
	return &d, nil
}

func (s *ReferenceService) DeleteMall(ctx context.Context, id string) error {
	return s.malls.Delete(ctx, id)
}

func toMallDTO(m domain.Mall) MallDTO {
	return MallDTO{
		ID: m.ID, Code: m.Code, Name: m.Name, IsActive: m.IsActive,
		FloorMin: m.FloorMin, FloorMax: m.FloorMax, HasZero: m.HasZero, Floors: m.Floors(),
	}
}
