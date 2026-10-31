package service

import (
	"context"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type grantExpiryRepository interface {
	ExpireDue(ctx context.Context, now time.Time, limit int) (int, error)
}

type GrantExpiry struct {
	repo  grantExpiryRepository
	clock domain.Clock
}

func NewGrantExpiry(repo grantExpiryRepository, clock domain.Clock) *GrantExpiry {
	return &GrantExpiry{repo: repo, clock: clock}
}

func (s *GrantExpiry) RunOnce(ctx context.Context) (int, error) {
	return s.repo.ExpireDue(ctx, s.clock.Now().UTC(), 100)
}
