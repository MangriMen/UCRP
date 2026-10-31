package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
)

type changingTenantPolicy struct {
	version int64
}

func (p *changingTenantPolicy) Authorize(context.Context, domain.Principal, string, access.Resource) (access.Decision, error) {
	return access.Decision{Allowed: true}, nil
}

func (p *changingTenantPolicy) PolicyVersion(context.Context) (int64, error) {
	return p.version, nil
}

type changingTenantGuestRepo struct {
	policy *changingTenantPolicy
}

func (r *changingTenantGuestRepo) Get(context.Context, string, string, time.Time) (*domain.TenantGuest, error) {
	r.policy.version++
	return &domain.TenantGuest{ID: "guest", FirstSeenAt: time.Now()}, nil
}

func (r *changingTenantGuestRepo) Identifiers(context.Context, string, string, time.Time, int) ([]domain.TenantGuestIdentifier, error) {
	return []domain.TenantGuestIdentifier{}, nil
}

func (r *changingTenantGuestRepo) Events(context.Context, string, string, time.Time, int, int) ([]domain.TenantGuestEvent, error) {
	return []domain.TenantGuestEvent{}, nil
}

func TestTenantGuestRejectsPolicyChangesDuringRead(t *testing.T) {
	principal := domain.Principal{Role: domain.RoleTenant, TenantID: "tenant"}
	for _, name := range []string{"profile", "events"} {
		t.Run(name, func(t *testing.T) {
			policy := &changingTenantPolicy{version: 1}
			svc := NewTenantGuestService(&changingTenantGuestRepo{policy: policy}, policy)
			var err error
			if name == "profile" {
				var result *TenantGuestProfile
				result, err = svc.Profile(context.Background(), principal, "guest")
				if result != nil {
					t.Fatal("returned a profile after the policy changed")
				}
			} else {
				var result []TenantGuestEvent
				result, err = svc.Events(context.Background(), principal, "guest", 10, 0)
				if result != nil {
					t.Fatal("returned events after the policy changed")
				}
			}
			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("expected a policy conflict, got %v", err)
			}
		})
	}
}
