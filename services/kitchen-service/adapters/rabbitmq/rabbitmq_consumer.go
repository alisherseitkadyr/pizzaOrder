package rabbitmq

import (
	"encoding/json"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/utils/logger"
)

type KitchenConsumer struct {
	client     *Client
	logger     *logger.Logger
	prefetch   int
	orderTypes []string
}

func NewKitchenConsumer(client *Client, prefetch int, orderTypes []string) (*KitchenConsumer, error) {
	consumer := &KitchenConsumer{
		client:     client,
		logger:     logger.New("kitchen-consumer"),
		prefetch:   prefetch,
		orderTypes: orderTypes,
	}

	// Declare kitchen queue
	if err := consumer.setupQueue(); err != nil {
		return nil, err
	}

	return consumer, nil
}

func (c *KitchenConsumer) setupQueue() error {
	// Declare queue
	queue, err := c.client.DeclareQueue("kitchen_orders")
	if err != nil {
		return err
	}

	// Bind to orders_topic exchange with appropriate routing keys
	routingKeys := c.generateRoutingKeys()
	for _, routingKey := range routingKeys {
		if err := c.client.BindQueue(queue.Name, "orders_topic", routingKey); err != nil {
			return err
		}
	}

	return nil
}

func (c *KitchenConsumer) generateRoutingKeys() []string {
	if len(c.orderTypes) == 0 {
		// Handle all order types
		return []string{"kitchen.*.*"}
	}

	var keys []string
	for _, orderType := range c.orderTypes {
		keys = append(keys, fmt.Sprintf("kitchen.%s.*", orderType))
	}
	return keys
}

func (c *KitchenConsumer) Consume() (<-chan domain.OrderMessage, error) {
	msgs, err := c.client.Consume("kitchen_orders", "kitchen-worker")
	if err != nil {
		return nil, err
	}

	orderChan := make(chan domain.OrderMessage)
	go func() {
		for delivery := range msgs {
			var order domain.OrderCreated
			if err := json.Unmarshal(delivery.Body, &order); err != nil {
				c.logger.Error("message_decode_failed", "Failed to decode order message", "", err)
				delivery.Nack(false, true) // Requeue
				continue
			}

			// Create message with delivery for acknowledgment
			orderMsg := domain.OrderMessage{
				OrderCreated: order,
				Delivery:     delivery,
			}

			orderChan <- orderMsg
		}
		close(orderChan)
	}()

	return orderChan, nil
}
