// Package config loads and validates application configuration from environment variables.
// Required values (DATABASE_URL, JWT_SECRET) cause a fatal error if missing.
// Non-critical values fall back to sensible defaults for local development.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the backend server.
// Fields are populated from environment variables via Load().
type Config struct {
	// HTTP server port (default: 8080).
	Port string

	// PostgreSQL connection string. Required — no default.
	DatabaseURL string

	// HMAC signing key for JWT access tokens. Required — no default.
	// Must be at least 32 characters for HS256 security.
	JWTSecret string

	// Access token time-to-live (default: 15 minutes).
	JWTAccessExpiry time.Duration

	// Refresh token time-to-live (default: 168 hours = 7 days).
	JWTRefreshExpiry time.Duration

	// Allowed CORS origin (default: http://localhost:3000).
	CORSOrigin string

	// bcrypt work factor (default: 12). Higher = slower but more secure.
	BCryptCost int
}

// Load reads environment variables and returns a validated Config.
// It first loads optional KEY=VALUE pairs from a .env file in the current
// working directory (development convenience; real env vars always win).
// It returns an error if required variables are missing or values are invalid.
func Load() (*Config, error) {
	// .env is optional: missing file is fine, a malformed-but-present file
	// only logs — never fail the whole startup over it.
	if err := loadEnvFileIntoEnv(".env"); err != nil && !errors.Is(err, errNoEnvFile) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	cfg := &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		CORSOrigin:  getEnvOrDefault("CORS_ORIGIN", "http://localhost:3000"),
	}

	// Parse JWT_ACCESS_EXPIRY (e.g. "15m", "1h").
	accessExpiry := getEnvOrDefault("JWT_ACCESS_EXPIRY", "15m")
	parsed, err := time.ParseDuration(accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY %q: %w", accessExpiry, err)
	}
	cfg.JWTAccessExpiry = parsed

	// Parse JWT_REFRESH_EXPIRY (e.g. "168h", "72h").
	refreshExpiry := getEnvOrDefault("JWT_REFRESH_EXPIRY", "168h")
	parsed, err = time.ParseDuration(refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY %q: %w", refreshExpiry, err)
	}
	cfg.JWTRefreshExpiry = parsed

	// Parse BCRYPT_COST as integer.
	bcryptCost := getEnvOrDefault("BCRYPT_COST", "12")
	cost, err := strconv.Atoi(bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("invalid BCRYPT_COST %q: must be an integer", bcryptCost)
	}
	cfg.BCryptCost = cost

	// Validate required fields — fail fast if missing.
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return cfg, nil
}

// getEnvOrDefault returns the env var value or falls back to the provided default.
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
