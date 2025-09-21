package rabbitmq

import (
	"encoding/json"
	"log"
	"restaurant-system/services/kitchen-service/domain/models"
)

type KitchenConsumer struct {
	client *Client
}

func NewKitchenConsumer(client *Client) *KitchenConsumer {
	return &KitchenConsumer{client: client}
}

func (c *KitchenConsumer) ConsumeOrders(queueName, workerName string, handler func(models.OrderMessage)) error {
	// Ensure queue exists
	queue, err := c.client.DeclareQueue(queueName)
	if err != nil {
		return err
	}

	// Bind to orders_topic exchange with appropriate routing key
	routingKey := "kitchen.*.*" // Listen to all kitchen messages
	err = c.client.BindQueue(queue.Name, "orders_topic", routingKey)
	if err != nil {
		return err
	}

	// Start consuming
	msgs, err := c.client.Consume(queue.Name, workerName)
	if err != nil {
		return err
	}

	log.Printf("Started consuming from queue: %s", queue.Name)

	go func() {
		for msg := range msgs {
			var order models.OrderMessage

			if err := json.Unmarshal(msg.Body, &order); err != nil {
				log.Printf("Error parsing order message: %v", err)
				msg.Nack(false, true) // requeue
				continue
			}

			handler(order)
			// Message acknowledgement is handled by the service layer
		}
	}()

	return nil
}
