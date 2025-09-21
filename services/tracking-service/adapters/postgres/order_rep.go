// services/tracking-service/adapters/postgres/order_rep.go
package postgres

import (
	"context"
	"restaurant-system/services/tracking-service/domain/models"
	"restaurant-system/services/tracking-service/domain/ports"
	"restaurant-system/services/tracking-service/utils/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	db     *pgxpool.Pool
	Logger *logger.Logger
}

func NewPostgresOrderRepository(db *pgxpool.Pool, serviceName string) ports.OrderRepository {
	return &PostgresOrderRepository{
		db:     db,
		Logger: logger.New(serviceName),
	}
}

func (r *PostgresOrderRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (models.OrderStatusResponse, error) {
	query := `
		SELECT 
			number, 
			status, 
			updated_at, 
			completed_at as estimated_completion, 
			processed_by
		FROM orders 
		WHERE number = $1
	`

	var statusResponse models.OrderStatusResponse
	err := r.db.QueryRow(ctx, query, orderNumber).Scan(
		&statusResponse.OrderNumber,
		&statusResponse.CurrentStatus,
		&statusResponse.UpdatedAt,
		&statusResponse.EstimatedCompletion,
		&statusResponse.ProcessedBy,
	)
	if err != nil {
		r.Logger.Error("get_order_by_number_failed", "Failed to get order by number", orderNumber, err)
		return models.OrderStatusResponse{}, err
	}

	return statusResponse, nil
}

func (r *PostgresOrderRepository) GetOrderStatusHistory(ctx context.Context, orderNumber string) ([]models.StatusHistory, error) {
	query := `
		SELECT 
			osl.status, 
			osl.changed_at as timestamp, 
			osl.changed_by
		FROM order_status_log osl
		JOIN orders o ON osl.order_id = o.id
		WHERE o.number = $1
		ORDER BY osl.changed_at ASC
	`

	rows, err := r.db.Query(ctx, query, orderNumber)
	if err != nil {
		r.Logger.Error("get_order_status_history_failed", "Failed to get order status history", orderNumber, err)
		return nil, err
	}
	defer rows.Close()

	var history []models.StatusHistory
	for rows.Next() {
		var entry models.StatusHistory
		err := rows.Scan(
			&entry.Status,
			&entry.Timestamp,
			&entry.ChangedBy,
		)
		if err != nil {
			r.Logger.Error("scan_status_history_failed", "Failed to scan status history", orderNumber, err)
			return nil, err
		}
		history = append(history, entry)
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("rows_iteration_failed", "Error iterating over order status history rows", orderNumber, err)
		return nil, err
	}

	return history, nil
}
