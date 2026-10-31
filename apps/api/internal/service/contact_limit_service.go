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
	GuestID   string `json:"guestId"`
	Contact   string `json:"contact"`
	IDType    string `json:"idType"`
	IsControl bool   `json:"isControl"`
}

func (s *ExportService) Members(ctx context.Context, exportID string, isControl *bool) ([]ExportMemberDTO, error) {
	if _, err := s.exports.GetByID(ctx, exportID); err != nil {
		return nil, err
	}
	items, err := s.exports.ListMembers(ctx, exportID)
	if err != nil {
		return nil, err
	}
	out := make([]ExportMemberDTO, 0, len(items))
	for _, m := range items {
		if isControl != nil && m.IsControl != *isControl {
			continue
		}
		idType := "phone"
		if strings.Contains(m.Contact, "@") {
			idType = "email"
		}
		out = append(out, ExportMemberDTO{
			GuestID:   m.GuestID,
			Contact:   m.Contact,
			IDType:    idType,
			IsControl: m.IsControl,
		})
	}
	return out, nil
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
			ExportID:    l.ExportID,
			ContactedAt: l.ContactedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

type ContactDTO struct {
	ID          string `json:"id"`
	ExportID    string `json:"exportId"`
	ContactedAt string `json:"contactedAt"`
}

func (s *ProfileService) SoftDelete(ctx context.Context, guestID string) error {
	if _, err := s.guests.GetByID(ctx, guestID); err != nil {
		return err
	}
	return s.guests.SoftDelete(ctx, guestID, s.clock.Now())
}
