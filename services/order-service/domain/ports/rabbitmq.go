package ports

import "restaurant-system/services/order-service/domain/models"

type RabbitMQPublisher interface {
	PublishOrder(order *models.OrderMessage) error
}
