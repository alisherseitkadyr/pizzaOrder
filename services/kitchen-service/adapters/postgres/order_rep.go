package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	DB *pgxpool.Pool
}

func (r *PostgresOrderRepository) UpdateOrderStatus(orderNumber, status, processedBy string) error {
	_, err := r.DB.Exec(context.Background(),
		"UPDATE orders SET status = $1, processed_by = $2, updated_at = NOW() WHERE number = $3",
		status, processedBy, orderNumber)
	return err
}

func (r *PostgresOrderRepository) SetOrderCompleted(orderNumber string) error {
	_, err := r.DB.Exec(context.Background(),
		"UPDATE orders SET status = 'ready', completed_at = NOW(), updated_at = NOW() WHERE number = $1",
		orderNumber)
	return err
}

func (r *PostgresOrderRepository) GetOrderStatus(orderNumber string) (string, error) {
	var status string
	err := r.DB.QueryRow(context.Background(),
		"SELECT status FROM orders WHERE number = $1", orderNumber).Scan(&status)
	return status, err
}
