package service

import (
	"context"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type accessAuditTestRepository struct {
	domain.AuditRepo
	logs    []domain.AuditLog
	grantID string
	actorID string
	from    *time.Time
	to      *time.Time
	limit   int
	offset  int
}

func (r *accessAuditTestRepository) ListAccessAudit(_ context.Context, grantID, actorID string, from, to *time.Time, limit, offset int) ([]domain.AuditLog, error) {
	r.grantID, r.actorID, r.from, r.to, r.limit, r.offset = grantID, actorID, from, to, limit, offset
	return r.logs, nil
}

type accessAuditPolicyRepo struct{ domain.AccessRepository }

func (accessAuditPolicyRepo) PolicyVersion(context.Context) (int64, error) { return 1, nil }
func (accessAuditPolicyRepo) RolePermissions(context.Context, domain.Role) ([]domain.AccessRolePermission, error) {
	return []domain.AccessRolePermission{{PermissionCode: "audit.read", Effect: "allow"}}, nil
}
func (accessAuditPolicyRepo) ActiveGrants(_ context.Context, principal domain.Principal, now time.Time) ([]domain.AccessGrant, error) {
	return []domain.AccessGrant{{SubjectType: "role", SubjectID: string(principal.Role), Status: "active", ValidFrom: now,
		Items: []domain.AccessGrantItem{{PermissionCode: "audit.read", Effect: "allow"}}}}, nil
}

func TestAccessAuditPassesFiltersAndReturnsDeniedEntries(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	from, to := now.Add(-time.Hour), now.Add(time.Hour)
	audit := &accessAuditTestRepository{logs: []domain.AuditLog{
		{ID: "denied", Action: "access.denied", Details: `{"method":"GET","route":"/guests/{id}"}`, CreatedAt: now},
		{ID: "grant", Action: "access.grant.revoke", CreatedAt: now.Add(-time.Minute)},
	}}
	svc := NewAccessGrantService(accessAuditPolicyRepo{}, grantTestClock{now}, grantTestIDs{})
	svc.SetAuditRepository(audit)
	result, err := svc.Audit(context.Background(), domain.Principal{UserID: "manager", Role: domain.RoleManager}, "", "actor", &from, &to, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	if audit.grantID != "" || audit.actorID != "actor" || audit.from != &from || audit.to != &to || audit.limit != 10 || audit.offset != 3 {
		t.Fatalf("incorrect audit filters: %+v", audit)
	}
	if len(result) != 2 || result[0].Action != "access.denied" || result[1].Action != "access.grant.revoke" {
		t.Fatalf("incorrect audit entries: %+v", result)
	}
	if _, ok := result[0].Details.(map[string]any); !ok {
		t.Fatalf("audit details are not decoded: %+v", result[0].Details)
	}
	_, err = svc.Audit(context.Background(), domain.Principal{UserID: "manager", Role: domain.RoleManager}, "grant-1", "", nil, nil, 5, 0)
	if err != nil || audit.grantID != "grant-1" {
		t.Fatalf("grant filter not forwarded: grant=%q err=%v", audit.grantID, err)
	}
}
