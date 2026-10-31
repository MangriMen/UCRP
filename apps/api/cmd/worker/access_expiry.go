package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/MangriMen/UCRP/internal/service"
)

func runGrantExpiry(ctx context.Context, expiry *service.GrantExpiry, log *slog.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		count, err := expiry.RunOnce(ctx)
		if err != nil && ctx.Err() == nil {
			log.Error("expire access grants", "err", err)
		}
		if err == nil && count == 100 {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
