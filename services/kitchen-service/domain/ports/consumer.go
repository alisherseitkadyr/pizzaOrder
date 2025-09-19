package ports

import "context"

type MessageConsumer interface {
	ConsumeOrders(ctx context.Context, handler func(message []byte) error) error
	ConsumeNotifications(ctx context.Context, handler func(message []byte) error) error
}
