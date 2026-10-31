package config

import (
	"fmt"
	"os"
	"strconv"
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
	Recommend    RecommendConfig
	Kev          KevConfig
}

type RecommendConfig struct {
	AlgorithmVersion     string
	WindowDays           int
	Lambda               float64
	RecencyHalfLifeDays  float64
	WeightBehavior       float64
	WeightCategory       float64
	WeightPopularity     float64
	WeightPriority       float64
	CompetitorPenalty    float64
	SameCategoryAffinity float64
	StaleAfter           time.Duration
}

type KevConfig struct {
	BaseURL string
	Timeout time.Duration
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
		Recommend:    DefaultRecommendConfig(),
		Kev: KevConfig{
			BaseURL: os.Getenv("UCRP_KEV_BASE_URL"),
			Timeout: time.Duration(getenvInt("UCRP_KEV_TIMEOUT_SEC", 60)) * time.Second,
		},
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("UCRP_DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("UCRP_JWT_SECRET is required")
	}
	if cfg.Kev.Timeout <= 0 {
		return Config{}, fmt.Errorf("UCRP_KEV_TIMEOUT_SEC must be > 0")
	}
	if err := cfg.Recommend.loadEnv(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func DefaultRecommendConfig() RecommendConfig {
	return RecommendConfig{
		AlgorithmVersion:     getenv("UCRP_RECOMMEND_ALGO_VERSION", "v2"),
		WindowDays:           getenvInt("UCRP_RECOMMEND_WINDOW_DAYS", 365),
		Lambda:               getenvFloat("UCRP_RECOMMEND_LAMBDA", 5),
		RecencyHalfLifeDays:  getenvFloat("UCRP_RECOMMEND_RECENCY_HALF_LIFE_DAYS", 180),
		WeightBehavior:       getenvFloat("UCRP_RECOMMEND_W_BEHAVIOR", 0.45),
		WeightCategory:       getenvFloat("UCRP_RECOMMEND_W_CATEGORY", 0.15),
		WeightPopularity:     getenvFloat("UCRP_RECOMMEND_W_POPULARITY", 0.25),
		WeightPriority:       getenvFloat("UCRP_RECOMMEND_W_PRIORITY", 0.15),
		CompetitorPenalty:    getenvFloat("UCRP_RECOMMEND_COMPETITOR_PENALTY", 0.3),
		SameCategoryAffinity: getenvFloat("UCRP_RECOMMEND_SAME_CATEGORY_AFFINITY", 1),
		StaleAfter:           time.Duration(getenvInt("UCRP_RECOMMEND_STALE_HOURS", 24)) * time.Hour,
	}
}

func (c RecommendConfig) loadEnv() error {
	if c.WindowDays <= 0 {
		return fmt.Errorf("recommend window days must be > 0")
	}
	if c.Lambda < 0 {
		return fmt.Errorf("recommend lambda must be >= 0")
	}
	if c.RecencyHalfLifeDays <= 0 {
		return fmt.Errorf("recommend recency half-life days must be > 0")
	}
	sum := c.WeightBehavior + c.WeightCategory + c.WeightPopularity + c.WeightPriority
	if sum <= 0 {
		return fmt.Errorf("recommend weights must sum to a positive value")
	}
	if c.CompetitorPenalty < 0 || c.CompetitorPenalty > 1 {
		return fmt.Errorf("recommend competitor penalty must be in [0, 1]")
	}
	if c.SameCategoryAffinity < 0 || c.SameCategoryAffinity > 1 {
		return fmt.Errorf("recommend same-category affinity must be in [0, 1]")
	}
	return nil
}

func (c RecommendConfig) WeightSum() float64 {
	return c.WeightBehavior + c.WeightCategory + c.WeightPopularity + c.WeightPriority
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getenvFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
