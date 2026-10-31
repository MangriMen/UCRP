package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
)

type AccessPreset struct {
	Code      string      `json:"code"`
	Items     []GrantItem `json:"items"`
	Available bool        `json:"available"`
}

type SubjectPermissions struct {
	SubjectType        string   `json:"subjectType"`
	SubjectID          string   `json:"subjectId"`
	PolicyVersion      int64    `json:"policyVersion"`
	Permissions        []string `json:"permissions"`
	UserGrantsIncluded bool     `json:"userGrantsIncluded"`
}

type AccessAuditEntry struct {
	ID        string    `json:"id"`
	ActorID   string    `json:"actorId"`
	ActorRole string    `json:"actorRole"`
	Action    string    `json:"action"`
	EntityID  string    `json:"entityId"`
	Details   any       `json:"details"`
	RequestID string    `json:"requestId"`
	CreatedAt time.Time `json:"createdAt"`
}

type accessAuditLister interface {
	ListAccessAudit(context.Context, string, string, *time.Time, *time.Time, int, int) ([]domain.AuditLog, error)
}

func (s *AccessGrantService) Presets(ctx context.Context, actor domain.Principal) ([]AccessPreset, error) {
	policy, err := s.actorPolicy(ctx, actor)
	if err != nil {
		return nil, err
	}
	if !policy.delegates(actor, "access.grant.read") {
		return nil, domain.ErrForbidden
	}
	ceiling, err := s.repo.RolePermissions(ctx, domain.RoleTenant)
	if err != nil {
		return nil, err
	}
	tenantPermissions := make(map[string]bool, len(ceiling))
	for _, permission := range ceiling {
		tenantPermissions[permission.PermissionCode] = permission.Effect == "allow"
	}
	presets := []struct {
		code        string
		permissions []string
	}{
		{"tenant.none", nil},
		{"tenant.self_aggregates", []string{"tenant.aggregates.own", "metric.new_guests_via_mall.read", "metric.repeat_visits.read", "metric.visits_count.read", "reference.read"}},
		{"tenant.self_aggregates_extended", []string{"tenant.aggregates.own", "metric.new_guests_via_mall.read", "metric.repeat_visits.read", "metric.visits_count.read", "reference.read", "metric.segment_size.read", "metric.export_counts.read", "metric.contact_frequency.read"}},
		{"tenant.recommendations_inbound", []string{"tenant.aggregates.own", "metric.new_guests_via_mall.read"}},
		{"tenant.campaign_participant", []string{"campaign.read", "metric.campaign_results.read"}},
		{"tenant.own_events", []string{"guest.event.read", "field.event.amount.read"}},
		{"tenant.full_profile_read", []string{"guest.profile.read", "guest.identifier.read", "field.guest.identifiers.read", "field.guest.phone.read"}},
	}
	result := make([]AccessPreset, 0, len(presets))
	for _, preset := range presets {
		available := policy.delegates(actor, "access.grant.write") && preset.code != "tenant.none" && preset.code != "tenant.recommendations_inbound"
		item := AccessPreset{Code: preset.code, Available: available, Items: make([]GrantItem, 0, len(preset.permissions))}
		for _, permission := range preset.permissions {
			item.Items = append(item.Items, GrantItem{Permission: permission, Effect: "allow"})
			if !tenantPermissions[permission] || !policy.delegates(actor, permission) {
				item.Available = false
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *AccessGrantService) subjectPrincipal(ctx context.Context, subjectType, subjectID string) (domain.Principal, error) {
	tenantID, err := s.subjectTenant(ctx, subjectType, subjectID)
	if err != nil {
		return domain.Principal{}, err
	}
	if subjectType == "tenant" {
		if s.tenants == nil {
			return domain.Principal{}, errors.New("tenant repository is required")
		}
		if _, err := s.tenants.GetByID(ctx, tenantID); err != nil {
			return domain.Principal{}, err
		}
	}
	if tenantID == "self" {
		tenantID = "access-role-preview"
	}
	role, err := s.subjectRole(ctx, subjectType, subjectID)
	if err != nil {
		return domain.Principal{}, err
	}
	principal := domain.Principal{Role: role, TenantID: tenantID}
	if subjectType == "user" {
		principal.UserID = subjectID
	}
	return principal, nil
}

func (s *AccessGrantService) SubjectEffective(ctx context.Context, actor domain.Principal, subjectType, subjectID string) (*SubjectPermissions, error) {
	policy, err := s.actorPolicy(ctx, actor)
	if err != nil {
		return nil, err
	}
	if !policy.delegates(actor, "access.grant.read") {
		return nil, domain.ErrForbidden
	}
	target, err := s.subjectPrincipal(ctx, subjectType, subjectID)
	if err != nil {
		return nil, err
	}
	tenantID := target.TenantID
	if subjectType == "role" {
		tenantID = ""
	}
	if !policy.permits(actor, "access.grant.read", tenantID) {
		return nil, domain.ErrForbidden
	}
	ceiling, err := s.repo.RolePermissions(ctx, target.Role)
	if err != nil {
		return nil, err
	}
	grants, err := s.repo.ActiveGrants(ctx, target, policy.now)
	if err != nil {
		return nil, err
	}
	version, err := s.repo.PolicyVersion(ctx)
	if err != nil {
		return nil, err
	}
	if version != policy.version {
		return nil, domain.ErrConflict
	}
	return &SubjectPermissions{SubjectType: subjectType, SubjectID: subjectID, PolicyVersion: version,
		Permissions: effectivePermissions(target, ceiling, grants, policy.now), UserGrantsIncluded: subjectType == "user"}, nil
}

func (s *AccessGrantService) Simulate(ctx context.Context, actor domain.Principal, request GrantRequest, requestID string) (*SubjectPermissions, error) {
	policy, err := s.actorPolicy(ctx, actor)
	if err != nil {
		return nil, err
	}
	if !policy.delegates(actor, "access.simulate") {
		return nil, domain.ErrForbidden
	}
	target, err := s.subjectPrincipal(ctx, request.SubjectType, request.SubjectID)
	if err != nil {
		return nil, err
	}
	tenantID := target.TenantID
	if request.SubjectType == "role" {
		tenantID = ""
	}
	if !policy.permits(actor, "access.simulate", tenantID) {
		return nil, domain.ErrForbidden
	}
	if request.ValidTo == nil && !request.Indefinite || request.ValidTo != nil && request.Indefinite {
		return nil, domain.ErrValidation
	}
	if request.Status != "active" && request.Status != "draft" {
		return nil, domain.ErrValidation
	}
	scopes, err := toGrantScopesForRole(request.Scopes, target.Role)
	if err != nil {
		return nil, err
	}
	grant := domain.AccessGrant{ID: "simulation", SubjectType: request.SubjectType, SubjectID: request.SubjectID,
		Status: request.Status, ValidFrom: request.ValidFrom, ValidTo: request.ValidTo, Reason: request.Reason,
		PresetCode: request.PresetCode, Items: toGrantItems(request.Items), Scopes: scopes}
	if err := s.validate(ctx, actor, policy, grant); err != nil {
		return nil, err
	}
	ceiling, err := s.repo.RolePermissions(ctx, target.Role)
	if err != nil {
		return nil, err
	}
	grants, err := s.repo.ActiveGrants(ctx, target, policy.now)
	if err != nil {
		return nil, err
	}
	grants = append(grants, grant)
	version, err := s.repo.PolicyVersion(ctx)
	if err != nil {
		return nil, err
	}
	if version != policy.version {
		return nil, domain.ErrConflict
	}
	if s.auditRepo == nil {
		return nil, errors.New("audit repository is required")
	}
	details, err := json.Marshal(map[string]any{"subjectType": request.SubjectType, "subjectId": request.SubjectID, "grant": grantResponse(grant)})
	if err != nil {
		return nil, err
	}
	audit := &domain.AuditLog{ID: s.ids.NewID(ctx), ActorUserID: actor.UserID, ActorRole: string(actor.Role), Action: "access.simulate", EntityType: "access_grant", EntityID: request.SubjectID, Details: string(details), RequestID: requestID, CreatedAt: policy.now}
	if err := s.auditRepo.Insert(ctx, audit); err != nil {
		return nil, err
	}
	return &SubjectPermissions{SubjectType: request.SubjectType, SubjectID: request.SubjectID, PolicyVersion: version,
		Permissions: effectivePermissions(target, ceiling, grants, policy.now), UserGrantsIncluded: request.SubjectType == "user"}, nil
}

func (s *AccessGrantService) Audit(ctx context.Context, actor domain.Principal, grantID, actorID string, from, to *time.Time, limit, offset int) ([]AccessAuditEntry, error) {
	if actor.Role != domain.RoleManager && actor.Role != domain.RoleAnalyst {
		return nil, domain.ErrForbidden
	}
	decision, err := s.Authorize(ctx, actor, "audit.read", access.Resource{})
	if err != nil {
		return nil, err
	}
	if !decision.Allowed {
		return nil, domain.ErrForbidden
	}
	if s.auditRepo == nil {
		return nil, errors.New("audit repository is required")
	}
	if from != nil && to != nil && from.After(*to) || offset < 0 || offset > 10000 {
		return nil, domain.ErrValidation
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	lister, ok := s.auditRepo.(accessAuditLister)
	if !ok {
		return nil, errors.New("access audit repository is required")
	}
	logs, err := lister.ListAccessAudit(ctx, grantID, actorID, from, to, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]AccessAuditEntry, 0, len(logs))
	for _, entry := range logs {
		var details any
		if entry.Details != "" && json.Unmarshal([]byte(entry.Details), &details) != nil {
			details = entry.Details
		}
		result = append(result, AccessAuditEntry{ID: entry.ID, ActorID: entry.ActorUserID, ActorRole: entry.ActorRole, Action: entry.Action, EntityID: entry.EntityID, Details: details, RequestID: entry.RequestID, CreatedAt: entry.CreatedAt})
	}
	return result, nil
}
