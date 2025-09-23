package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
)

// Store wraps the pgx connection pool and exposes repositories.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a pgx pool using the provided database configuration.
func New(ctx context.Context, cfg config.DatabaseConfig) (*Store, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		poolCfg.MinConns = int32(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Close releases database resources.
func (s *Store) Close() {
	s.pool.Close()
}

// Pool exposes the underlying pgx pool for lower-level use when needed.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// Ping checks database connectivity within the provided timeout.
func (s *Store) Ping(ctx context.Context) error {
	deadlineCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.pool.Ping(deadlineCtx)
}
