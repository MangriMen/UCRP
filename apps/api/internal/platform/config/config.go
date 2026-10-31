package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	JWTSecret    string
	UploadDir    string
	LogLevel     string
	JWTAccessTTL time.Duration
	HTTPTimeout  time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:     getenv("UCRP_HTTP_ADDR", ":8080"),
		DatabaseURL:  os.Getenv("UCRP_DATABASE_URL"),
		JWTSecret:    os.Getenv("UCRP_JWT_SECRET"),
		UploadDir:    getenv("UCRP_UPLOAD_DIR", "./uploads"),
		LogLevel:     getenv("UCRP_LOG_LEVEL", "info"),
		JWTAccessTTL: 15 * time.Minute,
		HTTPTimeout:  30 * time.Second,
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("UCRP_DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("UCRP_JWT_SECRET is required")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
