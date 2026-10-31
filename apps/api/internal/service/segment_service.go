package service

import (
	"context"
	"encoding/json"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type SegmentService struct {
	segments  domain.SegmentRepo
	evaluator domain.SegmentEvaluator
	idGen     domain.IDGen
	clock     domain.Clock
}

func NewSegmentService(repo domain.SegmentRepo, evaluator domain.SegmentEvaluator) *SegmentService {
	return &SegmentService{
		segments:  repo,
		evaluator: evaluator,
		idGen:     idgen.UUID{},
		clock:     domain.SystemClock{},
	}
}

type SegmentPreviewRequest struct {
	Was        []string `json:"was"`
	WasNot     []string `json:"was_not"`
	PeriodDays int      `json:"period_days"`
}

type SegmentPreviewResponse struct {
	Size     int      `json:"size"`
	Examples []string `json:"examples"`
}

type SegmentCreateRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	TemplateCode string   `json:"template_code"`
	Was          []string `json:"was"`
	WasNot       []string `json:"was_not"`
	PeriodDays   int      `json:"period_days"`
}

type SegmentDTO struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	TemplateCode     string          `json:"template_code"`
	CurrentVersionNo int             `json:"current_version_no"`
	Status           string          `json:"status"`
	VersionStatus    string          `json:"version_status"`
	MemberCount      *int            `json:"member_count"`
	ComputedAt       string          `json:"computed_at,omitempty"`
	Rule             json.RawMessage `json:"rule,omitempty"`
}

func (s *SegmentService) Preview(ctx context.Context, req SegmentPreviewRequest) (*SegmentPreviewResponse, error) {
	size, examples, err := s.evaluator.Preview(ctx, domain.SegmentRule{
		Was:        req.Was,
		WasNot:     req.WasNot,
		PeriodDays: req.PeriodDays,
	})
	if err != nil {
		return nil, err
	}
	return &SegmentPreviewResponse{Size: size, Examples: examples}, nil
}

func (s *SegmentService) Create(ctx context.Context, req SegmentCreateRequest, actorUserID string) (*SegmentDTO, error) {
	rule := domain.SegmentRule{Was: req.Was, WasNot: req.WasNot, PeriodDays: req.PeriodDays}
	members, err := s.evaluator.Evaluate(ctx, rule)
	if err != nil {
		return nil, err
	}
	ruleJSON, _ := json.Marshal(rule)
	now := s.clock.Now()
	seg := &domain.Segment{
		ID:               s.idGen.NewID(ctx),
		Name:             req.Name,
		Description:      req.Description,
		TemplateCode:     req.TemplateCode,
		CurrentVersionNo: 1,
		Status:           domain.SegmentActive,
		CreatedBy:        actorUserID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	ver := &domain.SegmentVersion{
		ID:        s.idGen.NewID(ctx),
		SegmentID: seg.ID,
		VersionNo: 1,
		Rule:      string(ruleJSON),
		Status:    domain.VersionPending,
		CreatedBy: actorUserID,
		CreatedAt: now,
	}
	if err := s.segments.Insert(ctx, seg, ver); err != nil {
		return nil, err
	}
	if err := s.segments.ReplaceMembers(ctx, ver.ID, members, now); err != nil {
		return nil, err
	}
	return s.Get(ctx, seg.ID)
}

func (s *SegmentService) Get(ctx context.Context, id string) (*SegmentDTO, error) {
	seg, ver, err := s.segments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := SegmentDTO{
		ID:               seg.ID,
		Name:             seg.Name,
		Description:      seg.Description,
		TemplateCode:     seg.TemplateCode,
		CurrentVersionNo: seg.CurrentVersionNo,
		Status:           string(seg.Status),
		VersionStatus:    string(ver.Status),
		MemberCount:      ver.MemberCount,
		Rule:             json.RawMessage(ver.Rule),
	}
	if ver.ComputedAt != nil {
		dto.ComputedAt = ver.ComputedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return &dto, nil
}

func (s *SegmentService) List(ctx context.Context, limit, offset int) ([]SegmentDTO, error) {
	items, err := s.segments.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]SegmentDTO, 0, len(items))
	for _, seg := range items {
		out = append(out, SegmentDTO{
			ID:               seg.ID,
			Name:             seg.Name,
			Description:      seg.Description,
			TemplateCode:     seg.TemplateCode,
			CurrentVersionNo: seg.CurrentVersionNo,
			Status:           string(seg.Status),
		})
	}
	return out, nil
}
