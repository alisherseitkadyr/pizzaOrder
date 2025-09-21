package ports

import (
	"context"
	domain "restaurant-system/services/kitchen-service/domain/models"
)

type MessageConsumer interface {
	ConsumeOrders(ctx context.Context) (<-chan domain.OrderMessage, error)
	AckMessage(message domain.OrderMessage) error
	NackMessage(message domain.OrderMessage, requeue bool) error
}
