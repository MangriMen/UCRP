package service

import (
	"context"
	"encoding/json"

	"github.com/MangriMen/UCRP/internal/domain"
)

type IdentityService struct {
	candidates domain.MergeCandidateRepo
	blocks     domain.MergeBlockRepo
	manual     domain.MergeManualRepo
	queue      domain.Queue
	clock      domain.Clock
}

func NewIdentityService(candidates domain.MergeCandidateRepo, blocks domain.MergeBlockRepo, manual domain.MergeManualRepo, queue domain.Queue) *IdentityService {
	return &IdentityService{
		candidates: candidates,
		blocks:     blocks,
		manual:     manual,
		queue:      queue,
		clock:      domain.SystemClock{},
	}
}

type MergeCandidateDTO struct {
	ID            string `json:"id"`
	GuestAID      string `json:"guestAId"`
	GuestBID      string `json:"guestBId"`
	MatchType     string `json:"matchType"`
	ReasonCode    string `json:"reasonCode"`
	ReasonText    string `json:"reasonText"`
	Score         int    `json:"score"`
	Resolution    string `json:"resolution"`
	WinnerGuestID string `json:"winnerGuestId,omitempty"`
	MergedGuestID string `json:"mergedGuestId,omitempty"`
}

func (s *IdentityService) List(ctx context.Context, resolution string, limit, offset int) ([]MergeCandidateDTO, error) {
	items, err := s.candidates.List(ctx, domain.MergeResolution(resolution), limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]MergeCandidateDTO, 0, len(items))
	for _, c := range items {
		out = append(out, toMergeDTO(c))
	}
	return out, nil
}

func (s *IdentityService) Get(ctx context.Context, id string) (*MergeCandidateDTO, error) {
	c, err := s.candidates.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toMergeDTO(*c)
	return &dto, nil
}

func (s *IdentityService) Confirm(ctx context.Context, id, actorUserID string) (*MergeCandidateDTO, error) {
	c, err := s.candidates.Confirm(ctx, id, actorUserID, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.enqueueRecommendRefresh(ctx, c.GuestAID, c.GuestBID); err != nil {
		return nil, err
	}
	dto := toMergeDTO(*c)
	return &dto, nil
}

func (s *IdentityService) Reject(ctx context.Context, id, actorUserID string) (*MergeCandidateDTO, error) {
	c, err := s.candidates.Reject(ctx, id, actorUserID, s.clock.Now())
	if err != nil {
		return nil, err
	}
	dto := toMergeDTO(*c)
	return &dto, nil
}

func (s *IdentityService) Revert(ctx context.Context, id, actorUserID string) (*MergeCandidateDTO, error) {
	c, err := s.candidates.Revert(ctx, id, actorUserID, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.enqueueRecommendRefresh(ctx, c.WinnerGuestID, c.MergedGuestID); err != nil {
		return nil, err
	}
	dto := toMergeDTO(*c)
	return &dto, nil
}

func (s *IdentityService) Merge(ctx context.Context, sourceGuestID, targetGuestID, actorUserID string) (*MergeCandidateDTO, error) {
	c, err := s.manual.ManualMerge(ctx, sourceGuestID, targetGuestID, actorUserID, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.enqueueRecommendRefresh(ctx, c.WinnerGuestID, c.MergedGuestID); err != nil {
		return nil, err
	}
	d := toMergeDTO(*c)
	return &d, nil
}

func (s *IdentityService) enqueueRecommendRefresh(ctx context.Context, guestIDs ...string) error {
	if s.queue == nil {
		return nil
	}
	for _, guestID := range guestIDs {
		if guestID == "" {
			continue
		}
		payload, err := json.Marshal(map[string]string{"guest_id": guestID})
		if err != nil {
			return err
		}
		if err := s.queue.Enqueue(ctx, domain.Job{
			JobType:  "recommend.refresh",
			Payload:  string(payload),
			DedupKey: "reco:" + guestID,
		}); err != nil {
			return err
		}
	}
	return nil
}

type MergeBlockDTO struct {
	ID        string `json:"id"`
	GuestAID  string `json:"guestAId"`
	GuestBID  string `json:"guestBId"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"createdAt"`
}

func (s *IdentityService) ListBlocks(ctx context.Context, limit, offset int) ([]MergeBlockDTO, error) {
	items, err := s.blocks.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]MergeBlockDTO, 0, len(items))
	for _, b := range items {
		out = append(out, MergeBlockDTO{
			ID:        b.ID,
			GuestAID:  b.GuestAID,
			GuestBID:  b.GuestBID,
			Reason:    b.Reason,
			CreatedAt: b.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return out, nil
}

func (s *IdentityService) DeleteBlock(ctx context.Context, id string) error {
	if _, err := s.blocks.GetByID(ctx, id); err != nil {
		return err
	}
	return s.blocks.Delete(ctx, id)
}

func toMergeDTO(c domain.MergeCandidate) MergeCandidateDTO {
	return MergeCandidateDTO{
		ID:            c.ID,
		GuestAID:      c.GuestAID,
		GuestBID:      c.GuestBID,
		MatchType:     c.MatchType,
		ReasonCode:    c.ReasonCode,
		ReasonText:    c.ReasonText,
		Score:         c.Score,
		Resolution:    string(c.Resolution),
		WinnerGuestID: c.WinnerGuestID,
		MergedGuestID: c.MergedGuestID,
	}
}
