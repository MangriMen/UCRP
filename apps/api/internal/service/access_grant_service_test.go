package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

func TestGrantDelegationRequiresUnrestrictedAuthority(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	actor := domain.Principal{UserID: "manager-1", Role: domain.RoleManager}
	grant := domain.AccessGrant{
		SubjectType: "user", SubjectID: actor.UserID, Status: "active", ValidFrom: now,
		Items: []domain.AccessGrantItem{{PermissionCode: "guest.profile.read", Effect: "allow"}},
	}
	ceiling := []domain.AccessRolePermission{{PermissionCode: "guest.profile.read", Effect: "allow"}}
	type delegationCase struct {
		name   string
		grants []domain.AccessGrant
		want   bool
	}
	cases := []delegationCase{
		{"unrestricted", []domain.AccessGrant{grant}, true},
		{"no grant", nil, false},
	}
	roleBound := grant
	roleBound.Scopes = []domain.AccessGrantScope{{Type: "principal_role", Value: "manager"}}
	cases = append(cases, delegationCase{"manager-bound", []domain.AccessGrant{roleBound}, true})
	scoped := grant
	scoped.Scopes = []domain.AccessGrantScope{{Type: "tenant", Value: "tenant-1"}}
	cases = append(cases, delegationCase{"scoped authority", []domain.AccessGrant{grant, scoped}, false})
	denied := grant
	denied.Items = []domain.AccessGrantItem{{PermissionCode: "guest.profile.read", Effect: "deny"}}
	denied.Scopes = []domain.AccessGrantScope{{Type: "tenant", Value: "tenant-1"}}
	cases = append(cases, delegationCase{"scoped deny", []domain.AccessGrant{grant, denied}, false})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			policy := grantActorPolicy{now: now, ceiling: ceiling, grants: tc.grants}
			if got := policy.delegates(actor, "guest.profile.read"); got != tc.want {
				t.Fatalf("delegates = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestTenantAggregateDelegationRequiresBroaderRead(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	actor := domain.Principal{UserID: "manager-1", Role: domain.RoleManager}
	policy := grantActorPolicy{
		now:     now,
		ceiling: []domain.AccessRolePermission{{PermissionCode: "tenant.aggregates.read", Effect: "allow"}},
		grants: []domain.AccessGrant{{
			SubjectType: "role", SubjectID: "manager", Status: "active", ValidFrom: now,
			Items: []domain.AccessGrantItem{{PermissionCode: "tenant.aggregates.read", Effect: "allow"}},
		}},
	}
	if !policy.delegates(actor, "tenant.aggregates.own") {
		t.Fatal("manager with unrestricted aggregate access cannot delegate own aggregates")
	}
	policy.grants = nil
	if policy.delegates(actor, "tenant.aggregates.own") {
		t.Fatal("manager without an active aggregate grant can delegate own aggregates")
	}
}

func TestGrantScopesRejectInternalRole(t *testing.T) {
	_, err := toGrantScopes([]GrantScope{{Type: "principal_role", Value: "manager"}})
	if err == nil {
		t.Fatal("client supplied an internal principal role")
	}
	scopes, err := toGrantScopes([]GrantScope{{Type: "tenant", Value: "tenant-1"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(scopes) != 2 || scopes[1].Type != "principal_role" || scopes[1].Value != "tenant" {
		t.Fatalf("expected tenant role restriction, got %+v", scopes)
	}
}

type grantTestClock struct{ now time.Time }

func (c grantTestClock) Now() time.Time { return c.now }

type grantTestIDs struct{}

func (grantTestIDs) NewID(context.Context) string { return "audit-id" }

type grantTestRepository struct {
	grant domain.AccessGrant
}

func (r *grantTestRepository) ListPermissions(context.Context) ([]domain.Permission, error) {
	return []domain.Permission{{Code: "guest.profile.read", IsActive: true}}, nil
}

func (r *grantTestRepository) RolePermissions(_ context.Context, role domain.Role) ([]domain.AccessRolePermission, error) {
	if role == domain.RoleTenant {
		return []domain.AccessRolePermission{{PermissionCode: "guest.profile.read", Effect: "allow"}}, nil
	}
	return []domain.AccessRolePermission{{PermissionCode: "access.grant.write", Effect: "allow"}}, nil
}

func (r *grantTestRepository) PolicyVersion(context.Context) (int64, error) { return 1, nil }

func (r *grantTestRepository) ActiveGrants(_ context.Context, actor domain.Principal, now time.Time) ([]domain.AccessGrant, error) {
	return []domain.AccessGrant{{SubjectType: "role", SubjectID: "manager", Status: "active", ValidFrom: now,
		Items: []domain.AccessGrantItem{{PermissionCode: "access.grant.write", Effect: "allow"}}}}, nil
}

func (r *grantTestRepository) GetGrant(context.Context, string) (*domain.AccessGrant, error) {
	grant := r.grant
	return &grant, nil
}

func (r *grantTestRepository) ListGrants(context.Context, domain.AccessGrantFilter) ([]domain.AccessGrant, error) {
	return nil, nil
}

func (r *grantTestRepository) CreateGrant(context.Context, *domain.AccessGrant, int64, *domain.AuditLog) error {
	return nil
}

func (r *grantTestRepository) UpdateGrant(_ context.Context, grant *domain.AccessGrant, version int64, audit *domain.AuditLog) error {
	r.grant = *grant
	return nil
}

func TestManagerCanRemoveAllowAfterLosingDelegatedPermission(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	actor := domain.Principal{UserID: "manager-1", Role: domain.RoleManager}
	for _, action := range []string{"suspend", "revoke"} {
		t.Run(action, func(t *testing.T) {
			repo := &grantTestRepository{grant: domain.AccessGrant{
				ID: "grant-1", SubjectType: "tenant", SubjectID: "tenant-1", Status: "active", ValidFrom: now,
				Reason: "original reason", Items: []domain.AccessGrantItem{{PermissionCode: "guest.profile.read", Effect: "allow"}},
				Scopes: []domain.AccessGrantScope{{Type: "tenant", Value: "tenant-1"}, {Type: "principal_role", Value: "tenant"}},
			}}
			svc := NewAccessGrantService(repo, grantTestClock{now}, grantTestIDs{})
			if _, err := svc.Transition(context.Background(), actor, "grant-1", action, "withdraw access", "request-1"); err != nil {
				t.Fatalf("%s allow grant: %v", action, err)
			}
		})
	}
}

func TestManagerCannotRemoveDenyWithoutDelegatedPermission(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	actor := domain.Principal{UserID: "manager-1", Role: domain.RoleManager}
	repo := &grantTestRepository{grant: domain.AccessGrant{
		ID: "grant-1", SubjectType: "tenant", SubjectID: "tenant-1", Status: "active", ValidFrom: now,
		Reason: "original reason", Items: []domain.AccessGrantItem{{PermissionCode: "guest.profile.read", Effect: "deny"}},
		Scopes: []domain.AccessGrantScope{{Type: "tenant", Value: "tenant-1"}, {Type: "principal_role", Value: "tenant"}},
	}}
	svc := NewAccessGrantService(repo, grantTestClock{now}, grantTestIDs{})
	_, err := svc.Transition(context.Background(), actor, "grant-1", "revoke", "withdraw restriction", "request-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("revoke deny grant: got %v, want forbidden", err)
	}
}
