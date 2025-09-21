package ports

import (
	"context"
	domain "restaurant-system/services/kitchen-service/domain/models"
)

type KitchenOrderRepository interface {
	// Локальное управление заказами кухни
	UpdateOrderStatus(ctx context.Context, orderNumber string, status domain.OrderStatus, processedBy string) error
}
