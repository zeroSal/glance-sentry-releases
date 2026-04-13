package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	SentryOrg            string
	SentryToken          string
	GlanceSentryPort     string
	GlanceSentryHost     string
	CacheIntervalMinutes int
}

func NewEnv() *Env {
	return &Env{}
}

func (e *Env) Load() error {
	_ = godotenv.Load()
	e.SentryOrg = os.Getenv("SENTRY_ORG")
	e.SentryToken = os.Getenv("SENTRY_AUTH_TOKEN")
	e.GlanceSentryPort = getEnv("GLANCE_SENTRY_PORT", "8099")
	e.GlanceSentryHost = getEnv("GLANCE_SENTRY_HOST", "127.0.0.1")
	e.CacheIntervalMinutes = getEnvInt("CACHE_INTERVAL_MINUTES", 5)
	return nil
}

func (e *Env) Validate() error {
	if e.SentryOrg == "" {
		return fmt.Errorf("SENTRY_ORG is required")
	}
	if e.SentryToken == "" {
		return fmt.Errorf("SENTRY_AUTH_TOKEN is required")
	}
	return nil
}

func (e *Env) GetProxyAddr() string {
	return e.GlanceSentryHost + ":" + e.GlanceSentryPort
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
