// Package database provides PostgreSQL connection pool setup and health checks.
// It uses pgxpool (jackc/pgx/v5) for efficient connection pooling with native
// PostgreSQL type support.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds settings for the PostgreSQL connection pool.
// These are separate from the main app config to keep database concerns isolated.
type PoolConfig struct {
	// DatabaseURL is the PostgreSQL connection string.
	DatabaseURL string

	// MaxConns limits the pool size (default: 10).
	MaxConns int32

	// MinConns keeps this many connections warm in the pool (default: 2).
	MinConns int32

	// MaxConnLifetime closes connections older than this (default: 30 min).
	MaxConnLifetime time.Duration

	// MaxConnIdleTime closes idle connections after this (default: 5 min).
	MaxConnIdleTime time.Duration

	// HealthCheckInterval runs pool health checks at this interval (default: 30s).
	HealthCheckInterval time.Duration
}

// DefaultPoolConfig returns sensible defaults for local development.
// Production deployments may want higher MaxConns and longer lifetimes.
func DefaultPoolConfig(databaseURL string) PoolConfig {
	return PoolConfig{
		DatabaseURL:         databaseURL,
		MaxConns:            10,
		MinConns:            2,
		MaxConnLifetime:     30 * time.Minute,
		MaxConnIdleTime:     5 * time.Minute,
		HealthCheckInterval: 30 * time.Second,
	}
}

// NewPostgresDB creates a new pgxpool connection pool with the given configuration.
// The pool is ready to use immediately — callers must call Pool.Close() on shutdown.
func NewPostgresDB(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckInterval

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Verify the connection is alive before returning.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// HealthCheck verifies the database connection is alive by pinging it.
// Returns nil if healthy, or an error describing the problem.
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}
