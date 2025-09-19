package domain

import amqp "github.com/rabbitmq/amqp091-go"

// OrderMessage для обработки заказов из RabbitMQ
type OrderMessage struct {
	OrderCreated
	Delivery amqp.Delivery // Для manual acknowledgment
}

// NotificationMessage для уведомлений
type NotificationMessage struct {
	OrderStatusUpdated
	Delivery amqp.Delivery // Для manual acknowledgment
}
