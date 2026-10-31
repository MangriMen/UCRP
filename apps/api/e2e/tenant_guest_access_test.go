package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/harness"
	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/repository"
)

func TestTenantGuestDataAccess(t *testing.T) {
	s := harness.New(t).WithDB()
	w := buildWorld(t, s, "tenant-guest-access")
	ctx, cancel := s.Ctx()
	defer cancel()
	manager, err := s.Reg.User(ctx, "tenant-guest-manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}
	tenant, err := s.Reg.User(ctx, "tenant-guest-user", "tenant", w.tenantA.ID)
	if err != nil {
		t.Fatalf("fixture tenant: %v", err)
	}
	otherTenant, err := s.Reg.User(ctx, "tenant-guest-other-user", "tenant", w.tenantB.ID)
	if err != nil {
		t.Fatalf("fixture other tenant: %v", err)
	}
	otherSource, err := s.Reg.DataSource(ctx, "tenant-guest-other", w.tenantB.ID)
	if err != nil {
		t.Fatalf("fixture other source: %v", err)
	}
	guest, err := s.Reg.Guest(ctx, "tenant-guest-shared")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	otherGuest, err := s.Reg.Guest(ctx, "tenant-guest-other-only")
	if err != nil {
		t.Fatalf("fixture other guest: %v", err)
	}
	ownPhone := s.Reg.Phone(101)
	otherPhone := s.Reg.Phone(102)
	ownEmail := s.Reg.Name("private-email") + "@example.test"
	for _, identifier := range []struct{ kind, value, source string }{
		{"phone", ownPhone, w.source.ID},
		{"email", ownEmail, w.source.ID},
		{"phone", otherPhone, otherSource.ID},
	} {
		if _, err := s.Reg.Identifier(ctx, guest.ID, identifier.kind, identifier.value, identifier.source); err != nil {
			t.Fatalf("fixture identifier: %v", err)
		}
	}
	amount := int64(12345)
	ownEvent, err := s.Reg.Event(ctx, guest.ID, w.tenantA.ID, w.eventType.ID, w.source.ID, "tenant-guest-own", time.Now().UTC().Add(-time.Hour), &amount)
	if err != nil {
		t.Fatalf("fixture own event: %v", err)
	}
	otherEvent, err := s.Reg.Event(ctx, guest.ID, w.tenantB.ID, w.eventType.ID, otherSource.ID, "tenant-guest-other", time.Now().UTC().Add(-time.Hour), &amount)
	if err != nil {
		t.Fatalf("fixture other event: %v", err)
	}
	if _, err := s.Reg.Event(ctx, otherGuest.ID, w.tenantB.ID, w.eventType.ID, otherSource.ID, "tenant-guest-other-only", time.Now().UTC().Add(-time.Hour), &amount); err != nil {
		t.Fatalf("fixture other-only event: %v", err)
	}
	if err := s.DB.Exec(ctx, `UPDATE event SET raw_payload = $2 WHERE id = $1`, ownEvent.ID, "sensitive-raw-payload"); err != nil {
		t.Fatalf("fixture raw payload: %v", err)
	}

	managerGrantID := s.Reg.NewID()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		for _, query := range []string{
			`DELETE FROM access_grant_item WHERE grant_id IN (SELECT id FROM access_grant WHERE (subject_type = 'tenant' AND subject_id = $1) OR id = $2)`,
			`DELETE FROM access_grant_scope WHERE grant_id IN (SELECT id FROM access_grant WHERE (subject_type = 'tenant' AND subject_id = $1) OR id = $2)`,
			`DELETE FROM audit_log WHERE entity_id IN (SELECT id FROM access_grant WHERE (subject_type = 'tenant' AND subject_id = $1) OR id = $2)`,
			`DELETE FROM access_grant WHERE (subject_type = 'tenant' AND subject_id = $1) OR id = $2`,
		} {
			if err := s.DB.Exec(cleanupCtx, query, w.tenantA.ID, managerGrantID); err != nil {
				t.Errorf("cleanup access fixture: %v", err)
			}
		}
		if err := s.DB.Exec(cleanupCtx, `DELETE FROM audit_log WHERE actor_user_id IN ($1,$2,$3)`, manager.ID, tenant.ID, otherTenant.ID); err != nil {
			t.Errorf("cleanup access audit: %v", err)
		}
		if err := s.DB.Exec(cleanupCtx, `UPDATE access_policy_state SET version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = 1`); err != nil {
			t.Errorf("invalidate access policy after fixture cleanup: %v", err)
		}
	})

	repo := repository.NewAccessSQL(s.DB.DB)
	version, err := repo.PolicyVersion(ctx)
	if err != nil {
		t.Fatalf("read policy version: %v", err)
	}
	now := time.Now().UTC()
	managerGrant := &domain.AccessGrant{
		ID: managerGrantID, SubjectType: "user", SubjectID: manager.ID, Status: "active",
		ValidFrom: now.Add(-time.Minute), CreatedBy: manager.ID, Reason: "E2E fixture manager",
		CreatedAt: now, UpdatedAt: now,
		Scopes: []domain.AccessGrantScope{{Type: "principal_role", Value: "manager"}},
	}
	for _, permission := range []string{"guest.profile.read", "guest.event.read", "guest.identifier.read", "field.guest.identifiers.read", "field.guest.phone.read", "field.event.amount.read", "field.event.external_id.read"} {
		managerGrant.Items = append(managerGrant.Items, domain.AccessGrantItem{PermissionCode: permission, Effect: "allow"})
	}
	audit := &domain.AuditLog{ID: s.Reg.NewID(), ActorUserID: manager.ID, ActorRole: "manager", Action: "access.grant.create", EntityType: "access_grant", EntityID: managerGrantID, CreatedAt: now}
	if err := repo.CreateGrant(ctx, managerGrant, version, audit); err != nil {
		t.Fatalf("fixture manager access: %v", err)
	}

	managerClient := s.As(manager.Email, manager.Password)
	tenantClient := s.As(tenant.Email, tenant.Password)
	otherClient := s.As(otherTenant.Email, otherTenant.Password)
	profilePath := "/api/v1/guests/" + guest.ID
	eventsPath := profilePath + "/events"
	get := func(c *client.Client, path string, want int) *client.Response {
		t.Helper()
		resp, err := c.Get(ctx, path, nil)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		s.RequireStatus(resp, want, "GET "+path)
		return resp
	}
	create := func(permissions ...string) string {
		t.Helper()
		items := make([]map[string]string, 0, len(permissions))
		for _, permission := range permissions {
			items = append(items, map[string]string{"permission": permission, "effect": "allow"})
		}
		response, err := managerClient.Post(ctx, "/api/v1/access/grants", map[string]any{
			"subjectType": "tenant", "subjectId": w.tenantA.ID, "status": "active",
			"validFrom": now.Add(-time.Minute).Format(time.RFC3339), "validTo": now.Add(time.Hour).Format(time.RFC3339),
			"reason": "E2E tenant guest access", "items": items,
			"scopes": []map[string]string{{"type": "tenant", "value": w.tenantA.ID}},
		})
		if err != nil {
			t.Fatalf("create tenant grant: %v", err)
		}
		s.RequireStatus(response, 201, "create tenant grant")
		var grant struct {
			ID string `json:"id"`
		}
		if err := response.Decode(&grant); err != nil || grant.ID == "" {
			t.Fatalf("decode tenant grant: %v, %s", err, response.Snippet())
		}
		return grant.ID
	}
	get(tenantClient, profilePath, 403)
	baseGrant := create("guest.profile.read", "guest.event.read", "guest.identifier.read", "field.guest.identifiers.read", "field.guest.phone.read")
	profile := get(tenantClient, profilePath, 200)
	if !strings.Contains(string(profile.Body), ownPhone) {
		t.Fatalf("own authorized phone missing: %s", profile.Snippet())
	}
	for _, value := range []string{otherPhone, ownEmail, w.tenantB.Name, guest.DisplayName} {
		if strings.Contains(string(profile.Body), value) {
			t.Fatalf("profile exposes a hidden value: %s", profile.Snippet())
		}
	}
	events := get(tenantClient, eventsPath, 200)
	var eventBody struct {
		Items []struct {
			ID          string  `json:"id"`
			TenantID    string  `json:"tenantId"`
			AmountMinor *int64  `json:"amountMinor"`
			ExternalID  *string `json:"externalId"`
		} `json:"items"`
	}
	if err := events.Decode(&eventBody); err != nil {
		t.Fatalf("decode events: %v", err)
	}
	if len(eventBody.Items) != 1 || eventBody.Items[0].ID != ownEvent.ID || eventBody.Items[0].TenantID != w.tenantA.ID || eventBody.Items[0].AmountMinor != nil || eventBody.Items[0].ExternalID != nil {
		t.Fatalf("events must contain only own event with masked fields: %s", events.Snippet())
	}
	if strings.Contains(string(events.Body), otherEvent.ID) || strings.Contains(string(events.Body), "sensitive-raw-payload") {
		t.Fatalf("events expose another tenant or raw payload: %s", events.Snippet())
	}
	get(tenantClient, "/api/v1/guests/"+otherGuest.ID, 404)
	get(otherClient, profilePath, 403)
	create("field.event.amount.read", "field.event.external_id.read")
	events = get(tenantClient, eventsPath, 200)
	if err := events.Decode(&eventBody); err != nil || len(eventBody.Items) != 1 || eventBody.Items[0].AmountMinor == nil || *eventBody.Items[0].AmountMinor != amount || eventBody.Items[0].ExternalID == nil || *eventBody.Items[0].ExternalID != ownEvent.ExternalID {
		t.Fatalf("granted event fields are missing: %v, %s", err, events.Snippet())
	}
	revoked, err := managerClient.Post(ctx, fmt.Sprintf("/api/v1/access/grants/%s/revoke", baseGrant), map[string]string{"reason": "E2E revoke"})
	if err != nil {
		t.Fatalf("revoke tenant grant: %v", err)
	}
	s.RequireStatus(revoked, 200, "revoke tenant grant")
	get(tenantClient, profilePath, 403)
	get(tenantClient, eventsPath, 403)
}
