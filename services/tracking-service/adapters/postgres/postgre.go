package postgres

import (
	"context"
	"fmt"
	"restaurant-system/services/tracking-service/config"
	"restaurant-system/services/tracking-service/utils/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(cfg config.DatabaseConfig, serviceName string) (*pgxpool.Pool, error) {
	log := logger.New(serviceName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Connection tuning
	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("db_connected", "Connected to PostgreSQL database", "")
	return pool, nil
}
