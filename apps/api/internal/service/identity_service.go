package service

import (
	"context"

	"github.com/MangriMen/UCRP/internal/domain"
)

type IdentityService struct {
	candidates domain.MergeCandidateRepo
	clock      domain.Clock
}

func NewIdentityService(candidates domain.MergeCandidateRepo) *IdentityService {
	return &IdentityService{candidates: candidates, clock: domain.SystemClock{}}
}

type MergeCandidateDTO struct {
	ID            string `json:"id"`
	GuestAID      string `json:"guest_a_id"`
	GuestBID      string `json:"guest_b_id"`
	MatchType     string `json:"match_type"`
	ReasonCode    string `json:"reason_code"`
	ReasonText    string `json:"reason_text"`
	Score         int    `json:"score"`
	Resolution    string `json:"resolution"`
	WinnerGuestID string `json:"winner_guest_id,omitempty"`
	MergedGuestID string `json:"merged_guest_id,omitempty"`
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
	dto := toMergeDTO(*c)
	return &dto, nil
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
