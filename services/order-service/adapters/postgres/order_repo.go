package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"restaurant-system/services/order-service/domain/models"
	"restaurant-system/services/order-service/utils/logger"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	DB     *pgxpool.Pool
	Logger *logger.Logger
}

func NewPostgresOrderRepository(db *pgxpool.Pool, serviceName string) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		DB:     db,
		Logger: logger.New(serviceName),
	}
}

func (r *PostgresOrderRepository) SaveOrderWithItems(ctx context.Context, order *models.Order, items []models.OrderItem) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Save order
	orderQuery := `
		INSERT INTO orders (number, customer_name, type, table_number, delivery_address, total_amount, priority, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	var tableNumber *int
	if order.TableNumber != nil {
		tableNumber = order.TableNumber
	}

	err = tx.QueryRow(ctx, orderQuery,
		order.OrderNumber,
		order.CustomerName,
		order.OrderType,
		tableNumber,
		order.DeliveryAddress,
		order.TotalAmount,
		order.Priority,
		order.Status,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Save order items
	itemQuery := `
		INSERT INTO order_items (order_id, name, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	for i := range items {
		items[i].OrderID = order.ID
		err := tx.QueryRow(ctx, itemQuery,
			items[i].OrderID,
			items[i].Name,
			items[i].Quantity,
			items[i].Price,
		).Scan(&items[i].ID, &items[i].CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to save order item: %w", err)
		}
	}

	// Save status log
	statusLogQuery := `
		INSERT INTO order_status_log (order_id, status, changed_by, changed_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	statusLog := models.OrderStatusLog{
		OrderID:   order.ID,
		Status:    order.Status,
		ChangedBy: "order-service",
		ChangedAt: time.Now(),
	}

	err = tx.QueryRow(ctx, statusLogQuery,
		statusLog.OrderID,
		statusLog.Status,
		statusLog.ChangedBy,
		statusLog.ChangedAt,
	).Scan(&statusLog.ID, &statusLog.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save status log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.Logger.Info("order_saved", fmt.Sprintf("Order %s saved successfully", order.OrderNumber), "")
	return nil
}

func (r *PostgresOrderRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	query := `
		SELECT id, created_at, updated_at, number, customer_name, type, 
		       table_number, delivery_address, total_amount, priority, status,
		       processed_by, completed_at
		FROM orders 
		WHERE number = $1
	`

	var order models.Order
	var tableNumber sql.NullInt32
	var deliveryAddress, processedBy sql.NullString
	var completedAt sql.NullTime

	err := r.DB.QueryRow(ctx, query, orderNumber).Scan(
		&order.ID,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.OrderNumber,
		&order.CustomerName,
		&order.OrderType,
		&tableNumber,
		&deliveryAddress,
		&order.TotalAmount,
		&order.Priority,
		&order.Status,
		&processedBy,
		&completedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if tableNumber.Valid {
		val := int(tableNumber.Int32)
		order.TableNumber = &val
	}
	if deliveryAddress.Valid {
		order.DeliveryAddress = &deliveryAddress.String
	}

	return &order, nil
}

func (r *PostgresOrderRepository) GetOrderItems(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	query := `
		SELECT id, created_at, order_id, name, quantity, price
		FROM order_items 
		WHERE order_id = $1
		ORDER BY id
	`

	rows, err := r.DB.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.CreatedAt,
			&item.OrderID,
			&item.Name,
			&item.Quantity,
			&item.Price,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return items, nil
}

func (r *PostgresOrderRepository) UpdateOrderStatus(ctx context.Context, orderID int, status string, processedBy string) error {
	query := `
		UPDATE orders 
		SET status = $1, processed_by = $2, updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.DB.Exec(ctx, query, status, processedBy, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}
