package ports

import (
	"context"
	domain "restaurant-system/services/kitchen-service/domain/models"
)

type StatusPublisher interface {
	// Публикация событий изменения статусов
	PublishStatusUpdate(ctx context.Context, event domain.OrderStatusUpdated) error
	PublishCookingStarted(ctx context.Context, order domain.OrderMessage, workerName string) error
	PublishOrderReady(ctx context.Context, order domain.OrderMessage, workerName string) error
}
