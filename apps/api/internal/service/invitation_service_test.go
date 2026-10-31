package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

type invitationTestClock struct{ now time.Time }

func (c *invitationTestClock) Now() time.Time { return c.now }

type invitationTestRepo struct {
	invite            domain.UserInvitation
	revoked, consumed bool
	passwordHash      string
}

func (r *invitationTestRepo) CreateInvitation(_ context.Context, invite domain.UserInvitation) error {
	r.invite = invite
	return nil
}
func (r *invitationTestRepo) RevokeInvitation(_ context.Context, _ string, _ time.Time) error {
	r.revoked = true
	return nil
}
func (r *invitationTestRepo) ValidInvitation(_ context.Context, hash []byte, now time.Time) (bool, error) {
	return !r.revoked && !r.consumed && now.Before(r.invite.ExpiresAt) && bytes.Equal(hash, r.invite.TokenHash), nil
}
func (r *invitationTestRepo) AcceptInvitation(ctx context.Context, hash []byte, _, passwordHash, _ string, now time.Time) error {
	valid, _ := r.ValidInvitation(ctx, hash, now)
	if !valid {
		return domain.ErrUnauthorized
	}
	r.consumed, r.passwordHash = true, passwordHash
	return nil
}

type invitationTestAuthorizer struct{ err error }

func (a invitationTestAuthorizer) CanInvite(context.Context, domain.Principal, domain.Role, string) (int64, error) {
	return 12, a.err
}

type invitationTestSender struct {
	token string
	err   error
}

func (s *invitationTestSender) SendInvitation(_ context.Context, _, token string, _ time.Time) error {
	s.token = token
	return s.err
}

func TestInvitationLifecycle(t *testing.T) {
	ctx := context.Background()
	clock := &invitationTestClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	repo, sender := &invitationTestRepo{}, &invitationTestSender{}
	svc := NewInvitationService(repo, invitationTestAuthorizer{}, sender, clock, idgen.UUID{})
	actor := domain.Principal{UserID: "manager", Role: domain.RoleManager}
	if err := svc.Issue(ctx, actor, "employee@example.test", "Employee", domain.RoleTenant, "tenant"); err != nil {
		t.Fatal(err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(sender.token)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(raw)
	if !bytes.Equal(repo.invite.TokenHash, hash[:]) || bytes.Equal(repo.invite.TokenHash, raw) || repo.invite.PolicyVersion != 12 {
		t.Fatal("invitation token was not stored as a policy-bound hash")
	}
	if err := svc.Accept(ctx, sender.token, "employee-password"); err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword(repo.passwordHash, "employee-password") {
		t.Fatal("password was not hashed correctly")
	}
	if err := svc.Accept(ctx, sender.token, "employee-password"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("reused token: %v", err)
	}
}

func TestInvitationDisabledExpiredAndFailedDelivery(t *testing.T) {
	ctx := context.Background()
	clock := &invitationTestClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	repo := &invitationTestRepo{}
	actor := domain.Principal{UserID: "manager", Role: domain.RoleManager}
	disabled := NewInvitationService(repo, invitationTestAuthorizer{}, nil, clock, idgen.UUID{})
	if err := disabled.Issue(ctx, actor, "employee@example.test", "Employee", domain.RoleTenant, "tenant"); !errors.Is(err, ErrInvitationDeliveryUnavailable) || repo.invite.ID != "" {
		t.Fatalf("disabled delivery: %v", err)
	}
	sender := &invitationTestSender{err: errors.New("delivery failed")}
	svc := NewInvitationService(repo, invitationTestAuthorizer{}, sender, clock, idgen.UUID{})
	if err := svc.Issue(ctx, actor, "employee@example.test", "Employee", domain.RoleTenant, "tenant"); !errors.Is(err, ErrInvitationDeliveryUnavailable) || !repo.revoked {
		t.Fatalf("failed delivery: %v", err)
	}
	repo.revoked, sender.err = false, nil
	if err := svc.Issue(ctx, actor, "employee@example.test", "Employee", domain.RoleTenant, "tenant"); err != nil {
		t.Fatal(err)
	}
	clock.now = repo.invite.ExpiresAt
	if err := svc.Accept(ctx, sender.token, "employee-password"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expired token: %v", err)
	}
}

func TestRefreshRejectsTokenFromBeforePasswordRecovery(t *testing.T) {
	manager := auth.NewManager("test-secret", time.Hour)
	tokens, err := manager.Issue(domain.Principal{UserID: "u", Role: domain.RoleManager})
	if err != nil { t.Fatal(err) }
	users := delegationUsers{users: map[string]domain.User{"u": {ID: "u", Role: domain.RoleManager, IsActive: true, AuthVersion: 1}}}
	svc := NewAuthService(users, manager, domain.SystemClock{}, idgen.UUID{})
	if _, err := svc.Refresh(context.Background(), tokens.RefreshToken); !errors.Is(err, domain.ErrUnauthorized) { t.Fatalf("stale refresh token: %v", err) }
}

func TestRecoveryTargetsExistingAccount(t *testing.T) {
	clock := &invitationTestClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	repo, sender := &invitationTestRepo{}, &invitationTestSender{}
	users := delegationUsers{users: map[string]domain.User{
		"u": {ID: "u", Email: "employee@example.test", DisplayName: "Employee", Role: domain.RoleTenant, TenantID: "tenant", IsActive: true, AuthVersion: 3},
	}}
	svc := NewInvitationService(repo, invitationTestAuthorizer{}, sender, clock, idgen.UUID{}, users)
	if err := svc.IssueRecovery(context.Background(), domain.Principal{UserID: "manager", Role: domain.RoleManager}, "u"); err != nil { t.Fatal(err) }
	if repo.invite.Purpose != "recovery" || repo.invite.TargetUserID != "u" || repo.invite.TargetAuthVersion != 3 || sender.token == "" {
		t.Fatalf("unexpected recovery invitation: %+v", repo.invite)
	}
}
