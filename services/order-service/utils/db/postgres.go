package database

import (
	"context"
	"fmt"
	"restaurant-system/services/order-service/config"
	"restaurant-system/services/order-service/utils/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPool struct {
	*pgxpool.Pool
}

func NewPostgresPool(dbConfig config.DatabaseConfig, serviceName string) (*PostgresPool, error) {
	log := logger.New(serviceName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(dbConfig.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("db_connected", "Connected to PostgreSQL database", "")
	return &PostgresPool{pool}, nil
}

func (p *PostgresPool) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
