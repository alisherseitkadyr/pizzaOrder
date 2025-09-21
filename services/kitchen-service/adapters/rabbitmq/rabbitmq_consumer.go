package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/utils/logger"
)

type KitchenConsumer struct {
	client    *Client
	logger    *logger.Logger
	prefetch  int
	orderType string
}

func NewKitchenConsumer(client *Client, prefetch int, orderType string) (*KitchenConsumer, error) {
	consumer := &KitchenConsumer{
		client:    client,
		logger:    logger.New("kitchen-consumer"),
		prefetch:  prefetch,
		orderType: orderType,
	}

	if err := consumer.setupQueue(); err != nil {
		return nil, err
	}

	return consumer, nil
}

func (c *KitchenConsumer) setupQueue() error {
	queue, err := c.client.DeclareQueue("kitchen_orders")
	if err != nil {
		return err
	}

	// строго один тип
	routingKey := fmt.Sprintf("kitchen.%s.*", c.orderType)
	if err := c.client.BindQueue(queue.Name, "orders_topic", routingKey); err != nil {
		return err
	}

	return nil
}

func (c *KitchenConsumer) ConsumeOrders(ctx context.Context) (<-chan domain.OrderMessage, error) {
	msgs, err := c.client.Consume("kitchen_orders", "kitchen-worker")
	if err != nil {
		return nil, err
	}

	orderChan := make(chan domain.OrderMessage)
	go func() {
		defer close(orderChan)
		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-msgs:
				if !ok {
					return
				}

				var order domain.OrderMessage
				if err := json.Unmarshal(delivery.Body, &order); err != nil {
					c.logger.Error("message_decode_failed", "Failed to decode order message", "", err)
					continue
				}
				order.Delivery = delivery
				orderChan <- order
			}
		}
	}()

	return orderChan, nil
}

func (c *KitchenConsumer) AckMessage(msg domain.OrderMessage) error {
	return msg.Delivery.Ack(false)
}

func (c *KitchenConsumer) NackMessage(msg domain.OrderMessage, requeue bool) error {
	return msg.Delivery.Nack(false, requeue)
}
