package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderStatusLogRepository struct {
	DB *pgxpool.Pool
}

func (r *PostgresOrderStatusLogRepository) LogStatusChange(orderNumber, status, changedBy string) error {
	// First get the order ID
	var orderID int
	err := r.DB.QueryRow(context.Background(),
		"SELECT id FROM orders WHERE number = $1", orderNumber).Scan(&orderID)
	if err != nil {
		return err
	}

	// Insert status log
	_, err = r.DB.Exec(context.Background(),
		"INSERT INTO order_status_log (order_id, status, changed_by, changed_at) VALUES ($1, $2, $3, NOW())",
		orderID, status, changedBy)
	return err
}