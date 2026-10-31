package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type invitePolicyRepo struct {
	domain.AccessRepository
	now           time.Time
	issuerMissing string
	issuerExpires *time.Time
}

func (r invitePolicyRepo) PolicyVersion(context.Context) (int64, error) { return 2, nil }
func (r invitePolicyRepo) RolePermissions(context.Context, domain.Role) ([]domain.AccessRolePermission, error) {
	return []domain.AccessRolePermission{{PermissionCode: "user.write", Effect: "allow"}, {PermissionCode: "segment.write", Effect: "allow"}}, nil
}
func (r invitePolicyRepo) ActiveGrants(_ context.Context, p domain.Principal, _ time.Time) ([]domain.AccessGrant, error) {
	items := []domain.AccessGrantItem{{PermissionCode: "user.write", Effect: "allow"}, {PermissionCode: "segment.write", Effect: "allow"}}
	if p.UserID != "" && r.issuerMissing != "" {
		items = []domain.AccessGrantItem{{PermissionCode: r.issuerMissing, Effect: "allow"}}
	}
	grant := domain.AccessGrant{SubjectType: "role", SubjectID: "manager", Status: "active", ValidFrom: r.now.Add(-time.Hour), Items: items}
	if p.UserID != "" {
		grant.ValidTo = r.issuerExpires
	}
	return []domain.AccessGrant{grant}, nil
}

func TestCanInviteRequiresPermanentAuthority(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	actor := domain.Principal{UserID: "manager", Role: domain.RoleManager}
	end := now.Add(time.Hour)
	for _, tc := range []struct {
		name, missing string
		expires       *time.Time
		want          error
	}{
		{"matching authority", "", nil, nil},
		{"missing role permission", "user.write", nil, domain.ErrForbidden},
		{"temporary authority", "", &end, domain.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewAccessGrantService(invitePolicyRepo{now: now, issuerMissing: tc.missing, issuerExpires: tc.expires}, grantTestClock{now}, grantTestIDs{})
			_, err := svc.CanInvite(context.Background(), actor, domain.RoleManager, "")
			if !errors.Is(err, tc.want) {
				t.Fatalf("CanInvite returned %v, want %v", err, tc.want)
			}
		})
	}
}
