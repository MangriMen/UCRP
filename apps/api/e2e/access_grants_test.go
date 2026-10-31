package e2e

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestAccessGrantLifecycle(t *testing.T) {
	s := harness.New(t).WithDB()
	w := buildWorld(t, s, "access-grants")
	ctx, cancel := s.Ctx()
	defer cancel()

	manager, err := s.Reg.User(ctx, "access-manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}
	tenant, err := s.Reg.User(ctx, "access-tenant", "tenant", w.tenantA.ID)
	if err != nil {
		t.Fatalf("fixture tenant user: %v", err)
	}
	var privateValues []string
	for i := 0; i < 6; i++ {
		guest, err := s.Reg.Guest(ctx, fmt.Sprintf("access-guest-%d", i))
		if err != nil {
			t.Fatalf("fixture guest: %v", err)
		}
		phone := s.Reg.Phone(i)
		if _, err := s.Reg.Identifier(ctx, guest.ID, "phone", phone, w.source.ID); err != nil {
			t.Fatalf("fixture identifier: %v", err)
		}
		w.visit(t, s, guest.ID, w.tenantA.ID, fmt.Sprintf("access-event-%d", i), 1)
		privateValues = append(privateValues, guest.ID, phone)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		var exists bool
		if err := s.DB.QueryRowContext(cleanupCtx, `SELECT to_regclass('access_grant') IS NOT NULL`).Scan(&exists); err != nil {
			t.Errorf("check access grant table: %v", err)
			return
		}
		if !exists {
			return
		}
		queries := []string{
			`DELETE FROM access_grant_item WHERE grant_id IN (SELECT id FROM access_grant WHERE subject_type = 'tenant' AND subject_id = $1)`,
			`DELETE FROM access_grant_scope WHERE grant_id IN (SELECT id FROM access_grant WHERE subject_type = 'tenant' AND subject_id = $1)`,
			`DELETE FROM audit_log WHERE entity_id IN (SELECT id FROM access_grant WHERE subject_type = 'tenant' AND subject_id = $1)`,
			`DELETE FROM access_grant WHERE subject_type = 'tenant' AND subject_id = $1`,
		}
		for _, query := range queries {
			if err := s.DB.Exec(cleanupCtx, query, w.tenantA.ID); err != nil {
				t.Errorf("cleanup access grants: %v", err)
			}
		}
		if err := s.DB.Exec(cleanupCtx, `DELETE FROM audit_log WHERE actor_user_id = $1 OR actor_user_id = $2`, manager.ID, tenant.ID); err != nil {
			t.Errorf("cleanup access audit: %v", err)
		}
	})

	managerClient := s.As(manager.Email, manager.Password)
	tenantClient := s.As(tenant.Email, tenant.Password)
	ownPath := "/api/v1/tenants/" + w.tenantA.ID + "/aggregates"
	otherPath := "/api/v1/tenants/" + w.tenantB.ID + "/aggregates"
	ownReferencePath := "/api/v1/tenants/" + w.tenantA.ID
	otherReferencePath := "/api/v1/tenants/" + w.tenantB.ID
	grantPath := "/api/v1/access/grants"

	getStatus := func(path string, want int) *client.Response {
		t.Helper()
		requestCtx, requestCancel := s.Ctx()
		defer requestCancel()
		resp, err := tenantClient.Get(requestCtx, path, nil)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		s.RequireStatus(resp, want, "GET "+path+" as tenant")
		return resp
	}
	grantBody := func(from, to time.Time) map[string]any {
		return map[string]any{
			"subjectType": "tenant",
			"subjectId":   w.tenantA.ID,
			"status":      "active",
			"validFrom":   from.UTC().Format(time.RFC3339),
			"validTo":     to.UTC().Format(time.RFC3339),
			"reason":      "E2E access test",
			"items": []map[string]string{
				{"permission": "tenant.aggregates.own", "effect": "allow"},
				{"permission": "metric.visits_count.read", "effect": "allow"},
				{"permission": "reference.read", "effect": "allow"},
			},
			"scopes": []map[string]string{{"type": "tenant", "value": w.tenantA.ID}},
		}
	}
	createGrant := func(body map[string]any) string {
		t.Helper()
		requestCtx, requestCancel := s.Ctx()
		defer requestCancel()
		resp, err := managerClient.Post(requestCtx, grantPath, body)
		if err != nil {
			t.Fatalf("create grant: %v", err)
		}
		s.RequireStatus(resp, 201, "manager creates grant")
		var result struct {
			ID string `json:"id"`
		}
		if err := resp.Decode(&result); err != nil {
			t.Fatalf("decode grant: %v", err)
		}
		if result.ID == "" {
			t.Fatalf("created grant has no id: %s", resp.Snippet())
		}
		return result.ID
	}

	getStatus(ownPath, 403)
	getStatus("/api/v1/tenants", 403)
	getStatus(ownReferencePath, 403)
	now := time.Now().UTC()
	unauthorized, err := tenantClient.Post(ctx, grantPath, grantBody(now.Add(-time.Minute), now.Add(time.Hour)))
	if err != nil {
		t.Fatalf("tenant creates grant: %v", err)
	}
	s.RequireStatus(unauthorized, 403, "tenant cannot create grant")

	createGrant(grantBody(now.Add(time.Hour), now.Add(2*time.Hour)))
	getStatus(ownPath, 403)
	getStatus(ownReferencePath, 403)

	grantID := createGrant(grantBody(now.Add(-time.Minute), now.Add(time.Hour)))
	page, err := managerClient.Get(ctx, grantPath, url.Values{"subject_type": {"tenant"}, "subject_id": {w.tenantA.ID}, "page": {"2"}, "page_size": {"1"}})
	if err != nil {
		t.Fatalf("list grants: %v", err)
	}
	s.RequireStatus(page, 200, "manager lists tenant grants")
	var pageBody struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
		Meta struct {
			Page  int `json:"page"`
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := page.Decode(&pageBody); err != nil {
		t.Fatalf("decode grant page: %v", err)
	}
	if len(pageBody.Items) != 1 || pageBody.Meta.Page != 2 || pageBody.Meta.Total != 2 {
		t.Fatalf("grant page must report the honest total: %s", page.Snippet())
	}
	unknownTenant := grantBody(now.Add(-time.Minute), now.Add(time.Hour))
	unknownTenant["subjectId"] = s.Reg.NewID()
	unknownTenant["scopes"] = []map[string]string{{"type": "tenant", "value": unknownTenant["subjectId"].(string)}}
	missing, err := managerClient.Post(ctx, grantPath, unknownTenant)
	if err != nil {
		t.Fatalf("grant for unknown tenant: %v", err)
	}
	s.RequireStatus(missing, 404, "grant for unknown tenant")
	own := getStatus(ownPath, 200)
	getStatus(ownPath+"?metrics=visits_count", 200)
	getStatus(ownPath+"?metrics=metric.cross_tenant_distribution.read", 403)
	getStatus(otherPath, 403)
	getStatus(otherReferencePath, 403)
	list := getStatus("/api/v1/tenants", 200)
	listBody, err := list.JSON()
	if err != nil {
		t.Fatalf("decode tenant list: %v", err)
	}
	items, ok := harness.Items(listBody)
	if !ok || len(items) != 1 {
		t.Fatalf("tenant list must contain only own tenant: %s", list.Snippet())
	}
	listItem, ok := items[0].(map[string]any)
	if !ok || listItem["id"] != w.tenantA.ID {
		t.Fatalf("tenant list contains a different tenant: %s", list.Snippet())
	}
	if _, ok := listItem["priority"]; ok {
		t.Fatalf("tenant list reveals priority: %s", list.Snippet())
	}
	detail := getStatus(ownReferencePath, 200)
	detailBody, err := detail.JSON()
	if err != nil {
		t.Fatalf("decode tenant reference: %v", err)
	}
	detailItem, ok := detailBody.(map[string]any)
	if !ok || detailItem["id"] != w.tenantA.ID {
		t.Fatalf("tenant reference does not match own tenant: %s", detail.Snippet())
	}
	if _, ok := detailItem["priority"]; ok {
		t.Fatalf("tenant reference reveals priority: %s", detail.Snippet())
	}

	body, err := own.JSON()
	if err != nil {
		t.Fatalf("decode tenant aggregates: %v", err)
	}
	assertNoGuestData(t, body, privateValues)

	revoked, err := managerClient.Post(ctx, grantPath+"/"+grantID+"/revoke", map[string]string{"reason": "E2E revocation"})
	if err != nil {
		t.Fatalf("revoke grant: %v", err)
	}
	s.RequireStatus(revoked, 200, "manager revokes grant")
	s.WaitFor(60*time.Second, "revocation to affect the existing token", func(pollCtx context.Context) (bool, error) {
		resp, err := tenantClient.Get(pollCtx, ownPath, nil)
		if err != nil {
			return false, err
		}
		return resp.Status == 403, nil
	})
	getStatus(ownReferencePath, 403)
	getStatus("/api/v1/tenants", 403)

	now = time.Now().UTC()
	createGrant(grantBody(now.Add(-time.Minute), now.Add(15*time.Second)))
	getStatus(ownPath, 200)
	s.WaitFor(25*time.Second, "expired grant to stop authorizing", func(pollCtx context.Context) (bool, error) {
		resp, err := tenantClient.Get(pollCtx, ownPath, nil)
		if err != nil {
			return false, err
		}
		return resp.Status == 403, nil
	})
	getStatus(ownReferencePath, 403)
}

func assertNoGuestData(t *testing.T, value any, privateValues []string) {
	t.Helper()
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			normalized := strings.ToLower(strings.ReplaceAll(key, "_", ""))
			switch normalized {
			case "guestid", "phone", "email", "identifiers":
				t.Fatalf("tenant aggregates contain private field %q", key)
			}
			assertNoGuestData(t, item, privateValues)
		}
	case []any:
		for _, item := range v {
			assertNoGuestData(t, item, privateValues)
		}
	case string:
		for _, private := range privateValues {
			if private != "" && strings.Contains(v, private) {
				t.Fatal("tenant aggregates contain a guest identifier")
			}
		}
	}
}

func TestAccessGrantStaffRoleGrantCannotBeChanged(t *testing.T) {
	s := harness.New(t).WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	manager, err := s.Reg.User(ctx, "access-role-grant-manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}
	managerClient := s.As(manager.Email, manager.Password)
	for _, action := range []string{"suspend", "revoke"} {
		resp, err := managerClient.Post(ctx, "/api/v1/access/grants/bootstrap-role-marketer/"+action, map[string]string{"reason": "E2E staff role grant"})
		if err != nil {
			t.Fatalf("%s staff role grant: %v", action, err)
		}
		s.RequireStatus(resp, 403, "manager cannot "+action+" a staff role grant")
	}
	grant, err := managerClient.Get(ctx, "/api/v1/access/grants/bootstrap-role-marketer", nil)
	if err != nil {
		t.Fatalf("get staff role grant: %v", err)
	}
	s.RequireStatus(grant, 200, "manager reads staff role grant")
	var body struct {
		Status string `json:"status"`
	}
	if err := grant.Decode(&body); err != nil {
		t.Fatalf("decode staff role grant: %v", err)
	}
	if body.Status != "active" {
		t.Fatalf("staff role grant status = %q, want active", body.Status)
	}
}
