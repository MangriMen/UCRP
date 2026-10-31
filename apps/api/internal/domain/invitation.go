package domain

import (
	"context"
	"time"
)

type UserInvitation struct {
	ID, Email, DisplayName, TenantID, CreatedBy, AuditID, Purpose, TargetUserID string
	Role                                                 Role
	TokenHash                                            []byte
	PolicyVersion                                        int64
	TargetAuthVersion                                    int64
	CreatedAt, ExpiresAt                                 time.Time
}

type InvitationRepo interface {
	CreateInvitation(context.Context, UserInvitation) error
	RevokeInvitation(context.Context, string, time.Time) error
	ValidInvitation(context.Context, []byte, time.Time) (bool, error)
	AcceptInvitation(context.Context, []byte, string, string, string, time.Time) error
}
