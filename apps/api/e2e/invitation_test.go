package e2e

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/harness"
	"github.com/google/uuid"
)

func cleanupInvitationArtifacts(t *testing.T, s *harness.Suite, emails ...string) {
	t.Helper()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, email := range emails {
			if err := s.DB.Exec(cleanupCtx, `
				DELETE FROM audit_log WHERE actor_user_id IN (SELECT id FROM app_user WHERE email = $1)
					OR entity_id IN (SELECT id FROM user_invitation WHERE email = $1)`, email); err != nil {
				t.Errorf("cleanup audit for %s: %v", email, err)
			}
			if err := s.DB.Exec(cleanupCtx, `DELETE FROM user_invitation WHERE email = $1`, email); err != nil {
				t.Errorf("cleanup invitations for %s: %v", email, err)
			}
			if err := s.DB.Exec(cleanupCtx, `DELETE FROM app_user WHERE email = $1`, email); err != nil {
				t.Errorf("cleanup user %s: %v", email, err)
			}
		}
	})
}

func insertInvitation(t *testing.T, s *harness.Suite, creatorID, email, displayName, role, tenantID, purpose, targetUserID string, targetAuthVersion any, tokenHash []byte) {
	t.Helper()
	ctx, cancel := s.Ctx()
	defer cancel()
	var version int64
	if err := s.DB.QueryRowContext(ctx, `SELECT version FROM access_policy_state WHERE id = 1`).Scan(&version); err != nil {
		t.Fatalf("policy version: %v", err)
	}
	var tenantArg any
	if tenantID != "" {
		tenantArg = tenantID
	}
	var targetUserArg, targetVersionArg any
	if targetUserID != "" {
		targetUserArg = targetUserID
		targetVersionArg = targetAuthVersion
	}
	now := time.Now().UTC()
	if err := s.DB.Exec(ctx, `
		INSERT INTO user_invitation (id, email, display_name, role, tenant_id, purpose, target_user_id, target_auth_version, token_hash, created_by, policy_version, created_at, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		uuid.NewString(), email, displayName, role, tenantArg, purpose, targetUserArg, targetVersionArg, tokenHash, creatorID, version, now, now.Add(24*time.Hour)); err != nil {
		t.Fatalf("insert invitation: %v", err)
	}
}

func generateToken(t *testing.T) (string, []byte) {
	t.Helper()
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		t.Fatalf("generate token: %v", err)
	}
	hash := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(raw), hash[:]
}

func anonymousClient(s *harness.Suite) *client.Client {
	return client.New(s.Cfg.BaseURL, s.Cfg.RequestTimeout)
}

func TestInvitationAcceptCreatesTenantUser(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "invitation-accept")
	ctx, cancel := s.Ctx()
	defer cancel()

	manager, err := s.Reg.User(ctx, "inv-manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}

	email := "invited-" + s.Cfg.RunID + "@e2e.local"
	cleanupInvitationArtifacts(t, s, email)
	token, hash := generateToken(t)
	insertInvitation(t, s, manager.ID, email, "Invited Employee", "tenant", w.tenantA.ID, "invite", "", nil, hash)

	resp, err := anonymousClient(s).Post(ctx, "/api/v1/auth/invitations/accept", map[string]string{
		"token":    token,
		"password": "invited-password-1",
	})
	if err != nil {
		t.Fatalf("accept invitation: %v", err)
	}
	s.RequireStatus(resp, 204, "accept invitation")

	client := s.As(email, "invited-password-1")
	meResp, err := client.Get(ctx, "/api/v1/auth/me", nil)
	if err != nil {
		t.Fatalf("login as invited user: %v", err)
	}
	s.RequireStatus(meResp, 200, "login as invited user")
	var me struct {
		Role     string `json:"role"`
		TenantID string `json:"tenantId"`
	}
	if err := meResp.Decode(&me); err != nil {
		t.Fatalf("decode me: %v", err)
	}
	if me.Role != "tenant" || me.TenantID != w.tenantA.ID {
		t.Fatalf("unexpected identity: %+v", me)
	}
}

func TestInvitationAcceptSingleUse(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "invitation-single")
	ctx, cancel := s.Ctx()
	defer cancel()

	manager, err := s.Reg.User(ctx, "inv2-manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}

	email := "invited2-" + s.Cfg.RunID + "@e2e.local"
	cleanupInvitationArtifacts(t, s, email)
	token, hash := generateToken(t)
	insertInvitation(t, s, manager.ID, email, "Single Use", "tenant", w.tenantA.ID, "invite", "", nil, hash)

	anon := anonymousClient(s)
	resp, err := anon.Post(ctx, "/api/v1/auth/invitations/accept", map[string]string{"token": token, "password": "first-password-1"})
	if err != nil {
		t.Fatalf("first accept: %v", err)
	}
	s.RequireStatus(resp, 204, "first accept")

	resp, err = anon.Post(ctx, "/api/v1/auth/invitations/accept", map[string]string{"token": token, "password": "second-password-1"})
	if err != nil {
		t.Fatalf("second accept: %v", err)
	}
	s.RequireStatus(resp, 401, "second accept with same token")
}

func TestInvitationAcceptRejectsInvalidToken(t *testing.T) {
	s := harness.New(t).WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := anonymousClient(s).Post(ctx, "/api/v1/auth/invitations/accept", map[string]string{
		"token":    "not-a-real-token",
		"password": "any-password-1",
	})
	if err != nil {
		t.Fatalf("accept invalid token: %v", err)
	}
	s.RequireStatus(resp, 401, "accept invalid token")
}

func TestRecoveryInvalidatesOldPassword(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	w := buildWorld(t, s, "invitation-recovery")
	ctx, cancel := s.Ctx()
	defer cancel()

	manager, err := s.Reg.User(ctx, "rec-manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}
	employee, err := s.Reg.User(ctx, "rec-employee", "tenant", w.tenantA.ID)
	if err != nil {
		t.Fatalf("fixture employee: %v", err)
	}
	cleanupInvitationArtifacts(t, s, employee.Email)

	token, hash := generateToken(t)
	insertInvitation(t, s, manager.ID, employee.Email, employee.Email, "tenant", w.tenantA.ID, "recovery", employee.ID, int64(0), hash)

	anon := anonymousClient(s)
	resp, err := anon.Post(ctx, "/api/v1/auth/invitations/accept", map[string]string{
		"token":    token,
		"password": "recovered-password-1",
	})
	if err != nil {
		t.Fatalf("accept recovery: %v", err)
	}
	s.RequireStatus(resp, 204, "accept recovery")

	oldResp, err := anonymousClient(s).Login(ctx, employee.Email, employee.Password)
	if err != nil {
		t.Fatalf("old password after recovery: %v", err)
	}
	s.RequireStatus(oldResp, 401, "old password after recovery")

	newClient := s.As(employee.Email, "recovered-password-1")
	newResp, err := newClient.Get(ctx, "/api/v1/auth/me", nil)
	if err != nil {
		t.Fatalf("new password after recovery: %v", err)
	}
	s.RequireStatus(newResp, 200, "new password after recovery")
}
