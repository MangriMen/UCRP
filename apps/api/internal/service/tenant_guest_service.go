package service

import (
	"context"
	"time"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
)

type tenantGuestAuthorizer interface {
	Authorize(context.Context, domain.Principal, string, access.Resource) (access.Decision, error)
	PolicyVersion(context.Context) (int64, error)
}

type TenantGuestService struct {
	repo   domain.TenantGuestRepo
	policy tenantGuestAuthorizer
}

func NewTenantGuestService(repo domain.TenantGuestRepo, policy tenantGuestAuthorizer) *TenantGuestService {
	return &TenantGuestService{repo: repo, policy: policy}
}

type TenantGuestProfile struct {
	ID          string                  `json:"id"`
	DisplayName *string                 `json:"displayName"`
	Status      string                  `json:"status"`
	FirstSeenAt time.Time               `json:"firstSeenAt"`
	Identifiers []TenantGuestIdentifier `json:"identifiers"`
	Tenants     []string                `json:"tenants"`
	Segments    []string                `json:"segments"`
}

type TenantGuestIdentifier struct {
	ID              string    `json:"id"`
	GuestID         string    `json:"guestId"`
	IDType          string    `json:"idType"`
	ValueNormalized string    `json:"valueNormalized"`
	CreatedAt       time.Time `json:"createdAt"`
}

type TenantGuestEvent struct {
	ID          string    `json:"id"`
	EventTypeID string    `json:"eventTypeId"`
	TenantID    string    `json:"tenantId"`
	ProjectID   *string   `json:"projectId"`
	ExternalID  *string   `json:"externalId"`
	OccurredAt  time.Time `json:"occurredAt"`
	AmountMinor *int64    `json:"amountMinor"`
	Currency    *string   `json:"currency"`
}

type tenantGuestLimits struct {
	periodDays int
	rowLimit   int
	masked     map[string]bool
}

func (l *tenantGuestLimits) include(d access.Decision) {
	if d.PeriodDays > 0 && (l.periodDays == 0 || d.PeriodDays < l.periodDays) {
		l.periodDays = d.PeriodDays
	}
	if d.RowLimit > 0 && (l.rowLimit == 0 || d.RowLimit < l.rowLimit) {
		l.rowLimit = d.RowLimit
	}
	for _, field := range d.MaskFields {
		l.masked[field] = true
	}
}

func (l tenantGuestLimits) since() time.Time {
	if l.periodDays == 0 {
		return time.Time{}
	}
	return time.Now().UTC().AddDate(0, 0, -l.periodDays)
}

func (s *TenantGuestService) decide(ctx context.Context, principal domain.Principal, permission string, limits *tenantGuestLimits) (bool, error) {
	decision, err := s.policy.Authorize(ctx, principal, permission, access.Resource{TenantID: principal.TenantID})
	if err != nil {
		return false, err
	}
	if decision.Allowed {
		limits.include(decision)
	}
	return decision.Allowed, nil
}

func (s *TenantGuestService) Profile(ctx context.Context, principal domain.Principal, guestID string) (result *TenantGuestProfile, err error) {
	version, err := s.policy.PolicyVersion(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = s.checkVersion(ctx, version)
			if err != nil {
				result = nil
			}
		}
	}()
	limits := tenantGuestLimits{masked: make(map[string]bool)}
	allowed, err := s.decide(ctx, principal, "guest.profile.read", &limits)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	identifierAccess, err := s.decide(ctx, principal, "guest.identifier.read", &limits)
	if err != nil {
		return nil, err
	}
	if identifierAccess {
		identifierAccess, err = s.decide(ctx, principal, "field.guest.identifiers.read", &limits)
		if err != nil {
			return nil, err
		}
	}
	allowedTypes := make(map[string]bool)
	if identifierAccess {
		for _, kind := range []string{"phone", "email", "card"} {
			allowedTypes[kind], err = s.decide(ctx, principal, "field.guest."+kind+".read", &limits)
			if err != nil {
				return nil, err
			}
		}
	}
	since := limits.since()
	guest, err := s.repo.Get(ctx, principal.TenantID, guestID, since)
	if err != nil {
		return nil, err
	}
	result = &TenantGuestProfile{ID: guest.ID, Status: "active", FirstSeenAt: guest.FirstSeenAt,
		Identifiers: []TenantGuestIdentifier{}, Tenants: []string{}, Segments: []string{}}
	if !identifierAccess {
		return result, nil
	}
	limit := 200
	if limits.rowLimit > 0 && limits.rowLimit < limit {
		limit = limits.rowLimit
	}
	identifiers, err := s.repo.Identifiers(ctx, principal.TenantID, guestID, since, limit)
	if err != nil {
		return nil, err
	}
	for _, identifier := range identifiers {
		if !allowedTypes[identifier.Type] || limits.masked[identifier.Type] {
			continue
		}
		result.Identifiers = append(result.Identifiers, TenantGuestIdentifier{
			ID: identifier.ID, GuestID: guestID, IDType: identifier.Type,
			ValueNormalized: identifier.ValueNormalized, CreatedAt: identifier.FirstSeenAt,
		})
	}
	return result, nil
}

func (s *TenantGuestService) Events(ctx context.Context, principal domain.Principal, guestID string, limit, offset int) (result []TenantGuestEvent, total int, err error) {
	version, err := s.policy.PolicyVersion(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err == nil {
			err = s.checkVersion(ctx, version)
			if err != nil {
				result = nil
				total = 0
			}
		}
	}()
	limits := tenantGuestLimits{masked: make(map[string]bool)}
	allowed, err := s.decide(ctx, principal, "guest.event.read", &limits)
	if err != nil {
		return nil, 0, err
	}
	if !allowed {
		return nil, 0, domain.ErrForbidden
	}
	amountAllowed, err := s.decide(ctx, principal, "field.event.amount.read", &limits)
	if err != nil {
		return nil, 0, err
	}
	externalAllowed, err := s.decide(ctx, principal, "field.event.external_id.read", &limits)
	if err != nil {
		return nil, 0, err
	}
	since := limits.since()
	if _, err := s.repo.Get(ctx, principal.TenantID, guestID, since); err != nil {
		return nil, 0, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	total, err = s.repo.CountEvents(ctx, principal.TenantID, guestID, since)
	if err != nil {
		return nil, 0, err
	}
	if limits.rowLimit > 0 {
		if total > limits.rowLimit {
			total = limits.rowLimit
		}
		if offset >= limits.rowLimit {
			return []TenantGuestEvent{}, total, nil
		}
		if limit > limits.rowLimit-offset {
			limit = limits.rowLimit - offset
		}
	}
	events, err := s.repo.Events(ctx, principal.TenantID, guestID, since, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result = make([]TenantGuestEvent, 0, len(events))
	for _, event := range events {
		item := TenantGuestEvent{ID: event.ID, EventTypeID: event.EventTypeID, TenantID: event.TenantID, OccurredAt: event.OccurredAt}
		if event.ProjectID != "" {
			item.ProjectID = &event.ProjectID
		}
		if externalAllowed {
			item.ExternalID = &event.ExternalID
		}
		if amountAllowed && !limits.masked["amount"] {
			item.AmountMinor = event.AmountMinor
			item.Currency = &event.Currency
		}
		result = append(result, item)
	}
	return result, total, nil
}

func (s *TenantGuestService) checkVersion(ctx context.Context, version int64) error {
	current, err := s.policy.PolicyVersion(ctx)
	if err != nil {
		return err
	}
	if current != version {
		return domain.ErrConflict
	}
	return nil
}
