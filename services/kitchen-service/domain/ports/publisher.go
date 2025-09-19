package ports

import (
	"context"
	domain "restaurant-system/services/kitchen-service/domain/models"
)

type MessagePublisher interface {
	PublishOrderCreated(ctx context.Context, event domain.OrderCreated) error
	PublishStatusUpdate(ctx context.Context, event domain.OrderStatusUpdated) error
}
