package postgre

import (
	"context"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/ports"
	"restaurant-system/services/kitchen-service/utils/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresWorkerRepo struct {
	db     *pgxpool.Pool
	Logger *logger.Logger
}

func NewPostgresWorkerRepo(db *pgxpool.Pool, serviceName string) ports.WorkerRepository {
	return &PostgresWorkerRepo{
		db:     db,
		Logger: logger.New(serviceName),
	}
}

func (r *PostgresWorkerRepo) Register(ctx context.Context, worker *domain.Worker) error {
	query := `
		INSERT INTO workers(name, type, status, orders_processed, last_seen, created_at)
		VALUES($1,$2,$3,$4,$5,now())
		RETURNING id
	`
	err := r.db.QueryRow(
		ctx,
		query,
		worker.Name,
		worker.Type,
		worker.Status,
		worker.OrdersProcessed,
		worker.LastSeen,
	).Scan(&worker.ID)
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}
	return nil
}

func (r *PostgresWorkerRepo) Update(ctx context.Context, worker *domain.Worker) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE workers
		SET status = $1, orders_processed = $2, last_seen = $3
		WHERE id = $4`,
		worker.Status,
		worker.OrdersProcessed,
		worker.LastSeen,
		worker.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update worker: %w", err)
	}
	return nil
}

func (r *PostgresWorkerRepo) GetAll(ctx context.Context) ([]domain.Worker, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, type, status, orders_processed, last_seen, created_at
		 FROM workers ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query workers: %w", err)
	}
	defer rows.Close()

	var workers []domain.Worker
	for rows.Next() {
		var worker domain.Worker
		err := rows.Scan(
			&worker.ID,
			&worker.Name,
			&worker.Type,
			&worker.Status,
			&worker.OrdersProcessed,
			&worker.LastSeen,
			&worker.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan worker: %w", err)
		}
		workers = append(workers, worker)
	}
	return workers, nil
}

func (r *PostgresWorkerRepo) GetByName(ctx context.Context, name string) (*domain.Worker, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, type, status, orders_processed, last_seen, created_at
		 FROM workers WHERE name = $1`,
		name,
	)
	var worker domain.Worker
	err := row.Scan(
		&worker.ID,
		&worker.Name,
		&worker.Type,
		&worker.Status,
		&worker.OrdersProcessed,
		&worker.LastSeen,
		&worker.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get worker by name: %w", err)
	}
	return &worker, nil
}
