package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
)

var ErrInvitationDeliveryUnavailable = errors.New("invitation delivery is unavailable")

type InvitationSender interface {
	SendInvitation(context.Context, string, string, time.Time) error
}

type InvitationAuthorizer interface {
	CanInvite(context.Context, domain.Principal, domain.Role, string) (int64, error)
}

type InvitationService struct {
	repo       domain.InvitationRepo
	authorizer InvitationAuthorizer
	sender     InvitationSender
	clock      domain.Clock
	ids        domain.IDGen
	users      domain.UserRepo
}

func NewInvitationService(repo domain.InvitationRepo, authorizer InvitationAuthorizer, sender InvitationSender, clock domain.Clock, ids domain.IDGen, users ...domain.UserRepo) *InvitationService {
	svc := &InvitationService{repo: repo, authorizer: authorizer, sender: sender, clock: clock, ids: ids}
	if len(users) > 0 { svc.users = users[0] }
	return svc
}

func (s *InvitationService) Issue(ctx context.Context, actor domain.Principal, email, displayName string, role domain.Role, tenantID string) error {
	if s.sender == nil {
		return ErrInvitationDeliveryUnavailable
	}
	email, displayName = strings.TrimSpace(email), strings.TrimSpace(displayName)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || displayName == "" || len(displayName) > 200 || len(email) > 320 {
		return domain.ErrValidation
	}
	if role != domain.RoleManager && role != domain.RoleTenant || role == domain.RoleManager && tenantID != "" || role == domain.RoleTenant && tenantID == "" {
		return domain.ErrValidation
	}
	return s.send(ctx, actor, domain.UserInvitation{Email: email, DisplayName: displayName, Role: role, TenantID: tenantID, Purpose: "invite"})
}

func (s *InvitationService) IssueRecovery(ctx context.Context, actor domain.Principal, userID string) error {
	if s.sender == nil { return ErrInvitationDeliveryUnavailable }
	if s.users == nil { return domain.ErrForbidden }
	user, err := s.users.GetByID(ctx, userID)
	if err != nil { return err }
	if !user.IsActive { return domain.ErrForbidden }
	return s.send(ctx, actor, domain.UserInvitation{Email: user.Email, DisplayName: user.DisplayName, Role: user.Role,
		TenantID: user.TenantID, Purpose: "recovery", TargetUserID: user.ID, TargetAuthVersion: user.AuthVersion})
}

func (s *InvitationService) send(ctx context.Context, actor domain.Principal, invite domain.UserInvitation) error {
	version, err := s.authorizer.CanInvite(ctx, actor, invite.Role, invite.TenantID)
	if err != nil {
		return err
	}
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes[:])
	hash := sha256.Sum256(bytes[:])
	now := s.clock.Now().UTC()
	invite.ID, invite.AuditID, invite.CreatedBy = s.ids.NewID(ctx), s.ids.NewID(ctx), actor.UserID
	invite.TokenHash, invite.PolicyVersion, invite.CreatedAt, invite.ExpiresAt = hash[:], version, now, now.Add(24*time.Hour)
	if err := s.repo.CreateInvitation(ctx, invite); err != nil {
		return err
	}
	if err := s.sender.SendInvitation(ctx, invite.Email, token, invite.ExpiresAt); err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		if revokeErr := s.repo.RevokeInvitation(cleanup, invite.ID, s.clock.Now().UTC()); revokeErr != nil {
			return revokeErr
		}
		return ErrInvitationDeliveryUnavailable
	}
	return nil
}

func (s *InvitationService) Accept(ctx context.Context, token, password string) error {
	if s.sender == nil {
		return ErrInvitationDeliveryUnavailable
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 {
		return domain.ErrUnauthorized
	}
	if len(password) < 8 || len(password) > 72 {
		return domain.ErrValidation
	}
	hash := sha256.Sum256(raw)
	valid, err := s.repo.ValidInvitation(ctx, hash[:], s.clock.Now().UTC())
	if err != nil {
		return err
	}
	if !valid {
		return domain.ErrUnauthorized
	}
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	return s.repo.AcceptInvitation(ctx, hash[:], s.ids.NewID(ctx), passwordHash, s.ids.NewID(ctx), s.clock.Now().UTC())
}
