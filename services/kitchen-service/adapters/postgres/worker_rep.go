package postgres

import (
	"context"
	"errors"
	"restaurant-system/services/kitchen-service/domain/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresWorkerRepository struct {
	DB *pgxpool.Pool
}

func (r *PostgresWorkerRepository) RegisterWorker(name, workerType string) error {
	ctx := context.Background()

	// Check if worker already exists and is online
	var existingStatus string
	err := r.DB.QueryRow(ctx,
		"SELECT status FROM workers WHERE name = $1", name).Scan(&existingStatus)

	if err == nil && existingStatus == "online" {
		return errors.New("worker already exists and is online")
	}

	// Use upsert to register or update worker
	query := `
		INSERT INTO workers (name, type, status, last_seen, orders_processed)
		VALUES ($1, $2, 'online', NOW(), 0)
		ON CONFLICT (name) 
		DO UPDATE SET 
			status = 'online', 
			last_seen = NOW(),
			orders_processed = 0
		RETURNING id
	`

	_, err = r.DB.Exec(ctx, query, name, workerType)
	return err
}

func (r *PostgresWorkerRepository) UpdateWorkerStatus(name, status string) error {
	_, err := r.DB.Exec(context.Background(),
		"UPDATE workers SET status = $1, last_seen = NOW() WHERE name = $2",
		status, name)
	return err
}

func (r *PostgresWorkerRepository) UpdateWorkerHeartbeat(name string) error {
	_, err := r.DB.Exec(context.Background(),
		"UPDATE workers SET last_seen = NOW() WHERE name = $1", name)
	return err
}

func (r *PostgresWorkerRepository) IncrementOrdersProcessed(name string) error {
	_, err := r.DB.Exec(context.Background(),
		"UPDATE workers SET orders_processed = orders_processed + 1, last_seen = NOW() WHERE name = $1", name)
	return err
}

func (r *PostgresWorkerRepository) GetWorkerByName(name string) (models.Worker, error) {
	var worker models.Worker
	err := r.DB.QueryRow(context.Background(),
		"SELECT id, name, type, status, last_seen, orders_processed, created_at FROM workers WHERE name = $1", name).
		Scan(&worker.ID, &worker.Name, &worker.Type, &worker.Status, &worker.LastSeen, &worker.OrdersProcessed, &worker.CreatedAt)
	return worker, err
}
