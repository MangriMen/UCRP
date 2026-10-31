package service

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type SegmentService struct {
	segments  domain.SegmentRepo
	evaluator domain.SegmentEvaluator
	queue     domain.Queue
	idGen     domain.IDGen
	clock     domain.Clock
}

func NewSegmentService(repo domain.SegmentRepo, evaluator domain.SegmentEvaluator, q domain.Queue) *SegmentService {
	return &SegmentService{
		segments:  repo,
		evaluator: evaluator,
		queue:     q,
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

func (s *SegmentService) Update(ctx context.Context, id string, req SegmentCreateRequest, actorUserID string) (*SegmentDTO, error) {
	seg, _, err := s.segments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if seg.Status == domain.SegmentArchived {
		return nil, domain.ErrConflict
	}
	rule := domain.SegmentRule{Was: req.Was, WasNot: req.WasNot, PeriodDays: req.PeriodDays}
	ruleJSON, _ := json.Marshal(rule)
	now := s.clock.Now()
	ver := &domain.SegmentVersion{
		ID:        s.idGen.NewID(ctx),
		SegmentID: seg.ID,
		VersionNo: seg.CurrentVersionNo + 1,
		Rule:      string(ruleJSON),
		Status:    domain.VersionPending,
		CreatedBy: actorUserID,
		CreatedAt: now,
	}
	if err := s.segments.InsertVersion(ctx, ver); err != nil {
		return nil, err
	}
	if err := s.queue.Enqueue(ctx, domain.Job{
		JobType:  "segment.recompute",
		Payload:  `{"segment_id":"` + seg.ID + `","version_id":"` + ver.ID + `"}`,
		DedupKey: "segment:" + seg.ID + ":" + strconv.Itoa(ver.VersionNo),
	}); err != nil {
		return nil, err
	}
	return s.Get(ctx, seg.ID)
}

func (s *SegmentService) Delete(ctx context.Context, id string) error {
	seg, _, err := s.segments.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.segments.Archive(ctx, seg.ID, s.clock.Now())
}

func (s *SegmentService) Members(ctx context.Context, id string, versionNo, limit, offset int) ([]string, error) {
	var (
		seg *domain.Segment
		ver *domain.SegmentVersion
		err error
	)
	if versionNo > 0 {
		seg, ver, err = s.segments.GetVersion(ctx, id, versionNo)
	} else {
		seg, ver, err = s.segments.GetByID(ctx, id)
	}
	if err != nil {
		return nil, err
	}
	_ = seg
	return s.segments.ListMembers(ctx, ver.ID, limit, offset)
}

func (s *SegmentService) Versions(ctx context.Context, id string) ([]SegmentVersionDTO, error) {
	if _, _, err := s.segments.GetByID(ctx, id); err != nil {
		return nil, err
	}
	items, err := s.segments.ListVersions(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]SegmentVersionDTO, 0, len(items))
	for _, v := range items {
		d := SegmentVersionDTO{
			ID:          v.ID,
			VersionNo:   v.VersionNo,
			Rule:        json.RawMessage(v.Rule),
			Status:      string(v.Status),
			MemberCount: v.MemberCount,
			CreatedAt:   v.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		}
		if v.ComputedAt != nil {
			d.ComputedAt = v.ComputedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		}
		out = append(out, d)
	}
	return out, nil
}

type SegmentVersionDTO struct {
	ID          string          `json:"id"`
	VersionNo   int             `json:"version_no"`
	Rule        json.RawMessage `json:"rule"`
	Status      string          `json:"status"`
	MemberCount *int            `json:"member_count"`
	ComputedAt  string          `json:"computed_at,omitempty"`
	CreatedAt   string          `json:"created_at"`
}

func (s *SegmentService) Recompute(ctx context.Context, id string) (*SegmentDTO, error) {
	seg, ver, err := s.segments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ver.Status == domain.VersionComputing || ver.Status == domain.VersionPending {
		if err := s.segments.MarkComputing(ctx, ver.ID); err != nil {
			return nil, err
		}
	}
	if err := s.queue.Enqueue(ctx, domain.Job{
		JobType:  "segment.recompute",
		Payload:  `{"segment_id":"` + seg.ID + `","version_id":"` + ver.ID + `"}`,
		DedupKey: "segment:" + seg.ID + ":" + strconv.Itoa(ver.VersionNo),
	}); err != nil {
		return nil, err
	}
	return s.Get(ctx, seg.ID)
}

func (s *SegmentService) RecomputeVersion(ctx context.Context, segmentID, versionID string) error {
	_, ver, err := s.segments.GetByID(ctx, segmentID)
	if err != nil {
		return err
	}
	if versionID != "" && versionID != ver.ID {
		return domain.ErrNotFound
	}
	var rule domain.SegmentRule
	if err := json.Unmarshal([]byte(ver.Rule), &rule); err != nil {
		return err
	}
	members, err := s.evaluator.Evaluate(ctx, rule)
	if err != nil {
		return err
	}
	return s.segments.ReplaceMembers(ctx, ver.ID, members, s.clock.Now())
}

func (s *SegmentService) Templates(ctx context.Context) ([]SegmentDTO, error) {
	items, err := s.segments.ListTemplates(ctx)
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
