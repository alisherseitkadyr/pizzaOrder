package ports

import (
	"context"
	"restaurant-system/shared/events"
)

type MessagePublisher interface {
	PublishOrderCreated(ctx context.Context, event events.OrderCreated) error
	PublishStatusUpdate(ctx context.Context, event events.OrderStatusUpdated) error
}
