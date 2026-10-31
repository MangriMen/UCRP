package access

import (
	"reflect"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

func TestDecide(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	end := now.Add(time.Hour)
	principal := domain.Principal{UserID: "user-1", Role: domain.RoleTenant, TenantID: "tenant-1"}
	ceiling := []domain.AccessRolePermission{{PermissionCode: "read", Effect: "allow"}}
	grant := domain.AccessGrant{
		SubjectType: "tenant", SubjectID: "tenant-1", Status: "active", ValidFrom: now,
		ValidTo: &end, Items: []domain.AccessGrantItem{{PermissionCode: "read", Effect: "allow"}},
		Scopes: []domain.AccessGrantScope{{Type: "tenant", Value: "self"}},
	}
	cases := []struct {
		name      string
		principal domain.Principal
		ceiling   []domain.AccessRolePermission
		grants    []domain.AccessGrant
		resource  Resource
		now       time.Time
		allowed   bool
		reason    string
	}{
		{name: "valid from inclusive", principal: principal, ceiling: ceiling, grants: []domain.AccessGrant{grant}, resource: Resource{TenantID: "tenant-1"}, now: now, allowed: true},
		{name: "before end", principal: principal, ceiling: ceiling, grants: []domain.AccessGrant{grant}, resource: Resource{TenantID: "tenant-1"}, now: end.Add(-time.Nanosecond), allowed: true},
		{name: "before start", principal: principal, ceiling: ceiling, grants: []domain.AccessGrant{grant}, resource: Resource{TenantID: "tenant-1"}, now: now.Add(-time.Nanosecond), reason: "forbidden_permission"},
		{name: "valid to exclusive", principal: principal, ceiling: ceiling, grants: []domain.AccessGrant{grant}, resource: Resource{TenantID: "tenant-1"}, now: end, reason: "forbidden_permission"},
		{name: "missing grant", principal: principal, ceiling: ceiling, resource: Resource{TenantID: "tenant-1"}, now: now, reason: "forbidden_permission"},
		{name: "missing ceiling", principal: principal, grants: []domain.AccessGrant{grant}, now: now, reason: "forbidden_permission"},
		{name: "ceiling deny", principal: principal, ceiling: append(append([]domain.AccessRolePermission{}, ceiling...), domain.AccessRolePermission{PermissionCode: "read", Effect: "deny"}), grants: []domain.AccessGrant{grant}, now: now, reason: "forbidden_permission"},
		{name: "other tenant", principal: principal, ceiling: ceiling, grants: []domain.AccessGrant{grant}, resource: Resource{TenantID: "tenant-2"}, now: now, reason: "forbidden_tenant"},
		{name: "missing resource tenant", principal: principal, ceiling: ceiling, grants: []domain.AccessGrant{grant}, now: now, reason: "forbidden_tenant"},
		{name: "missing principal tenant", principal: domain.Principal{Role: domain.RoleTenant}, ceiling: ceiling, grants: []domain.AccessGrant{grant}, now: now, reason: "forbidden_tenant"},
		{name: "different subject", principal: domain.Principal{UserID: "user-2", Role: domain.RoleManager}, ceiling: ceiling, grants: []domain.AccessGrant{grant}, now: now, reason: "forbidden_permission"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Decide(tc.principal, "read", tc.resource, tc.ceiling, tc.grants, tc.now)
			if got.Allowed != tc.allowed || got.Reason != tc.reason {
				t.Fatalf("got %+v, want allowed=%t reason=%q", got, tc.allowed, tc.reason)
			}
		})
	}
}

func TestSubjectsAndScopes(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	principal := domain.Principal{UserID: "u", Role: domain.RoleTenant, TenantID: "t"}
	ceiling := []domain.AccessRolePermission{{PermissionCode: "read", Effect: "allow"}}
	grant := domain.AccessGrant{Status: "active", ValidFrom: now, Items: []domain.AccessGrantItem{{PermissionCode: "read", Effect: "allow"}}, Scopes: []domain.AccessGrantScope{{Type: "tenant", Value: "self"}}}
	for _, subject := range []struct{ kind, id string }{{"user", "u"}, {"tenant", "t"}, {"role", "tenant"}} {
		t.Run(subject.kind, func(t *testing.T) {
			g := grant
			g.SubjectType, g.SubjectID = subject.kind, subject.id
			if got := Decide(principal, "read", Resource{TenantID: "t"}, ceiling, []domain.AccessGrant{g}, now); !got.Allowed {
				t.Fatalf("got %+v", got)
			}
		})
	}
	grant.SubjectType, grant.SubjectID = "user", "u"
	unscoped := grant
	unscoped.Scopes = nil
	if got := Decide(principal, "read", Resource{TenantID: "t"}, ceiling, []domain.AccessGrant{unscoped}, now); got.Allowed {
		t.Fatalf("unscoped tenant grant allowed: %+v", got)
	}
	wrongTenant := grant
	wrongTenant.Scopes = []domain.AccessGrantScope{{Type: "tenant", Value: "other"}}
	if got := Decide(principal, "read", Resource{TenantID: "t"}, ceiling, []domain.AccessGrant{wrongTenant}, now); got.Allowed {
		t.Fatalf("foreign tenant grant allowed: %+v", got)
	}
	if got := Decide(principal, "read", Resource{TenantID: "t"}, ceiling, []domain.AccessGrant{grant}, now); !got.Allowed || got.KThreshold != 5 {
		t.Fatalf("base threshold not enforced: %+v", got)
	}
	grant.Scopes = []domain.AccessGrantScope{
		{Type: "tenant", Value: "self"}, {Type: "category", Value: "food"},
		{Type: "category", Value: "leisure"}, {Type: "channel", Value: "email"},
		{Type: "metric", Value: "visits"}, {Type: "metric", Value: "revenue"},
		{Type: "period_days", Value: "30"}, {Type: "row_limit", Value: "100"},
		{Type: "k_threshold", Value: "7"}, {Type: "field_mask", Value: "phone"},
	}
	resource := Resource{TenantID: "t", Category: "food", Channel: "email", Metric: "visits", PeriodDays: 30, RowCount: 100}
	got := Decide(principal, "read", resource, ceiling, []domain.AccessGrant{grant}, now)
	if !got.Allowed || got.PeriodDays != 30 || got.RowLimit != 100 || got.KThreshold != 7 || !reflect.DeepEqual(got.MaskFields, []string{"phone"}) {
		t.Fatalf("unexpected decision: %+v", got)
	}
	for _, tc := range []struct {
		name     string
		resource Resource
	}{
		{"wrong category", Resource{TenantID: "t", Category: "retail", Channel: "email", Metric: "visits"}},
		{"wrong channel", Resource{TenantID: "t", Category: "food", Channel: "sms", Metric: "visits"}},
		{"wrong metric", Resource{TenantID: "t", Category: "food", Channel: "email", Metric: "other"}},
		{"missing metric", Resource{TenantID: "t", Category: "food", Channel: "email"}},
		{"period too long", Resource{TenantID: "t", Category: "food", Channel: "email", Metric: "visits", PeriodDays: 31}},
		{"too many rows", Resource{TenantID: "t", Category: "food", Channel: "email", Metric: "visits", RowCount: 101}},
		{"unspecified tenant", Resource{Category: "food", Channel: "email", Metric: "visits"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Decide(principal, "read", tc.resource, ceiling, []domain.AccessGrant{grant}, now); got.Allowed {
				t.Fatalf("unexpected decision: %+v", got)
			}
		})
	}
	for _, scope := range []domain.AccessGrantScope{{Type: "unknown", Value: "x"}, {Type: "k_threshold", Value: "4"}, {Type: "period_days", Value: "-1"}, {Type: "field_mask", Value: "other"}} {
		g := grant
		g.Scopes = []domain.AccessGrantScope{{Type: "tenant", Value: "self"}, scope}
		if got := Decide(principal, "read", resource, ceiling, []domain.AccessGrant{g}, now); got.Allowed {
			t.Fatalf("scope %+v unexpectedly allowed: %+v", scope, got)
		}
	}
}

func TestDenyAndNarrowing(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	principal := domain.Principal{UserID: "u", Role: domain.RoleManager}
	ceiling := []domain.AccessRolePermission{{PermissionCode: "read", Effect: "allow"}}
	allow := domain.AccessGrant{SubjectType: "user", SubjectID: "u", Status: "active", ValidFrom: now,
		Items:  []domain.AccessGrantItem{{PermissionCode: "read", Effect: "allow"}},
		Scopes: []domain.AccessGrantScope{{Type: "period_days", Value: "30"}, {Type: "k_threshold", Value: "8"}, {Type: "field_mask", Value: "phone"}}}
	other := allow
	other.Scopes = []domain.AccessGrantScope{{Type: "period_days", Value: "60"}, {Type: "k_threshold", Value: "5"}, {Type: "field_mask", Value: "email"}}
	got := Decide(principal, "read", Resource{PeriodDays: 30}, ceiling, []domain.AccessGrant{allow, other}, now)
	if !got.Allowed || got.PeriodDays != 30 || got.KThreshold != 8 || !reflect.DeepEqual(got.MaskFields, []string{"email", "phone"}) {
		t.Fatalf("unexpected decision: %+v", got)
	}
	if got = Decide(principal, "read", Resource{PeriodDays: 31}, ceiling, []domain.AccessGrant{allow, other}, now); got.Allowed {
		t.Fatalf("narrower scope ignored: %+v", got)
	}
	deny := domain.AccessGrant{SubjectType: "role", SubjectID: "manager", Status: "active", ValidFrom: now,
		Items: []domain.AccessGrantItem{{PermissionCode: "read", Effect: "deny"}}, Scopes: []domain.AccessGrantScope{{Type: "tenant", Value: "t1"}}}
	for _, tenant := range []string{"t1", "t2"} {
		resource := Resource{TenantID: tenant}
		before := Decide(principal, "read", resource, ceiling, []domain.AccessGrant{allow}, now)
		after := Decide(principal, "read", resource, ceiling, []domain.AccessGrant{allow, deny}, now)
		if !before.Allowed || after.Allowed != (tenant == "t2") || (after.Allowed && !before.Allowed) {
			t.Fatalf("tenant %q: before=%+v after=%+v", tenant, before, after)
		}
	}
	deny.Scopes = []domain.AccessGrantScope{{Type: "period_days", Value: "10"}}
	if got := Decide(principal, "read", Resource{PeriodDays: 20}, ceiling, []domain.AccessGrant{allow, deny}, now); got.Allowed {
		t.Fatalf("deny with overlapping period ignored: %+v", got)
	}
}

func TestPrincipalRoleScope(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	ceiling := []domain.AccessRolePermission{{PermissionCode: "read", Effect: "allow"}}
	grant := domain.AccessGrant{
		SubjectType: "user", SubjectID: "u", Status: "active", ValidFrom: now,
		Items:  []domain.AccessGrantItem{{PermissionCode: "read", Effect: "allow"}},
		Scopes: []domain.AccessGrantScope{{Type: "principal_role", Value: "manager"}},
	}
	for _, tc := range []struct {
		role    domain.Role
		allowed bool
	}{
		{role: domain.RoleManager, allowed: true},
		{role: domain.RoleMarketer, allowed: false},
		{role: "", allowed: false},
	} {
		principal := domain.Principal{UserID: "u", Role: tc.role}
		got := Decide(principal, "read", Resource{}, ceiling, []domain.AccessGrant{grant}, now)
		if got.Allowed != tc.allowed {
			t.Fatalf("role %q: got %+v, want allowed=%t", tc.role, got, tc.allowed)
		}
	}
}
