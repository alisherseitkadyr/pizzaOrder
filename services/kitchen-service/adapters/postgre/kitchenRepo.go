package postgre

import (
	"context"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/ports"
	"restaurant-system/services/kitchen-service/utils/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresKitchenRepo struct {
	db     *pgxpool.Pool
	Logger *logger.Logger
}

func NewPostgresKitchenRepo(db *pgxpool.Pool, serviceName string) ports.KitchenOrderRepository {
	return &PostgresKitchenRepo{
		db:     db,
		Logger: logger.New(serviceName),
	}
}

func (r *PostgresKitchenRepo) UpdateOrderStatus(ctx context.Context, orderNumber string, status domain.OrderStatus, processedBy string) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		r.Logger.Error("db_transaction", "failed to begin transaction", orderNumber, err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// если tx == nil, Rollback сразу не вызовется, но безопасно - проверим при Commit
	defer func() {
		_ = tx.Rollback(ctx) // безопасно: если уже committed, rollback вернёт ошибку, но мы её игнорируем
	}()

	// 1) Получаем id и текущий статус под блокировкой
	var orderID int
	var currentStatus string
	queryGet := `SELECT id, status FROM orders WHERE number = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, queryGet, orderNumber).Scan(&orderID, &currentStatus)
	if err != nil {
		r.Logger.Error("order_lookup", "failed to get order id", orderNumber, err)
		return fmt.Errorf("failed to get order id: %w", err)
	}

	// 2) Идемпотентность: если статус уже такой — ничего не делаем
	if currentStatus == string(status) {
		r.Logger.Info("status_idempotent", fmt.Sprintf("Order %s already in status %s", orderNumber, status), orderNumber)
		// не логируем в status_log повторно
		if err := tx.Commit(ctx); err != nil {
			r.Logger.Error("db_commit", "failed to commit (idempotent)", orderNumber, err)
			return fmt.Errorf("failed to commit (idempotent): %w", err)
		}
		return nil
	}

	// 3) Обновляем заказ
	queryUpdate := `
		UPDATE orders
		SET status = $1,
		    processed_by = $2,
		    updated_at = now(),
		    completed_at = CASE WHEN $1 = 'completed' THEN now() ELSE completed_at END
		WHERE id = $3
	`
	_, err = tx.Exec(ctx, queryUpdate, string(status), processedBy, orderID)
	if err != nil {
		r.Logger.Error("order_update", "failed to update order status", orderNumber, err)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// 4) Вставляем лог статуса
	queryLog := `
		INSERT INTO order_status_log (order_id, status, changed_by, changed_at, notes)
		VALUES ($1, $2, $3, now(), $4)
	`
	_, err = tx.Exec(ctx, queryLog, orderID, string(status), processedBy, fmt.Sprintf("status changed to %s by %s", status, processedBy))
	if err != nil {
		r.Logger.Error("status_log", "failed to insert status log", orderNumber, err)
		return fmt.Errorf("failed to insert status log: %w", err)
	}

	// 5) Commit
	if err := tx.Commit(ctx); err != nil {
		r.Logger.Error("db_commit", "failed to commit transaction", orderNumber, err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.Logger.Info("order_status_updated", fmt.Sprintf("Order %s set to %s by %s", orderNumber, status, processedBy), orderNumber)
	return nil
}
