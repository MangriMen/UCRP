package service

import (
	"context"

	"github.com/MangriMen/UCRP/internal/domain"
)

type ProfileService struct {
	guests      domain.GuestRepo
	identifiers domain.GuestIdentifierRepo
	events      domain.EventRepo
}

func NewProfileService(guests domain.GuestRepo, identifiers domain.GuestIdentifierRepo, events domain.EventRepo) *ProfileService {
	return &ProfileService{guests: guests, identifiers: identifiers, events: events}
}

type GuestListItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
}

type GuestProfile struct {
	ID           string               `json:"id"`
	DisplayName  string               `json:"display_name"`
	Status       string               `json:"status"`
	MergedIntoID string               `json:"merged_into_id"`
	TenantCount  int                  `json:"tenant_count"`
	Tenants      []string             `json:"tenants"`
	Identifiers  []GuestIdentifierDTO `json:"identifiers"`
}

type GuestIdentifierDTO struct {
	IDType string `json:"id_type"`
	Value  string `json:"value"`
}

type EventDTO struct {
	ID          string `json:"id"`
	TypeID      string `json:"event_type_id"`
	TenantID    string `json:"tenant_id"`
	ProjectID   string `json:"project_id"`
	ExternalID  string `json:"external_id"`
	OccurredAt  string `json:"occurred_at"`
	AmountMinor *int64 `json:"amount_minor"`
	Currency    string `json:"currency"`
}

func (s *ProfileService) Search(ctx context.Context, q string, limit, offset int) ([]GuestListItem, error) {
	guests, err := s.guests.Search(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]GuestListItem, 0, len(guests))
	for _, g := range guests {
		out = append(out, GuestListItem{ID: g.ID, DisplayName: g.DisplayName, Status: string(g.Status)})
	}
	return out, nil
}

func (s *ProfileService) Get(ctx context.Context, id string) (*GuestProfile, error) {
	g, err := s.guests.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tenantCount, err := s.guests.TenantCount(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	tenants, err := s.guests.TenantNames(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	ids, err := s.identifiers.ListByGuest(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	identifiers := make([]GuestIdentifierDTO, 0, len(ids))
	for _, id := range ids {
		identifiers = append(identifiers, GuestIdentifierDTO{IDType: id.IDType, Value: id.ValueNormalized})
	}
	return &GuestProfile{
		ID:           g.ID,
		DisplayName:  g.DisplayName,
		Status:       string(g.Status),
		MergedIntoID: g.MergedIntoID,
		TenantCount:  tenantCount,
		Tenants:      tenants,
		Identifiers:  identifiers,
	}, nil
}

func (s *ProfileService) Events(ctx context.Context, guestID string, limit, offset int) ([]EventDTO, error) {
	if _, err := s.guests.GetByID(ctx, guestID); err != nil {
		return nil, err
	}
	events, err := s.events.ListByGuest(ctx, guestID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]EventDTO, 0, len(events))
	for _, e := range events {
		out = append(out, EventDTO{
			ID:          e.ID,
			TypeID:      e.EventTypeID,
			TenantID:    e.TenantID,
			ProjectID:   e.ProjectID,
			ExternalID:  e.ExternalID,
			OccurredAt:  e.OccurredAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			AmountMinor: e.AmountMinor,
			Currency:    e.Currency,
		})
	}
	return out, nil
}
