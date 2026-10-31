package main

import (
	"context"
	"errors"
	"os"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
)

func seedAdmin(ctx context.Context, users domain.UserRepo, clock domain.Clock, idg domain.IDGen) error {
	email := getenv("UCRP_ADMIN_EMAIL", "admin@ucrp.local")
	password := os.Getenv("UCRP_ADMIN_PASSWORD")
	if password == "" {
		return nil
	}
	if _, err := users.GetByEmail(ctx, email); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	return users.Insert(ctx, &domain.User{
		ID:           idg.NewID(ctx),
		Email:        email,
		PasswordHash: hash,
		DisplayName:  "Admin",
		Role:         domain.RoleAnalyst,
		IsActive:     true,
		CreatedAt:    clock.Now(),
	})
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
