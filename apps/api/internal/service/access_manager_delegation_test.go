package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type delegationUsers struct {
	domain.UserRepo
	users map[string]domain.User
}

func (r delegationUsers) GetByID(_ context.Context, id string) (*domain.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &user, nil
}

type delegationRepository struct {
	domain.AccessRepository
	now      time.Time
	expires  *time.Time
	conflict bool
	created  *domain.AccessGrant
	grant    *domain.AccessGrant
}

func (r *delegationRepository) PolicyVersion(context.Context) (int64, error) { return 4, nil }

func (r *delegationRepository) ListPermissions(context.Context) ([]domain.Permission, error) {
	return []domain.Permission{{Code: "guest.profile.read", IsActive: true}, {Code: "access.grant.write", IsActive: true}}, nil
}

func (r *delegationRepository) RolePermissions(context.Context, domain.Role) ([]domain.AccessRolePermission, error) {
	return []domain.AccessRolePermission{{PermissionCode: "access.grant.read", Effect: "allow"}, {PermissionCode: "access.grant.write", Effect: "allow"}, {PermissionCode: "guest.profile.read", Effect: "allow"}}, nil
}

func (r *delegationRepository) ActiveGrants(_ context.Context, p domain.Principal, _ time.Time) ([]domain.AccessGrant, error) {
	return []domain.AccessGrant{
		{SubjectType: "role", SubjectID: string(p.Role), Status: "active", ValidFrom: r.now.Add(-time.Hour), Items: []domain.AccessGrantItem{{PermissionCode: "access.grant.read", Effect: "allow"}, {PermissionCode: "access.grant.write", Effect: "allow"}}},
		{SubjectType: "user", SubjectID: p.UserID, Status: "active", ValidFrom: r.now.Add(-time.Hour), ValidTo: r.expires,
			Items: []domain.AccessGrantItem{{PermissionCode: "guest.profile.read", Effect: "allow"}}, Scopes: []domain.AccessGrantScope{{Type: "principal_role", Value: "manager"}}},
	}, nil
}

func (r *delegationRepository) CreateGrant(_ context.Context, grant *domain.AccessGrant, version int64, _ *domain.AuditLog) error {
	if r.conflict || version != 4 {
		return domain.ErrConflict
	}
	r.created = grant
	return nil
}

func (r *delegationRepository) GetGrant(context.Context, string) (*domain.AccessGrant, error) {
	return r.grant, nil
}

func (r *delegationRepository) UpdateGrant(_ context.Context, grant *domain.AccessGrant, version int64, _ *domain.AuditLog) error {
	if r.conflict || version != 4 {
		return domain.ErrConflict
	}
	r.grant = grant
	return nil
}

func TestManagerPersonalGrantDelegation(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)
	end := now.Add(30 * time.Minute)
	actor := domain.Principal{UserID: "manager-a", Role: domain.RoleManager}
	users := delegationUsers{users: map[string]domain.User{
		"manager-a": {ID: "manager-a", Role: domain.RoleManager, IsActive: true},
		"manager-b": {ID: "manager-b", Role: domain.RoleManager, IsActive: true},
		"analyst":   {ID: "analyst", Role: domain.RoleAnalyst, IsActive: true},
	}}
	request := GrantRequest{SubjectType: "user", SubjectID: "manager-b", Status: "active", ValidFrom: now,
		ValidTo: &end, Reason: "delegate profile", Items: []GrantItem{{Permission: "guest.profile.read", Effect: "allow"}}}
	for _, test := range []struct {
		name string
		edit func(*GrantRequest, *delegationRepository)
		want error
	}{
		{"bounded delegation", func(*GrantRequest, *delegationRepository) {}, nil},
		{"self grant", func(r *GrantRequest, _ *delegationRepository) { r.SubjectID = "manager-a" }, domain.ErrForbidden},
		{"other role", func(r *GrantRequest, _ *delegationRepository) { r.SubjectID = "analyst" }, domain.ErrForbidden},
		{"management permission", func(r *GrantRequest, _ *delegationRepository) { r.Items[0].Permission = "access.grant.write" }, domain.ErrValidation},
		{"unbounded beyond issuer", func(r *GrantRequest, _ *delegationRepository) { r.ValidTo = nil; r.Indefinite = true }, domain.ErrForbidden},
		{"external scope", func(r *GrantRequest, _ *delegationRepository) {
			r.Scopes = []GrantScope{{Type: "tenant", Value: "any"}}
		}, domain.ErrValidation},
		{"version conflict", func(_ *GrantRequest, repo *delegationRepository) { repo.conflict = true }, domain.ErrConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &delegationRepository{now: now, expires: &expires}
			svc := NewAccessGrantService(repo, grantTestClock{now}, grantTestIDs{}, users)
			current := request
			current.Items = append([]GrantItem(nil), request.Items...)
			test.edit(&current, repo)
			_, err := svc.Create(context.Background(), actor, current, "request")
			if !errors.Is(err, test.want) {
				t.Fatalf("Create returned %v, want %v", err, test.want)
			}
			if test.want == nil && (repo.created == nil || len(repo.created.Scopes) != 1 || repo.created.Scopes[0] != (domain.AccessGrantScope{Type: "principal_role", Value: "manager"})) {
				t.Fatalf("manager grant is not bound to role: %+v", repo.created)
			}
		})
	}
}

func TestDelegatedPermissionCannotOutliveIssuer(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)
	later := expires.Add(time.Second)
	actor := domain.Principal{UserID: "manager-a", Role: domain.RoleManager}
	policy := grantActorPolicy{now: now, ceiling: []domain.AccessRolePermission{{PermissionCode: "guest.profile.read", Effect: "allow"}},
		grants: []domain.AccessGrant{{SubjectType: "user", SubjectID: actor.UserID, Status: "active", ValidFrom: now,
			ValidTo: &expires, Items: []domain.AccessGrantItem{{PermissionCode: "guest.profile.read", Effect: "allow"}}}}}
	if !policy.delegatesForPeriod(actor, "guest.profile.read", &expires) || policy.delegatesForPeriod(actor, "guest.profile.read", &later) || policy.delegatesForPeriod(actor, "guest.profile.read", nil) {
		t.Fatal("delegated time window exceeds issuer's permission")
	}
}

func TestGetGrant_ReturnsGrant_WhenSubjectUserIsDeactivated(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	actor := domain.Principal{UserID: "manager-a", Role: domain.RoleManager}
	users := delegationUsers{users: map[string]domain.User{
		"manager-b": {ID: "manager-b", Role: domain.RoleManager, IsActive: false},
	}}
	repo := &delegationRepository{now: now, grant: &domain.AccessGrant{ID: "grant-1", SubjectType: "user", SubjectID: "manager-b", Status: "active", ValidFrom: now}}
	svc := NewAccessGrantService(repo, grantTestClock{now}, grantTestIDs{}, users)
	grant, err := svc.Get(context.Background(), actor, "grant-1")
	if err != nil || grant.ID != "grant-1" {
		t.Fatalf("Get returned %+v, %v", grant, err)
	}
}

func TestPermissions_MarksGrantableSubjects_ByIssuerAuthority(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	actor := domain.Principal{UserID: "manager-a", Role: domain.RoleManager}
	svc := NewAccessGrantService(&delegationRepository{now: now}, grantTestClock{now}, grantTestIDs{})
	items, err := svc.Permissions(context.Background(), actor)
	if err != nil {
		t.Fatal(err)
	}
	grantable := make(map[string][]string, len(items))
	for _, item := range items {
		grantable[item.Code] = item.GrantableTo
	}
	if got := grantable["guest.profile.read"]; len(got) != 2 || got[0] != "tenant" || got[1] != "manager" {
		t.Fatalf("guest.profile.read grantable to %v, want tenant and manager", got)
	}
	if got := grantable["access.grant.write"]; len(got) != 1 || got[0] != "tenant" {
		t.Fatalf("access.grant.write grantable to %v, want tenant only", got)
	}
}
