package ports

import (
	"context"
	"restaurant-system/services/order-service/domain/models"
)

type OrderRepository interface {
	SaveOrderWithItems(ctx context.Context, order *models.Order, items []models.OrderItem) error
	GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error)
	GetOrderItems(ctx context.Context, orderID int) ([]models.OrderItem, error)
	UpdateOrderStatus(ctx context.Context, orderID int, status string, processedBy string) error
}
