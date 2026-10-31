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

type SegmentRuleBody struct {
	Was            []string `json:"was"`
	WasNot         []string `json:"wasNot"`
	PeriodDays     int      `json:"periodDays"`
	RequireContact string   `json:"requireContact"`
}

type SegmentPreviewRequest struct {
	Rule           *SegmentRuleBody `json:"rule"`
	Was            []string         `json:"was"`
	WasNot         []string         `json:"was_not"`
	PeriodDays     int              `json:"period_days"`
	RequireContact string           `json:"require_contact"`
	Limit          int              `json:"limit"`
}

func (r SegmentPreviewRequest) toRule() domain.SegmentRule {
	if r.Rule != nil {
		return domain.SegmentRule{
			Was:            r.Rule.Was,
			WasNot:         r.Rule.WasNot,
			PeriodDays:     r.Rule.PeriodDays,
			RequireContact: r.Rule.RequireContact,
		}
	}
	return domain.SegmentRule{
		Was:            r.Was,
		WasNot:         r.WasNot,
		PeriodDays:     r.PeriodDays,
		RequireContact: r.RequireContact,
	}
}

type SegmentPreviewExample struct {
	GuestID string `json:"guestId"`
}

type SegmentPreviewResponse struct {
	Size       *int                    `json:"size"`
	Examples   []SegmentPreviewExample `json:"examples"`
	Cached     bool                    `json:"cached"`
	Suppressed bool                    `json:"suppressed"`
}

type SegmentCreateRequest struct {
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	TemplateCode   string           `json:"templateCode"`
	Rule           *SegmentRuleBody `json:"rule"`
	Was            []string         `json:"was"`
	WasNot         []string         `json:"was_not"`
	PeriodDays     int              `json:"period_days"`
	RequireContact string           `json:"require_contact"`
}

func (r SegmentCreateRequest) toRule() domain.SegmentRule {
	if r.Rule != nil {
		return domain.SegmentRule{
			Was:            r.Rule.Was,
			WasNot:         r.Rule.WasNot,
			PeriodDays:     r.Rule.PeriodDays,
			RequireContact: r.Rule.RequireContact,
		}
	}
	return domain.SegmentRule{
		Was:            r.Was,
		WasNot:         r.WasNot,
		PeriodDays:     r.PeriodDays,
		RequireContact: r.RequireContact,
	}
}

type SegmentDTO struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	TemplateCode     string          `json:"templateCode"`
	IsTemplate       bool            `json:"isTemplate"`
	CurrentVersionNo int             `json:"currentVersionNo"`
	Status           string          `json:"status"`
	VersionStatus    string          `json:"versionStatus,omitempty"`
	MemberCount      *int            `json:"memberCount"`
	Suppressed       bool            `json:"suppressed"`
	ComputedAt       string          `json:"computedAt,omitempty"`
	Rule             json.RawMessage `json:"rule,omitempty"`
	CreatedBy        string          `json:"createdBy,omitempty"`
	CreatedAt        string          `json:"createdAt"`
	UpdatedAt        string          `json:"updatedAt"`
}

type SegmentPage struct {
	Items []SegmentDTO `json:"items"`
	Meta  PageMeta     `json:"meta"`
}

func (s *SegmentService) Preview(ctx context.Context, req SegmentPreviewRequest) (*SegmentPreviewResponse, error) {
	size, examples, err := s.evaluator.Preview(ctx, req.toRule())
	if err != nil {
		return nil, err
	}
	out := make([]SegmentPreviewExample, 0, len(examples))
	count, suppressed := visibleSegmentCount(&size)
	if !suppressed {
		for _, id := range examples {
			out = append(out, SegmentPreviewExample{GuestID: id})
		}
	}
	return &SegmentPreviewResponse{Size: count, Examples: out, Cached: false, Suppressed: suppressed}, nil
}

func visibleSegmentCount(count *int) (*int, bool) {
	if count == nil || *count < 5 {
		return nil, true
	}
	return count, false
}

func (s *SegmentService) Create(ctx context.Context, req SegmentCreateRequest, actorUserID string) (*SegmentDTO, error) {
	rule := req.toRule()
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
		Status:    domain.VersionReady,
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
	return toSegmentDTO(seg, ver), nil
}

func toSegmentDTO(seg *domain.Segment, ver *domain.SegmentVersion) *SegmentDTO {
	count, suppressed := visibleSegmentCount(ver.MemberCount)
	dto := &SegmentDTO{
		ID:               seg.ID,
		Name:             seg.Name,
		Description:      seg.Description,
		TemplateCode:     seg.TemplateCode,
		IsTemplate:       seg.IsTemplate,
		CurrentVersionNo: seg.CurrentVersionNo,
		Status:           string(seg.Status),
		VersionStatus:    string(ver.Status),
		MemberCount:      count,
		Suppressed:       suppressed,
		Rule:             json.RawMessage(ver.Rule),
		CreatedBy:        seg.CreatedBy,
		CreatedAt:        seg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        seg.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if ver.ComputedAt != nil {
		dto.ComputedAt = ver.ComputedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return dto
}

func (s *SegmentService) List(ctx context.Context, limit, offset int) (*SegmentPage, error) {
	items, err := s.segments.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	out := make([]SegmentDTO, 0, len(items))
	for _, seg := range items {
		out = append(out, SegmentDTO{
			ID:               seg.ID,
			Name:             seg.Name,
			Description:      seg.Description,
			TemplateCode:     seg.TemplateCode,
			IsTemplate:       seg.IsTemplate,
			CurrentVersionNo: seg.CurrentVersionNo,
			Status:           string(seg.Status),
			Suppressed:       true,
			CreatedAt:        seg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        seg.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	total := offset + len(out)
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	return &SegmentPage{
		Items: out,
		Meta: PageMeta{
			Page:       offset/limit + 1,
			PageSize:   limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *SegmentService) Update(ctx context.Context, id string, req SegmentCreateRequest, actorUserID string) (*SegmentDTO, error) {
	seg, _, err := s.segments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if seg.Status == domain.SegmentArchived {
		return nil, domain.ErrConflict
	}
	rule := req.toRule()
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
		count, suppressed := visibleSegmentCount(v.MemberCount)
		d := SegmentVersionDTO{
			ID:          v.ID,
			VersionNo:   v.VersionNo,
			Rule:        json.RawMessage(v.Rule),
			Status:      string(v.Status),
			MemberCount: count,
			Suppressed:  suppressed,
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
	VersionNo   int             `json:"versionNo"`
	Rule        json.RawMessage `json:"rule"`
	Status      string          `json:"status"`
	MemberCount *int            `json:"memberCount"`
	Suppressed  bool            `json:"suppressed"`
	ComputedAt  string          `json:"computedAt,omitempty"`
	CreatedAt   string          `json:"createdAt"`
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
			IsTemplate:       seg.IsTemplate,
			CurrentVersionNo: seg.CurrentVersionNo,
			Status:           string(seg.Status),
			Suppressed:       true,
			CreatedAt:        seg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        seg.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return out, nil
}
