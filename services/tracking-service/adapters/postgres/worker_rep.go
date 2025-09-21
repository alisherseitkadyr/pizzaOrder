// services/tracking-service/adapters/postgres/worker_rep.go
package postgres

import (
	"context"
	"restaurant-system/services/tracking-service/domain/models"
	"restaurant-system/services/tracking-service/domain/ports"
	"restaurant-system/services/tracking-service/utils/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresWorkerRepository struct {
	db     *pgxpool.Pool
	Logger *logger.Logger
}

func NewPostgresWorkerRepository(db *pgxpool.Pool, serviceName string) ports.WorkerRepository {
	return &PostgresWorkerRepository{
		db:     db,
		Logger: logger.New(serviceName),
	}
}

func (r *PostgresWorkerRepository) GetAllWorkersStatus(ctx context.Context) ([]models.WorkerStatus, error) {
	query := `
		SELECT 
			name as worker_name, 
			orders_processed, 
			last_seen
		FROM workers
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.Logger.Error("get_all_workers_failed", "Failed to get workers status", "", err)
		return nil, err
	}
	defer rows.Close()

	var workers []models.WorkerStatus
	for rows.Next() {
		var worker models.WorkerStatus
		err := rows.Scan(
			&worker.WorkerName,
			&worker.OrdersProcessed,
			&worker.LastSeen,
		)
		if err != nil {
			r.Logger.Error("scan_worker_failed", "Failed to scan worker", "", err)
			return nil, err
		}
		workers = append(workers, worker)
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("rows_iteration_failed", "Error iterating over workers", "", err)
		return nil, err
	}

	return workers, nil
}
