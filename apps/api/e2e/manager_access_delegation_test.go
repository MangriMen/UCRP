package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/harness"
	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/repository"
)

func TestManagerDelegatesPersonalAccess(t *testing.T) {
	s := harness.New(t).WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()
	issuer, err := s.Reg.User(ctx, "manager-delegation-issuer", "manager", "")
	if err != nil {
		t.Fatalf("fixture issuer: %v", err)
	}
	recipient, err := s.Reg.User(ctx, "manager-delegation-recipient", "manager", "")
	if err != nil {
		t.Fatalf("fixture recipient: %v", err)
	}
	issuerGrantID := s.Reg.NewID()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		for _, query := range []string{
			`DELETE FROM access_grant_item WHERE grant_id IN (SELECT id FROM access_grant WHERE subject_type = 'user' AND subject_id IN ($1, $2))`,
			`DELETE FROM access_grant_scope WHERE grant_id IN (SELECT id FROM access_grant WHERE subject_type = 'user' AND subject_id IN ($1, $2))`,
			`DELETE FROM audit_log WHERE entity_id IN (SELECT id FROM access_grant WHERE subject_type = 'user' AND subject_id IN ($1, $2))`,
			`DELETE FROM access_grant WHERE subject_type = 'user' AND subject_id IN ($1, $2)`,
			`DELETE FROM audit_log WHERE actor_user_id IN ($1, $2)`,
		} {
			if err := s.DB.Exec(cleanupCtx, query, issuer.ID, recipient.ID); err != nil {
				t.Errorf("cleanup manager access fixture: %v", err)
			}
		}
		if err := s.DB.Exec(cleanupCtx, `UPDATE access_policy_state SET version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = 1`); err != nil {
			t.Errorf("invalidate access policy: %v", err)
		}
	})
	repo := repository.NewAccessSQL(s.DB.DB)
	version, err := repo.PolicyVersion(ctx)
	if err != nil {
		t.Fatalf("read policy version: %v", err)
	}
	now := time.Now().UTC()
	grant := &domain.AccessGrant{ID: issuerGrantID, SubjectType: "user", SubjectID: issuer.ID,
		Status: "active", ValidFrom: now.Add(-time.Minute), CreatedBy: issuer.ID, Reason: "E2E fixture issuer",
		CreatedAt: now, UpdatedAt: now, Items: []domain.AccessGrantItem{{PermissionCode: "guest.profile.read", Effect: "allow"}},
		Scopes: []domain.AccessGrantScope{{Type: "principal_role", Value: "manager"}}}
	audit := &domain.AuditLog{ID: s.Reg.NewID(), ActorUserID: issuer.ID, ActorRole: "manager", Action: "access.grant.create", EntityType: "access_grant", EntityID: issuerGrantID, CreatedAt: now}
	if err := repo.CreateGrant(ctx, grant, version, audit); err != nil {
		t.Fatalf("fixture issuer permission: %v", err)
	}
	issuerClient := s.As(issuer.Email, issuer.Password)
	recipientClient := s.As(recipient.Email, recipient.Password)
	permissions := func(c *client.Client) map[string]bool {
		t.Helper()
		resp, err := c.Get(ctx, "/api/v1/auth/me", nil)
		if err != nil {
			t.Fatalf("GET auth/me: %v", err)
		}
		s.RequireStatus(resp, 200, "GET auth/me")
		var body struct {
			Permissions []string `json:"permissions"`
		}
		if err := resp.Decode(&body); err != nil {
			t.Fatalf("decode permissions: %v", err)
		}
		set := make(map[string]bool, len(body.Permissions))
		for _, permission := range body.Permissions {
			set[permission] = true
		}
		return set
	}
	if permissions(recipientClient)["guest.profile.read"] {
		t.Fatal("new manager has personal access before delegation")
	}
	request := map[string]any{
		"subjectType": "user", "subjectId": issuer.ID, "status": "active",
		"validFrom": now.Add(-time.Minute).Format(time.RFC3339), "validTo": now.Add(time.Hour).Format(time.RFC3339),
		"reason": "E2E manager delegation", "items": []map[string]string{{"permission": "guest.profile.read", "effect": "allow"}},
		"scopes": []map[string]string{},
	}
	response, err := issuerClient.Post(ctx, "/api/v1/access/grants", request)
	if err != nil {
		t.Fatalf("self-grant request: %v", err)
	}
	s.RequireStatus(response, 403, "self-grant denied")
	request["subjectId"] = recipient.ID
	response, err = issuerClient.Post(ctx, "/api/v1/access/grants", request)
	if err != nil {
		t.Fatalf("delegate request: %v", err)
	}
	s.RequireStatus(response, 201, "manager delegates access")
	var result struct {
		ID string `json:"id"`
	}
	if err := response.Decode(&result); err != nil || result.ID == "" {
		t.Fatalf("decode delegated grant: %v, %s", err, response.Snippet())
	}
	if !permissions(recipientClient)["guest.profile.read"] {
		t.Fatal("recipient did not acquire delegated access")
	}
	response, err = issuerClient.Post(ctx, "/api/v1/access/grants/"+result.ID+"/revoke", map[string]string{"reason": "E2E revoke"})
	if err != nil {
		t.Fatalf("revoke delegated grant: %v", err)
	}
	s.RequireStatus(response, 200, "revoke delegated grant")
	if permissions(recipientClient)["guest.profile.read"] {
		t.Fatal("revoked permission remained in the recipient's existing token")
	}
}
