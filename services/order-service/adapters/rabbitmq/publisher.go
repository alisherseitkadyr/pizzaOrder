package rabbitmq

import (
	"encoding/json"
	"fmt"
	"restaurant-system/services/order-service/domain/models"
	"restaurant-system/services/order-service/utils/logger"
)

type RabbitMQPublisher struct {
	client *Client
	logger *logger.Logger
}

func NewRabbitMQPublisher(client *Client, serviceName string) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		client: client,
		logger: logger.New(serviceName),
	}
}

func (p *RabbitMQPublisher) PublishOrder(order *models.Order) error {
	// Prepare message according to TZ format
	message := map[string]interface{}{
		"order_number":     order.OrderNumber,
		"customer_name":    order.CustomerName,
		"order_type":       order.OrderType,
		"table_number":     order.TableNumber,
		"delivery_address": order.DeliveryAddress,
		"items":            order.Items,
		"total_amount":     order.TotalAmount,
		"priority":         order.Priority,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Generate routing key according to TZ
	routingKey := fmt.Sprintf("kitchen.%s.%d", order.OrderType, order.Priority)

	// Publish with persistent delivery mode
	err = p.client.PublishWithPersistentDelivery("orders_topic", routingKey, messageBytes)
	if err != nil {
		p.logger.Error("rabbitmq_publish_failed", "Failed to publish order to RabbitMQ", order.OrderNumber, err)
		return fmt.Errorf("failed to publish order: %w", err)
	}

	p.logger.Debug("order_published", fmt.Sprintf("Order %s published to RabbitMQ with routing key %s", order.OrderNumber, routingKey), order.OrderNumber)
	return nil
}
