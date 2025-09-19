package ports

import "restaurant-system/services/order-service/domain/models"

type RabbitMQPublisher interface {
	PublishOrder(order *models.Order) error
}

type RabbitMQConsumer interface {
	ConsumeOrders() (<-chan models.Order, error)
	AcknowledgeMessage(messageID string) error
	RejectMessage(messageID string, requeue bool) error
}
