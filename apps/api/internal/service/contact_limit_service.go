package service

import (
	"context"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type ContactLimitService struct {
	limits domain.ContactLimitRepo
	clock  domain.Clock
	idGen  domain.IDGen
}

func NewContactLimitService(limits domain.ContactLimitRepo) *ContactLimitService {
	return &ContactLimitService{limits: limits, clock: domain.SystemClock{}, idGen: idgen.UUID{}}
}

type ContactLimitDTO struct {
	ID          string `json:"id"`
	Channel     string `json:"channel"`
	MaxContacts int    `json:"maxContacts"`
	PeriodDays  int    `json:"periodDays"`
	IsActive    bool   `json:"isActive"`
}

type ContactLimitRequest struct {
	Channel     string `json:"channel"`
	MaxContacts int    `json:"maxContacts"`
	PeriodDays  int    `json:"periodDays"`
	IsActive    *bool  `json:"isActive"`
}

func (s *ContactLimitService) List(ctx context.Context, limit, offset int) ([]ContactLimitDTO, error) {
	items, err := s.limits.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return toLimitDTOs(items), nil
}

func (s *ContactLimitService) Create(ctx context.Context, req ContactLimitRequest, actorUserID string) (*ContactLimitDTO, error) {
	channel := strings.ToLower(strings.TrimSpace(req.Channel))
	if channel != "" && channel != "sms" && channel != "push" && channel != "email" {
		return nil, domain.ErrValidation
	}
	if req.MaxContacts < 1 || req.PeriodDays < 1 {
		return nil, domain.ErrValidation
	}
	now := s.clock.Now()
	c := &domain.ContactLimit{
		ID:          s.idGen.NewID(ctx),
		Channel:     channel,
		MaxContacts: req.MaxContacts,
		PeriodDays:  req.PeriodDays,
		IsActive:    true,
		CreatedBy:   actorUserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.limits.Insert(ctx, c); err != nil {
		return nil, err
	}
	d := toLimitDTO(*c)
	return &d, nil
}

func (s *ContactLimitService) Update(ctx context.Context, id string, req ContactLimitRequest) (*ContactLimitDTO, error) {
	c, err := s.limits.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	channel := strings.ToLower(strings.TrimSpace(req.Channel))
	if channel != "" && channel != "sms" && channel != "push" && channel != "email" {
		return nil, domain.ErrValidation
	}
	if req.MaxContacts < 1 || req.PeriodDays < 1 {
		return nil, domain.ErrValidation
	}
	c.Channel = channel
	c.MaxContacts = req.MaxContacts
	c.PeriodDays = req.PeriodDays
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}
	c.UpdatedAt = s.clock.Now()
	if err := s.limits.Update(ctx, c); err != nil {
		return nil, err
	}
	d := toLimitDTO(*c)
	return &d, nil
}

func (s *ContactLimitService) Delete(ctx context.Context, id string) error {
	return s.limits.Delete(ctx, id)
}

func toLimitDTOs(items []domain.ContactLimit) []ContactLimitDTO {
	out := make([]ContactLimitDTO, 0, len(items))
	for _, c := range items {
		out = append(out, toLimitDTO(c))
	}
	return out
}

func toLimitDTO(c domain.ContactLimit) ContactLimitDTO {
	return ContactLimitDTO{
		ID:          c.ID,
		Channel:     c.Channel,
		MaxContacts: c.MaxContacts,
		PeriodDays:  c.PeriodDays,
		IsActive:    c.IsActive,
	}
}

type ExportMemberDTO struct {
	GuestID      string                  `json:"guestId"`
	DisplayName  *string                 `json:"displayName,omitempty"`
	IsControl    bool                    `json:"isControl"`
	IdentifierID string                  `json:"identifierId,omitempty"`
	Contact      *ExportMemberContactDTO `json:"contact,omitempty"`
}

type ExportMemberContactDTO struct {
	IDType string `json:"idType"`
	Value  string `json:"value"`
}

func (s *ExportService) Members(ctx context.Context, exportID string, isControl *bool, showNames bool, limit, offset int) ([]ExportMemberDTO, int, error) {
	if _, err := s.exports.GetByID(ctx, exportID); err != nil {
		return nil, 0, err
	}
	items, total, err := s.exports.ListMemberPage(ctx, exportID, isControl, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ExportMemberDTO, 0, len(items))
	for _, m := range items {
		member := ExportMemberDTO{GuestID: m.GuestID, IsControl: m.IsControl, IdentifierID: m.IdentifierID}
		if m.Contact != "" {
			member.Contact = &ExportMemberContactDTO{IDType: m.ContactType, Value: m.Contact}
		}
		if showNames && m.DisplayName != "" {
			member.DisplayName = &m.DisplayName
		}
		out = append(out, member)
	}
	return out, total, nil
}

func (s *ProfileService) Contacts(ctx context.Context, guestID string, limit, offset int) ([]ContactDTO, error) {
	if _, err := s.guests.GetByID(ctx, guestID); err != nil {
		return nil, err
	}
	logs, err := s.contacts.ListByGuest(ctx, guestID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]ContactDTO, 0, len(logs))
	for _, l := range logs {
		out = append(out, ContactDTO{
			ID:          l.ID,
			GuestID:     l.GuestID,
			ExportID:    l.ExportID,
			Channel:     l.Channel,
			ContactedAt: l.ContactedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

type ContactDTO struct {
	ID          string `json:"id"`
	GuestID     string `json:"guestId"`
	ExportID    string `json:"exportId"`
	Channel     string `json:"channel"`
	ContactedAt string `json:"contactedAt"`
}

func (s *ProfileService) SoftDelete(ctx context.Context, guestID string) error {
	if _, err := s.guests.GetByID(ctx, guestID); err != nil {
		return err
	}
	return s.guests.SoftDelete(ctx, guestID, s.clock.Now())
}
