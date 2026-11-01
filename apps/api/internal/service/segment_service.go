package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type SegmentService struct {
	segments  domain.SegmentRepo
	evaluator domain.SegmentEvaluator
	tenants   domain.TenantRepo
	queue     domain.Queue
	idGen     domain.IDGen
	clock     domain.Clock
}

func NewSegmentService(repo domain.SegmentRepo, evaluator domain.SegmentEvaluator, tenants domain.TenantRepo, q domain.Queue) *SegmentService {
	return &SegmentService{
		segments:  repo,
		evaluator: evaluator,
		tenants:   tenants,
		queue:     q,
		idGen:     idgen.UUID{},
		clock:     domain.SystemClock{},
	}
}

type SegmentPreviewRequest struct {
	Rule              *domain.SegmentRule `json:"rule"`
	VisitedTenantIDs  []string            `json:"visitedTenantIds"`
	ExcludedTenantIDs []string            `json:"excludedTenantIds"`
	PeriodDays        int                 `json:"periodDays"`
	RequireContact    string              `json:"requireContact"`
	Limit             int                 `json:"limit"`
}

func (r SegmentPreviewRequest) toRule() domain.SegmentRule {
	return ruleOrFlat(r.Rule, r.VisitedTenantIDs, r.ExcludedTenantIDs, r.PeriodDays, r.RequireContact)
}

func ruleOrFlat(rule *domain.SegmentRule, visited, excluded []string, periodDays int, requireContact string) domain.SegmentRule {
	if rule != nil {
		return *rule
	}
	return domain.SegmentRule{
		VisitedTenantIDs:  visited,
		ExcludedTenantIDs: excluded,
		PeriodDays:        periodDays,
		RequireContact:    requireContact,
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
	Name              string              `json:"name"`
	Description       string              `json:"description"`
	TemplateCode      string              `json:"templateCode"`
	Rule              *domain.SegmentRule `json:"rule"`
	VisitedTenantIDs  []string            `json:"visitedTenantIds"`
	ExcludedTenantIDs []string            `json:"excludedTenantIds"`
	PeriodDays        int                 `json:"periodDays"`
	RequireContact    string              `json:"requireContact"`
}

func (r SegmentCreateRequest) toRule() domain.SegmentRule {
	return ruleOrFlat(r.Rule, r.VisitedTenantIDs, r.ExcludedTenantIDs, r.PeriodDays, r.RequireContact)
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
	RuleText         string          `json:"ruleText,omitempty"`
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
	for _, id := range examples {
		out = append(out, SegmentPreviewExample{GuestID: id})
	}
	res := &SegmentPreviewResponse{Examples: out, Cached: false}
	if size < 5 {
		res.Suppressed = true
		res.Examples = []SegmentPreviewExample{}
		return res, nil
	}
	res.Size = &size
	return res, nil
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
	texts, err := s.ruleTexts(ctx, []string{ver.Rule})
	if err != nil {
		return nil, err
	}
	dto := toSegmentDTO(seg, ver)
	dto.RuleText = texts[0]
	return dto, nil
}

func toSegmentDTO(seg *domain.Segment, ver *domain.SegmentVersion) *SegmentDTO {
	dto := &SegmentDTO{
		ID:               seg.ID,
		Name:             seg.Name,
		Description:      seg.Description,
		TemplateCode:     seg.TemplateCode,
		IsTemplate:       seg.IsTemplate,
		CurrentVersionNo: seg.CurrentVersionNo,
		CreatedBy:        seg.CreatedBy,
		CreatedAt:        seg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        seg.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	switch {
	case seg.Status == domain.SegmentArchived:
		dto.Status = string(domain.SegmentArchived)
	case ver != nil:
		dto.Status = string(ver.Status)
	default:
		dto.Status = string(domain.VersionPending)
	}
	if ver == nil {
		dto.Suppressed = true
		return dto
	}
	dto.VersionStatus = string(ver.Status)
	dto.Rule = storedRuleJSON(ver.Rule)
	if ver.MemberCount != nil {
		dto.MemberCount = ver.MemberCount
	}
	if ver.MemberCount == nil || *ver.MemberCount < 5 {
		dto.Suppressed = true
		dto.MemberCount = nil
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
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	versions, err := s.currentVersions(ctx, items)
	if err != nil {
		return nil, err
	}
	rawRules := make([]string, 0, len(items))
	for _, seg := range items {
		if v := versions[seg.ID]; v != nil {
			rawRules = append(rawRules, v.Rule)
		} else {
			rawRules = append(rawRules, "")
		}
	}
	texts, err := s.ruleTexts(ctx, rawRules)
	if err != nil {
		return nil, err
	}
	out := make([]SegmentDTO, 0, len(items))
	for i, seg := range items {
		dto := toSegmentDTO(&seg, versions[seg.ID])
		dto.RuleText = texts[i]
		out = append(out, *dto)
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

type SegmentMemberView struct {
	GuestID     string     `json:"guestId"`
	DisplayName *string    `json:"displayName,omitempty"`
	TenantCount int        `json:"tenantCount"`
	LastSeenAt  *time.Time `json:"lastSeenAt"`
	AddedAt     time.Time  `json:"addedAt"`
}

type SegmentMemberPage struct {
	Items []SegmentMemberView `json:"items"`
	Meta  PageMeta            `json:"meta"`
}

func (s *SegmentService) Members(ctx context.Context, id string, versionNo int, showNames bool, limit, offset int) (*SegmentMemberPage, error) {
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
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	members, err := s.segments.ListMembers(ctx, ver.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.segments.CountMembers(ctx, ver.ID)
	if err != nil {
		return nil, err
	}
	items := make([]SegmentMemberView, 0, len(members))
	for _, member := range members {
		view := SegmentMemberView{GuestID: member.GuestID, TenantCount: member.TenantCount, LastSeenAt: member.LastSeenAt, AddedAt: member.AddedAt.UTC()}
		if showNames && member.DisplayName != "" {
			view.DisplayName = &member.DisplayName
		}
		items = append(items, view)
	}
	if offset < 0 {
		offset = 0
	}
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	return &SegmentMemberPage{
		Items: items,
		Meta: PageMeta{
			Page:       offset/limit + 1,
			PageSize:   limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
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
	rawRules := make([]string, 0, len(items))
	for _, v := range items {
		rawRules = append(rawRules, v.Rule)
	}
	texts, err := s.ruleTexts(ctx, rawRules)
	if err != nil {
		return nil, err
	}
	for i, v := range items {
		d := SegmentVersionDTO{
			ID:          v.ID,
			SegmentID:   id,
			VersionNo:   v.VersionNo,
			Rule:        storedRuleJSON(v.Rule),
			RuleText:    texts[i],
			Status:      string(v.Status),
			MemberCount: v.MemberCount,
			CreatedAt:   v.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		}
		if v.MemberCount == nil || *v.MemberCount < 5 {
			d.Suppressed = true
			d.MemberCount = nil
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
	SegmentID   string          `json:"segmentId"`
	VersionNo   int             `json:"versionNo"`
	Rule        json.RawMessage `json:"rule"`
	RuleText    string          `json:"ruleText,omitempty"`
	Status      string          `json:"status"`
	MemberCount *int            `json:"memberCount"`
	Suppressed  bool            `json:"suppressed"`
	ComputedAt  string          `json:"computedAt,omitempty"`
	CreatedAt   string          `json:"createdAt"`
}

type RecomputeAccepted struct {
	JobID   string `json:"jobId"`
	JobType string `json:"jobType"`
	Status  string `json:"status"`
}

func (s *SegmentService) Recompute(ctx context.Context, id string) (*RecomputeAccepted, error) {
	seg, ver, err := s.segments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ver.Status == domain.VersionComputing || ver.Status == domain.VersionPending {
		if err := s.segments.MarkComputing(ctx, ver.ID); err != nil {
			return nil, err
		}
	}
	job := domain.Job{
		ID:       s.idGen.NewID(ctx),
		JobType:  "segment.recompute",
		Payload:  `{"segment_id":"` + seg.ID + `","version_id":"` + ver.ID + `"}`,
		DedupKey: "segment:" + seg.ID + ":" + strconv.Itoa(ver.VersionNo),
	}
	if err := s.queue.Enqueue(ctx, job); err != nil {
		return nil, err
	}
	return &RecomputeAccepted{JobID: job.ID, JobType: job.JobType, Status: string(domain.JobQueued)}, nil
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

type SegmentTemplateView struct {
	Code        string          `json:"code"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Rule        json.RawMessage `json:"rule"`
}

func (s *SegmentService) Templates(ctx context.Context) ([]SegmentTemplateView, error) {
	items, err := s.segments.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}
	versions, err := s.currentVersions(ctx, items)
	if err != nil {
		return nil, err
	}
	out := make([]SegmentTemplateView, 0, len(items))
	for _, seg := range items {
		view := SegmentTemplateView{
			Code:        seg.TemplateCode,
			Name:        seg.Name,
			Description: seg.Description,
			Rule:        json.RawMessage("{}"),
		}
		if ver := versions[seg.ID]; ver != nil {
			view.Rule = storedRuleJSON(ver.Rule)
		}
		out = append(out, view)
	}
	return out, nil
}

func (s *SegmentService) currentVersions(ctx context.Context, segments []domain.Segment) (map[string]*domain.SegmentVersion, error) {
	ids := make([]string, 0, len(segments))
	for _, seg := range segments {
		ids = append(ids, seg.ID)
	}
	return s.segments.CurrentVersions(ctx, ids)
}

func storedRuleJSON(stored string) json.RawMessage {
	var rule domain.SegmentRule
	if err := json.Unmarshal([]byte(stored), &rule); err != nil {
		return json.RawMessage(stored)
	}
	normalized, err := json.Marshal(rule)
	if err != nil {
		return json.RawMessage(stored)
	}
	return normalized
}
