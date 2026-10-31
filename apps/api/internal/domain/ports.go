package domain

import (
	"context"
	"time"
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type IDGen interface {
	NewID(ctx context.Context) string
}
