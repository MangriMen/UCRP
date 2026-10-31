package service

import (
	"context"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type IdentityService struct {
	candidates  domain.MergeCandidateRepo
	blocks      domain.UnmergeBlockRepo
	guests      domain.GuestRepo
	identifiers domain.GuestIdentifierRepo
	events      domain.EventRepo
	idGen       domain.IDGen
	clock       domain.Clock
}

func NewIdentityService(
	candidates domain.MergeCandidateRepo,
	blocks domain.UnmergeBlockRepo,
	guests domain.GuestRepo,
	identifiers domain.GuestIdentifierRepo,
	events domain.EventRepo,
) *IdentityService {
	return &IdentityService{
		candidates:  candidates,
		blocks:      blocks,
		guests:      guests,
		identifiers: identifiers,
		events:      events,
		idGen:       idgen.UUID{},
		clock:       domain.SystemClock{},
	}
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
	c, err := s.candidates.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Resolution != domain.MergePending {
		return nil, domain.ErrConflict
	}
	blocked, err := s.blocks.HasBlock(ctx, c.GuestAID, c.GuestBID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, domain.ErrConflict
	}
	if err := s.guests.MarkMerged(ctx, c.GuestBID, c.GuestAID); err != nil {
		return nil, err
	}
	if err := s.events.MoveToGuest(ctx, c.GuestBID, c.GuestAID); err != nil {
		return nil, err
	}
	if err := s.identifiers.MoveToGuest(ctx, c.GuestBID, c.GuestAID); err != nil {
		return nil, err
	}
	now := s.clock.Now()
	c.Resolution = domain.MergeConfirmed
	c.WinnerGuestID = c.GuestAID
	c.MergedGuestID = c.GuestBID
	c.DecidedBy = actorUserID
	c.DecidedAt = &now
	if err := s.candidates.UpdateDecision(ctx, c); err != nil {
		return nil, err
	}
	dto := toMergeDTO(*c)
	return &dto, nil
}

func (s *IdentityService) Reject(ctx context.Context, id, actorUserID string) (*MergeCandidateDTO, error) {
	c, err := s.candidates.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Resolution != domain.MergePending {
		return nil, domain.ErrConflict
	}
	if err := s.createBlock(ctx, c, "rejected", actorUserID); err != nil {
		return nil, err
	}
	now := s.clock.Now()
	c.Resolution = domain.MergeRejected
	c.DecidedBy = actorUserID
	c.DecidedAt = &now
	if err := s.candidates.UpdateDecision(ctx, c); err != nil {
		return nil, err
	}
	dto := toMergeDTO(*c)
	return &dto, nil
}

func (s *IdentityService) Revert(ctx context.Context, id, actorUserID string) (*MergeCandidateDTO, error) {
	c, err := s.candidates.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Resolution != domain.MergeConfirmed && c.Resolution != domain.MergeAutoMerged {
		return nil, domain.ErrConflict
	}
	if c.MergedGuestID == "" {
		return nil, domain.ErrConflict
	}
	if err := s.events.RestoreOrigin(ctx, c.MergedGuestID); err != nil {
		return nil, err
	}
	if err := s.identifiers.RestoreOrigin(ctx, c.MergedGuestID); err != nil {
		return nil, err
	}
	if err := s.guests.MarkActive(ctx, c.MergedGuestID); err != nil {
		return nil, err
	}
	if err := s.createBlock(ctx, c, "manual_unmerge", actorUserID); err != nil {
		return nil, err
	}
	now := s.clock.Now()
	c.Resolution = domain.MergeReverted
	c.DecidedBy = actorUserID
	c.DecidedAt = &now
	if err := s.candidates.UpdateDecision(ctx, c); err != nil {
		return nil, err
	}
	dto := toMergeDTO(*c)
	return &dto, nil
}

func (s *IdentityService) createBlock(ctx context.Context, c *domain.MergeCandidate, reason, actorUserID string) error {
	a, b := c.GuestAID, c.GuestBID
	if a > b {
		a, b = b, a
	}
	blocked, err := s.blocks.HasBlock(ctx, a, b)
	if err != nil {
		return err
	}
	if blocked {
		return nil
	}
	return s.blocks.Insert(ctx, &domain.UnmergeBlock{
		ID:          s.idGen.NewID(ctx),
		GuestAID:    a,
		GuestBID:    b,
		Reason:      reason,
		CandidateID: c.ID,
		CreatedBy:   actorUserID,
		CreatedAt:   s.clock.Now(),
	})
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
