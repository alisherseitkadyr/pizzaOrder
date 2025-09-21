package ports

import (
	"context"
	"restaurant-system/services/tracking-service/domain/models"
)

type OrderRepository interface {
	GetOrderByNumber(ctx context.Context, orderNumber string) (models.OrderStatusResponse, error)
	GetOrderStatusHistory(ctx context.Context, orderNumber string) ([]models.StatusHistory, error)
}

type WorkerRepository interface {
	GetAllWorkersStatus(ctx context.Context) ([]models.WorkerStatus, error)
}
