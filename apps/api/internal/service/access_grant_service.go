package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
)

type AccessGrantService struct {
	repo      domain.AccessRepository
	clock     domain.Clock
	ids       domain.IDGen
	users     domain.UserRepo
	tenants   domain.TenantRepo
	auditRepo domain.AuditRepo
}

type GrantItem struct {
	Permission string `json:"permission"`
	Effect     string `json:"effect"`
}

type GrantScope struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type GrantRequest struct {
	SubjectType string       `json:"subjectType"`
	SubjectID   string       `json:"subjectId"`
	PresetCode  string       `json:"presetCode"`
	Status      string       `json:"status"`
	ValidFrom   time.Time    `json:"validFrom"`
	ValidTo     *time.Time   `json:"validTo"`
	Indefinite  bool         `json:"indefinite"`
	Reason      string       `json:"reason"`
	Items       []GrantItem  `json:"items"`
	Scopes      []GrantScope `json:"scopes"`
}

type GrantPatch struct {
	PresetCode *string       `json:"presetCode"`
	ValidFrom  *time.Time    `json:"validFrom"`
	ValidTo    *time.Time    `json:"validTo"`
	Indefinite *bool         `json:"indefinite"`
	Reason     *string       `json:"reason"`
	Items      *[]GrantItem  `json:"items"`
	Scopes     *[]GrantScope `json:"scopes"`
}

type GrantResponse struct {
	ID            string       `json:"id"`
	SubjectType   string       `json:"subjectType"`
	SubjectID     string       `json:"subjectId"`
	PresetCode    string       `json:"presetCode,omitempty"`
	Status        string       `json:"status"`
	ValidFrom     time.Time    `json:"validFrom"`
	ValidTo       *time.Time   `json:"validTo"`
	CreatedBy     string       `json:"createdBy"`
	Reason        string       `json:"reason"`
	PolicyVersion int64        `json:"policyVersion"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
	Items         []GrantItem  `json:"items"`
	Scopes        []GrantScope `json:"scopes"`
}

type PermissionResponse struct {
	Code        string `json:"code"`
	Domain      string `json:"domain"`
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsSensitive bool   `json:"isSensitive"`
	MinK        *int   `json:"minK"`
}

type grantActorPolicy struct {
	version int64
	ceiling []domain.AccessRolePermission
	grants  []domain.AccessGrant
	now     time.Time
}

func NewAccessGrantService(repo domain.AccessRepository, clock domain.Clock, ids domain.IDGen, users ...domain.UserRepo) *AccessGrantService {
	s := &AccessGrantService{repo: repo, clock: clock, ids: ids}
	if len(users) > 0 {
		s.users = users[0]
	}
	return s
}

func (s *AccessGrantService) PolicyVersion(ctx context.Context) (int64, error) {
	return s.repo.PolicyVersion(ctx)
}

func (s *AccessGrantService) CanInvite(ctx context.Context, actor domain.Principal, role domain.Role, tenantID string) (int64, error) {
	policy, err := s.actorPolicy(ctx, actor)
	if err != nil {
		return 0, err
	}
	if !policy.delegatesForPeriod(actor, "user.write", nil) {
		return 0, domain.ErrForbidden
	}
	var permissions []domain.AccessRolePermission
	switch role {
	case domain.RoleTenant:
		if tenantID == "" || s.tenants == nil {
			return 0, domain.ErrValidation
		}
		if _, err := s.tenants.GetByID(ctx, tenantID); err != nil {
			return 0, err
		}
		permissions, err = s.repo.RolePermissions(ctx, domain.RoleTenant)
	case domain.RoleManager:
		if tenantID != "" {
			return 0, domain.ErrValidation
		}
		var grants []domain.AccessGrant
		grants, err = s.repo.ActiveGrants(ctx, domain.Principal{Role: domain.RoleManager}, policy.now)
		for _, grant := range grants {
			for _, item := range grant.Items {
				permissions = append(permissions, domain.AccessRolePermission(item))
			}
		}
	default:
		return 0, domain.ErrForbidden
	}
	if err != nil {
		return 0, err
	}
	if len(permissions) == 0 {
		return 0, domain.ErrForbidden
	}
	for _, permission := range permissions {
		if permission.Effect == "allow" && !policy.delegatesForPeriod(actor, permission.PermissionCode, nil) {
			return 0, domain.ErrForbidden
		}
	}
	version, err := s.repo.PolicyVersion(ctx)
	if err != nil {
		return 0, err
	}
	if version != policy.version {
		return 0, domain.ErrConflict
	}
	return version, nil
}

func (s *AccessGrantService) SetTenantRepository(repo domain.TenantRepo) { s.tenants = repo }
func (s *AccessGrantService) SetAuditRepository(repo domain.AuditRepo)   { s.auditRepo = repo }

func (s *AccessGrantService) subjectTenant(ctx context.Context, subjectType, subjectID string) (string, error) {
	switch subjectType {
	case "tenant":
		if subjectID == "" {
			return "", domain.ErrValidation
		}
		return subjectID, nil
	case "role":
		if subjectID != string(domain.RoleTenant) {
			return "", domain.ErrForbidden
		}
		return "self", nil
	case "user":
		if s.users == nil {
			return "", domain.ErrForbidden
		}
		user, err := s.users.GetByID(ctx, subjectID)
		if err != nil {
			return "", err
		}
		if !user.IsActive {
			return "", domain.ErrForbidden
		}
		if user.Role == domain.RoleManager {
			return "", nil
		}
		if user.Role == domain.RoleTenant && user.TenantID != "" {
			return user.TenantID, nil
		}
		return "", domain.ErrForbidden
	default:
		return "", domain.ErrValidation
	}
}

func (s *AccessGrantService) actorPolicy(ctx context.Context, actor domain.Principal) (grantActorPolicy, error) {
	var policy grantActorPolicy
	if actor.Role != domain.RoleManager || actor.UserID == "" {
		return policy, domain.ErrForbidden
	}
	var err error
	policy.now = s.clock.Now().UTC()
	if policy.version, err = s.repo.PolicyVersion(ctx); err != nil {
		return policy, err
	}
	if policy.ceiling, err = s.repo.RolePermissions(ctx, actor.Role); err != nil {
		return policy, err
	}
	if policy.grants, err = s.repo.ActiveGrants(ctx, actor, policy.now); err != nil {
		return policy, err
	}
	return policy, nil
}

func (p grantActorPolicy) permits(actor domain.Principal, permission, tenantID string) bool {
	return access.Decide(actor, permission, access.Resource{TenantID: tenantID}, p.ceiling, p.grants, p.now).Allowed
}

func (p grantActorPolicy) delegates(actor domain.Principal, permission string) bool {
	if permission == "tenant.aggregates.own" {
		permission = "tenant.aggregates.read"
	}
	if !p.permits(actor, permission, "") {
		return false
	}
	for _, grant := range p.grants {
		if !p.actorGrantActive(actor, grant) {
			continue
		}
		for _, item := range grant.Items {
			if item.PermissionCode != permission {
				continue
			}
			if item.Effect != "allow" {
				return false
			}
			for _, scope := range grant.Scopes {
				if scope.Type != "principal_role" || scope.Value != string(actor.Role) {
					return false
				}
			}
		}
	}
	return true
}

func (p grantActorPolicy) delegatesForPeriod(actor domain.Principal, permission string, validTo *time.Time) bool {
	if !p.delegates(actor, permission) {
		return false
	}
	if permission == "tenant.aggregates.own" {
		permission = "tenant.aggregates.read"
	}
	for _, grant := range p.grants {
		if !p.actorGrantActive(actor, grant) {
			continue
		}
		for _, item := range grant.Items {
			if item.PermissionCode != permission || item.Effect != "allow" {
				continue
			}
			if grant.ValidTo == nil || validTo != nil && !grant.ValidTo.Before(*validTo) {
				return true
			}
		}
	}
	return false
}

func (p grantActorPolicy) actorGrantActive(actor domain.Principal, grant domain.AccessGrant) bool {
	if grant.Status != "active" || p.now.Before(grant.ValidFrom) || grant.ValidTo != nil && !p.now.Before(*grant.ValidTo) {
		return false
	}
	return grant.SubjectType == "user" && grant.SubjectID == actor.UserID || grant.SubjectType == "role" && grant.SubjectID == string(actor.Role)
}

func (s *AccessGrantService) subjectRole(ctx context.Context, subjectType, subjectID string) (domain.Role, error) {
	if subjectType != "user" {
		return domain.RoleTenant, nil
	}
	if s.users == nil {
		return "", domain.ErrForbidden
	}
	user, err := s.users.GetByID(ctx, subjectID)
	if err != nil {
		return "", err
	}
	if !user.IsActive || user.Role != domain.RoleManager && user.Role != domain.RoleTenant {
		return "", domain.ErrForbidden
	}
	return user.Role, nil
}

func (s *AccessGrantService) Permissions(ctx context.Context, actor domain.Principal) ([]PermissionResponse, error) {
	policy, err := s.actorPolicy(ctx, actor)
	if err != nil {
		return nil, err
	}
	if !policy.permits(actor, "access.grant.read", "") {
		return nil, domain.ErrForbidden
	}
	permissions, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]PermissionResponse, 0, len(permissions))
	for _, p := range permissions {
		if p.IsActive {
			result = append(result, PermissionResponse{p.Code, p.Domain, p.Kind, p.Title, p.Description, p.IsSensitive, p.MinK})
		}
	}
	return result, nil
}

func (s *AccessGrantService) List(ctx context.Context, actor domain.Principal, filter domain.AccessGrantFilter) ([]GrantResponse, error) {
	policy, err := s.actorPolicy(ctx, actor)
	if err != nil {
		return nil, err
	}
	if !policy.delegates(actor, "access.grant.read") {
		return nil, domain.ErrForbidden
	}
	grants, err := s.repo.ListGrants(ctx, filter)
	if err != nil {
		return nil, err
	}
	result := make([]GrantResponse, 0, len(grants))
	for _, grant := range grants {
		result = append(result, grantResponse(grant))
	}
	return result, nil
}

func (s *AccessGrantService) Get(ctx context.Context, actor domain.Principal, id string) (*GrantResponse, error) {
	policy, err := s.actorPolicy(ctx, actor)
	if err != nil {
		return nil, err
	}
	if !policy.delegates(actor, "access.grant.read") {
		return nil, domain.ErrForbidden
	}
	grant, err := s.repo.GetGrant(ctx, id)
	if err != nil {
		return nil, err
	}
	tenantID, err := s.subjectTenant(ctx, grant.SubjectType, grant.SubjectID)
	if err != nil {
		return nil, err
	}
	if tenantID == "self" {
		tenantID = ""
	}
	if !policy.permits(actor, "access.grant.read", tenantID) {
		return nil, domain.ErrForbidden
	}
	result := grantResponse(*grant)
	return &result, nil
}

func (s *AccessGrantService) Create(ctx context.Context, actor domain.Principal, request GrantRequest, requestID string) (*GrantResponse, error) {
	policy, err := s.actorPolicy(ctx, actor)
	if err != nil {
		return nil, err
	}
	if !policy.delegates(actor, "access.grant.write") {
		return nil, domain.ErrForbidden
	}
	tenantID, err := s.subjectTenant(ctx, request.SubjectType, request.SubjectID)
	if err != nil {
		return nil, err
	}
	if tenantID == "self" {
		tenantID = ""
	}
	if !policy.permits(actor, "access.grant.write", tenantID) {
		return nil, domain.ErrForbidden
	}
	if request.ValidTo == nil && !request.Indefinite || request.ValidTo != nil && request.Indefinite {
		return nil, domain.ErrValidation
	}
	targetRole, err := s.subjectRole(ctx, request.SubjectType, request.SubjectID)
	if err != nil {
		return nil, err
	}
	if targetRole == domain.RoleManager && request.SubjectID == actor.UserID {
		return nil, domain.ErrForbidden
	}
	grant := domain.AccessGrant{
		ID: s.ids.NewID(ctx), SubjectType: request.SubjectType, SubjectID: request.SubjectID,
		PresetCode: request.PresetCode, Status: request.Status, ValidFrom: request.ValidFrom,
		ValidTo: request.ValidTo, CreatedBy: actor.UserID, Reason: request.Reason,
		CreatedAt: policy.now, UpdatedAt: policy.now,
	}
	grant.Items = toGrantItems(request.Items)
	grant.Scopes, err = toGrantScopesForRole(request.Scopes, targetRole)
	if err != nil {
		return nil, err
	}
	if grant.Status != "active" && grant.Status != "draft" {
		return nil, domain.ErrValidation
	}
	if err := s.validate(ctx, actor, policy, grant); err != nil {
		return nil, err
	}
	audit, err := s.audit(ctx, actor, "access.grant.create", nil, &grant, requestID, policy.now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateGrant(ctx, &grant, policy.version, audit); err != nil {
		return nil, err
	}
	result := grantResponse(grant)
	return &result, nil
}

func (s *AccessGrantService) Update(ctx context.Context, actor domain.Principal, id string, patch GrantPatch, requestID string) (*GrantResponse, error) {
	return s.change(ctx, actor, id, "access.grant.update", requestID, func(g *domain.AccessGrant, now time.Time) error {
		if g.Status == "revoked" || g.Status == "expired" || g.ValidTo != nil && !now.Before(*g.ValidTo) {
			return domain.ErrConflict
		}
		if patch.PresetCode != nil {
			g.PresetCode = *patch.PresetCode
		}
		if patch.ValidFrom != nil {
			g.ValidFrom = *patch.ValidFrom
		}
		if patch.ValidTo != nil {
			g.ValidTo = patch.ValidTo
		}
		if patch.Indefinite != nil {
			if *patch.Indefinite {
				if patch.ValidTo != nil {
					return domain.ErrValidation
				}
				g.ValidTo = nil
			} else if g.ValidTo == nil {
				return domain.ErrValidation
			}
		}
		if patch.Reason != nil {
			g.Reason = *patch.Reason
		}
		if patch.Items != nil {
			g.Items = toGrantItems(*patch.Items)
		}
		if patch.Scopes != nil {
			var err error
			role, roleErr := s.subjectRole(ctx, g.SubjectType, g.SubjectID)
			if roleErr != nil {
				return roleErr
			}
			g.Scopes, err = toGrantScopesForRole(*patch.Scopes, role)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AccessGrantService) Transition(ctx context.Context, actor domain.Principal, id, action, reason, requestID string) (*GrantResponse, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, domain.ErrValidation
	}
	if action != "activate" && action != "suspend" && action != "revoke" {
		return nil, domain.ErrValidation
	}
	return s.change(ctx, actor, id, "access.grant."+action, requestID, func(g *domain.AccessGrant, now time.Time) error {
		if g.ValidTo != nil && !now.Before(*g.ValidTo) {
			return domain.ErrConflict
		}
		switch action {
		case "activate":
			if g.Status != "draft" && g.Status != "suspended" {
				return domain.ErrConflict
			}
			g.Status = "active"
		case "suspend":
			if g.Status != "active" {
				return domain.ErrConflict
			}
			g.Status = "suspended"
		case "revoke":
			if g.Status != "draft" && g.Status != "active" && g.Status != "suspended" {
				return domain.ErrConflict
			}
			g.Status = "revoked"
		}
		g.Reason = reason
		return nil
	})
}

func (s *AccessGrantService) change(ctx context.Context, actor domain.Principal, id, action, requestID string, mutate func(*domain.AccessGrant, time.Time) error) (*GrantResponse, error) {
	policy, err := s.actorPolicy(ctx, actor)
	if err != nil {
		return nil, err
	}
	if !policy.delegates(actor, "access.grant.write") {
		return nil, domain.ErrForbidden
	}
	grant, err := s.repo.GetGrant(ctx, id)
	if err != nil {
		return nil, err
	}
	tenantID := ""
	if grantChangeNeedsValidation(action, *grant) {
		tenantID, err = s.subjectTenant(ctx, grant.SubjectType, grant.SubjectID)
		if err != nil {
			return nil, err
		}
	}
	if tenantID == "self" {
		tenantID = ""
	}
	if !policy.permits(actor, "access.grant.write", tenantID) {
		return nil, domain.ErrForbidden
	}
	before := *grant
	needsValidation := grantChangeNeedsValidation(action, before)
	if needsValidation {
		if err := s.validate(ctx, actor, policy, before); err != nil {
			return nil, err
		}
	}
	if err := mutate(grant, policy.now); err != nil {
		return nil, err
	}
	grant.UpdatedAt = policy.now
	if needsValidation {
		if err := s.validate(ctx, actor, policy, *grant); err != nil {
			return nil, err
		}
	}
	audit, err := s.audit(ctx, actor, action, &before, grant, requestID, policy.now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateGrant(ctx, grant, policy.version, audit); err != nil {
		return nil, err
	}
	result := grantResponse(*grant)
	return &result, nil
}

func grantChangeNeedsValidation(action string, grant domain.AccessGrant) bool {
	if action != "access.grant.revoke" && action != "access.grant.suspend" {
		return true
	}
	for _, item := range grant.Items {
		if item.Effect != "allow" {
			return true
		}
	}
	return false
}

func (s *AccessGrantService) validate(ctx context.Context, actor domain.Principal, policy grantActorPolicy, grant domain.AccessGrant) error {
	if strings.TrimSpace(grant.Reason) == "" || grant.ValidFrom.IsZero() || grant.ValidTo != nil && !grant.ValidTo.After(grant.ValidFrom) {
		return domain.ErrValidation
	}
	role, err := s.subjectRole(ctx, grant.SubjectType, grant.SubjectID)
	if err != nil {
		return err
	}
	if role == domain.RoleManager {
		return s.validateManagerGrant(ctx, actor, policy, grant)
	}
	tenantID, err := s.subjectTenant(ctx, grant.SubjectType, grant.SubjectID)
	if err != nil {
		return err
	}
	catalog, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return err
	}
	tenantCeiling, err := s.repo.RolePermissions(ctx, domain.RoleTenant)
	if err != nil {
		return err
	}
	active := make(map[string]bool, len(catalog))
	for _, p := range catalog {
		active[p.Code] = p.IsActive
	}
	allowed := make(map[string]bool, len(tenantCeiling))
	for _, p := range tenantCeiling {
		allowed[p.PermissionCode] = p.Effect == "allow"
	}
	seen := make(map[string]bool)
	for _, item := range grant.Items {
		if !active[item.PermissionCode] || !allowed[item.PermissionCode] || seen[item.PermissionCode] || item.Effect != "allow" && item.Effect != "deny" {
			return domain.ErrValidation
		}
		if !policy.delegatesForPeriod(actor, item.PermissionCode, grant.ValidTo) {
			return domain.ErrForbidden
		}
		seen[item.PermissionCode] = true
	}
	if len(seen) == 0 {
		return domain.ErrValidation
	}
	seenScope := make(map[string]bool)
	ownTenant := false
	for _, scope := range grant.Scopes {
		key := scope.Type + "\x00" + scope.Value
		if seenScope[key] || scope.Value == "" {
			return domain.ErrValidation
		}
		seenScope[key] = true
		if scope.Type == "principal_role" && scope.Value != string(domain.RoleTenant) {
			return domain.ErrValidation
		}
		if scope.Type == "tenant" {
			if scope.Value != tenantID {
				return domain.ErrValidation
			}
			ownTenant = true
		}
	}
	if !ownTenant {
		return domain.ErrValidation
	}
	if !seenScope["principal_role\x00"+string(domain.RoleTenant)] {
		return domain.ErrValidation
	}
	for _, item := range grant.Items {
		probe := grant
		probeTenantID := tenantID
		if tenantID == "self" {
			probeTenantID = "access-probe-tenant"
		}
		probe.Status = "active"
		probe.ValidFrom = policy.now
		probe.ValidTo = nil
		probe.SubjectType = "tenant"
		probe.SubjectID = probeTenantID
		probe.Items = []domain.AccessGrantItem{{PermissionCode: item.PermissionCode, Effect: "allow"}}
		resource := access.Resource{TenantID: probeTenantID}
		for _, scope := range grant.Scopes {
			switch scope.Type {
			case "category":
				resource.Category = scope.Value
			case "channel":
				resource.Channel = scope.Value
			case "metric":
				resource.Metric = scope.Value
			}
		}
		decision := access.Decide(domain.Principal{Role: domain.RoleTenant, TenantID: probeTenantID}, item.PermissionCode, resource, tenantCeiling, []domain.AccessGrant{probe}, policy.now)
		if !decision.Allowed {
			return domain.ErrValidation
		}
	}
	return nil
}

func (s *AccessGrantService) validateManagerGrant(ctx context.Context, actor domain.Principal, policy grantActorPolicy, grant domain.AccessGrant) error {
	if grant.SubjectType != "user" || grant.SubjectID == actor.UserID {
		return domain.ErrForbidden
	}
	if grant.PresetCode != "" || len(grant.Scopes) != 1 || grant.Scopes[0].Type != "principal_role" || grant.Scopes[0].Value != string(domain.RoleManager) || len(grant.Items) == 0 {
		return domain.ErrValidation
	}
	catalog, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return err
	}
	ceiling, err := s.repo.RolePermissions(ctx, domain.RoleManager)
	if err != nil {
		return err
	}
	active := make(map[string]bool, len(catalog))
	for _, permission := range catalog {
		active[permission.Code] = permission.IsActive
	}
	allowed := make(map[string]bool, len(ceiling))
	for _, permission := range ceiling {
		allowed[permission.PermissionCode] = permission.Effect == "allow"
	}
	seen := make(map[string]bool, len(grant.Items))
	for _, item := range grant.Items {
		if !managerPersonalPermission(item.PermissionCode) || !active[item.PermissionCode] || !allowed[item.PermissionCode] || seen[item.PermissionCode] || item.Effect != "allow" && item.Effect != "deny" {
			return domain.ErrValidation
		}
		if !policy.delegatesForPeriod(actor, item.PermissionCode, grant.ValidTo) {
			return domain.ErrForbidden
		}
		seen[item.PermissionCode] = true
	}
	return nil
}

func managerPersonalPermission(code string) bool {
	switch code {
	case "guest.profile.read", "guest.profile.search", "guest.event.read", "guest.identifier.read", "guest.contact_history.read",
		"field.guest.display_name.read", "field.guest.phone.read", "field.guest.email.read", "field.guest.card.read",
		"field.guest.total_amount.read", "field.guest.identifiers.read", "field.guest.segments.read", "field.guest.tenants_visited.read",
		"field.event.amount.read", "field.event.raw_payload.read", "field.event.external_id.read":
		return true
	default:
		return false
	}
}

func (s *AccessGrantService) audit(ctx context.Context, actor domain.Principal, action string, before, after *domain.AccessGrant, requestID string, now time.Time) (*domain.AuditLog, error) {
	details, err := json.Marshal(map[string]any{"before": before, "after": after})
	if err != nil {
		return nil, err
	}
	return &domain.AuditLog{ID: s.ids.NewID(ctx), ActorUserID: actor.UserID, ActorRole: string(actor.Role), Action: action, EntityType: "access_grant", EntityID: after.ID, Details: string(details), RequestID: requestID, CreatedAt: now}, nil
}

func toGrantItems(items []GrantItem) []domain.AccessGrantItem {
	result := make([]domain.AccessGrantItem, 0, len(items))
	for _, item := range items {
		result = append(result, domain.AccessGrantItem{PermissionCode: item.Permission, Effect: item.Effect})
	}
	return result
}

func toGrantScopes(scopes []GrantScope) ([]domain.AccessGrantScope, error) {
	return toGrantScopesForRole(scopes, domain.RoleTenant)
}

func toGrantScopesForRole(scopes []GrantScope, role domain.Role) ([]domain.AccessGrantScope, error) {
	result := make([]domain.AccessGrantScope, 0, len(scopes)+1)
	for _, scope := range scopes {
		if scope.Type == "principal_role" {
			return nil, domain.ErrValidation
		}
		result = append(result, domain.AccessGrantScope{Type: scope.Type, Value: scope.Value})
	}
	result = append(result, domain.AccessGrantScope{Type: "principal_role", Value: string(role)})
	return result, nil
}

func grantResponse(grant domain.AccessGrant) GrantResponse {
	result := GrantResponse{ID: grant.ID, SubjectType: grant.SubjectType, SubjectID: grant.SubjectID, PresetCode: grant.PresetCode, Status: grant.Status, ValidFrom: grant.ValidFrom, ValidTo: grant.ValidTo, CreatedBy: grant.CreatedBy, Reason: grant.Reason, PolicyVersion: grant.PolicyVersion, CreatedAt: grant.CreatedAt, UpdatedAt: grant.UpdatedAt, Items: make([]GrantItem, 0, len(grant.Items)), Scopes: make([]GrantScope, 0, len(grant.Scopes))}
	for _, item := range grant.Items {
		result.Items = append(result.Items, GrantItem{Permission: item.PermissionCode, Effect: item.Effect})
	}
	for _, scope := range grant.Scopes {
		if scope.Type != "principal_role" {
			result.Scopes = append(result.Scopes, GrantScope{Type: scope.Type, Value: scope.Value})
		}
	}
	return result
}

func (s *AccessGrantService) Authorize(ctx context.Context, principal domain.Principal, permission string, resource access.Resource) (access.Decision, error) {
	version, err := s.repo.PolicyVersion(ctx)
	if err != nil {
		return access.Decision{}, err
	}
	ceiling, err := s.repo.RolePermissions(ctx, principal.Role)
	if err != nil {
		return access.Decision{}, err
	}
	now := s.clock.Now().UTC()
	grants, err := s.repo.ActiveGrants(ctx, principal, now)
	if err != nil {
		return access.Decision{}, err
	}
	currentVersion, err := s.repo.PolicyVersion(ctx)
	if err != nil {
		return access.Decision{}, err
	}
	if currentVersion != version {
		return access.Decision{}, domain.ErrConflict
	}
	return access.Decide(principal, permission, resource, ceiling, grants, now), nil
}

func (s *AccessGrantService) Effective(ctx context.Context, principal domain.Principal) ([]string, int64, error) {
	version, err := s.repo.PolicyVersion(ctx)
	if err != nil {
		return nil, 0, err
	}
	ceiling, err := s.repo.RolePermissions(ctx, principal.Role)
	if err != nil {
		return nil, 0, err
	}
	now := s.clock.Now().UTC()
	grants, err := s.repo.ActiveGrants(ctx, principal, now)
	if err != nil {
		return nil, 0, err
	}
	currentVersion, err := s.repo.PolicyVersion(ctx)
	if err != nil {
		return nil, 0, err
	}
	if currentVersion != version {
		return nil, 0, domain.ErrConflict
	}
	return effectivePermissions(principal, ceiling, grants, now), version, nil
}

func effectivePermissions(principal domain.Principal, ceiling []domain.AccessRolePermission, grants []domain.AccessGrant, now time.Time) []string {
	result := make([]string, 0)
	for _, permission := range ceiling {
		if permission.Effect != "allow" {
			continue
		}
		resources := make([]access.Resource, 0, 1+len(grants))
		resources = append(resources, access.Resource{TenantID: principal.TenantID})
		for _, grant := range grants {
			resource := access.Resource{TenantID: principal.TenantID}
			for _, scope := range grant.Scopes {
				switch scope.Type {
				case "tenant":
					if scope.Value != "self" {
						resource.TenantID = scope.Value
					}
				case "category":
					resource.Category = scope.Value
				case "channel":
					resource.Channel = scope.Value
				case "metric":
					resource.Metric = scope.Value
				}
			}
			resources = append(resources, resource)
		}
		for _, resource := range resources {
			if access.Decide(principal, permission.PermissionCode, resource, ceiling, grants, now).Allowed {
				result = append(result, permission.PermissionCode)
				break
			}
		}
	}
	return result
}
