package service

import (
	"context"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type refreshAccessRepository struct {
	domain.AccessRepository
}

func (refreshAccessRepository) PolicyVersion(context.Context) (int64, error) { return 7, nil }

func (refreshAccessRepository) RolePermissions(context.Context, domain.Role) ([]domain.AccessRolePermission, error) {
	return []domain.AccessRolePermission{{PermissionCode: "guest.profile.read", Effect: "allow"}}, nil
}

func (refreshAccessRepository) ActiveGrants(_ context.Context, principal domain.Principal, now time.Time) ([]domain.AccessGrant, error) {
	return []domain.AccessGrant{{
		SubjectType: "user", SubjectID: principal.UserID, Status: "active", ValidFrom: now.Add(-time.Hour),
		Items: []domain.AccessGrantItem{{PermissionCode: "guest.profile.read", Effect: "allow"}},
	}}, nil
}

func TestRefreshReturnsEffectivePermissions(t *testing.T) {
	manager := auth.NewManager("test-secret", time.Hour)
	tokens, err := manager.Issue(domain.Principal{UserID: "u", Role: domain.RoleAnalyst})
	if err != nil {
		t.Fatal(err)
	}
	users := delegationUsers{users: map[string]domain.User{"u": {ID: "u", Role: domain.RoleAnalyst, IsActive: true}}}
	svc := NewAuthService(users, manager, domain.SystemClock{}, idgen.UUID{})
	svc.SetAccessService(NewAccessGrantService(refreshAccessRepository{}, domain.SystemClock{}, idgen.UUID{}))

	result, err := svc.Refresh(context.Background(), tokens.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.User.Permissions) != 1 || result.User.Permissions[0] != "guest.profile.read" {
		t.Fatalf("refresh must return effective permissions, got %v", result.User.Permissions)
	}
	if result.User.PolicyVersion != 7 {
		t.Fatalf("refresh must return the policy version, got %d", result.User.PolicyVersion)
	}
}
