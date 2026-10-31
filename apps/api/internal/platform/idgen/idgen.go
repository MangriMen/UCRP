package idgen

import (
	"context"

	"github.com/google/uuid"
)

type UUID struct{}

func (UUID) NewID(ctx context.Context) string {
	return uuid.NewString()
}
