package service

import (
	"context"
	"strings"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/identity/normalize"
)

type ProfileService struct {
	guests      domain.GuestRepo
	identifiers domain.GuestIdentifierRepo
	events      domain.EventRepo
	contacts    domain.ContactLogRepo
	clock       domain.Clock
}

func NewProfileService(guests domain.GuestRepo, identifiers domain.GuestIdentifierRepo, events domain.EventRepo, contacts domain.ContactLogRepo) *ProfileService {
	return &ProfileService{
		guests:      guests,
		identifiers: identifiers,
		events:      events,
		contacts:    contacts,
		clock:       domain.SystemClock{},
	}
}

type PageMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type GuestListItem struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	Status       string `json:"status"`
	MergedIntoID string `json:"mergedIntoId,omitempty"`
	TenantCount  int    `json:"tenantCount"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type GuestPage struct {
	Items []GuestListItem `json:"items"`
	Meta  PageMeta        `json:"meta"`
}

type GuestProfile struct {
	ID           string               `json:"id"`
	DisplayName  string               `json:"displayName"`
	Status       string               `json:"status"`
	MergedIntoID string               `json:"mergedIntoId,omitempty"`
	TenantCount  int                  `json:"tenantCount"`
	Tenants      []GuestTenantEntry   `json:"tenants"`
	Segments     []GuestSegmentEntry  `json:"segments"`
	Identifiers  []GuestIdentifierDTO `json:"identifiers"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
}

type GuestTenantEntry struct {
	TenantID   string `json:"tenantId"`
	TenantName string `json:"tenantName"`
	EventCount int    `json:"eventCount"`
}

type GuestSegmentEntry struct {
	SegmentID string `json:"segmentId"`
	Name      string `json:"name"`
	VersionNo int    `json:"versionNo,omitempty"`
}

type GuestIdentifierDTO struct {
	ID              string `json:"id"`
	GuestID         string `json:"guestId"`
	IDType          string `json:"idType"`
	ValueNormalized string `json:"valueNormalized"`
}

type EventDTO struct {
	ID          string `json:"id"`
	TypeID      string `json:"eventTypeId"`
	TenantID    string `json:"tenantId"`
	ProjectID   string `json:"projectId"`
	ExternalID  string `json:"externalId"`
	OccurredAt  string `json:"occurredAt"`
	AmountMinor *int64 `json:"amountMinor,omitempty"`
	Currency    string `json:"currency"`
}

func (s *ProfileService) Search(ctx context.Context, q string, limit, offset int) (*GuestPage, error) {
	guests, err := s.guests.Search(ctx, normalizeQuery(q), limit, offset)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	out := make([]GuestListItem, 0, len(guests))
	for _, g := range guests {
		tenantCount, err := s.guests.TenantCount(ctx, g.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, GuestListItem{
			ID:           g.ID,
			DisplayName:  g.DisplayName,
			Status:       string(g.Status),
			MergedIntoID: g.MergedIntoID,
			TenantCount:  tenantCount,
			CreatedAt:    g.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    g.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	total := offset + len(out)
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	page := offset/limit + 1
	return &GuestPage{
		Items: out,
		Meta: PageMeta{
			Page:       page,
			PageSize:   limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func normalizeQuery(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return q
	}
	digits := 0
	for _, r := range q {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	if digits >= 7 && !strings.Contains(q, "@") {
		if p := normalize.Phone(q); p != "" {
			return p
		}
	}
	if strings.Contains(q, "@") {
		return normalize.Email(q)
	}
	return q
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
	tenantRows, err := s.guests.TenantSummaries(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	segmentRows, err := s.guests.SegmentSummaries(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	ids, err := s.identifiers.ListByGuest(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	identifiers := make([]GuestIdentifierDTO, 0, len(ids))
	for _, id := range ids {
		identifiers = append(identifiers, GuestIdentifierDTO{ID: id.ID, GuestID: id.GuestID, IDType: id.IDType, ValueNormalized: id.ValueNormalized})
	}
	tenants := make([]GuestTenantEntry, 0, len(tenantRows))
	for _, t := range tenantRows {
		tenants = append(tenants, GuestTenantEntry{TenantID: t.TenantID, TenantName: t.TenantName, EventCount: t.EventCount})
	}
	segments := make([]GuestSegmentEntry, 0, len(segmentRows))
	for _, sg := range segmentRows {
		segments = append(segments, GuestSegmentEntry{SegmentID: sg.SegmentID, Name: sg.Name, VersionNo: sg.VersionNo})
	}
	return &GuestProfile{
		ID:           g.ID,
		DisplayName:  g.DisplayName,
		Status:       string(g.Status),
		MergedIntoID: g.MergedIntoID,
		TenantCount:  tenantCount,
		Tenants:      tenants,
		Segments:     segments,
		Identifiers:  identifiers,
		CreatedAt:    g.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    g.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
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
