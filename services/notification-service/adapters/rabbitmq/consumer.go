package rabbitmq

import (
	"encoding/json"
	"log"
	"restaurant-system/services/notification-service/domain/models"
)

type NotificationConsumer struct {
	client *Client
}

func NewNotificationConsumer(client *Client) *NotificationConsumer {
	return &NotificationConsumer{client: client}
}

func (c *NotificationConsumer) Setup() error {
	// Declare fanout exchange
	err := c.client.DeclareExchange("notifications_fanout", "fanout")
	if err != nil {
		return err
	}

	// Declare queue (let RabbitMQ generate a unique name for each subscriber)
	queue, err := c.client.DeclareQueue("") // empty name = auto-generate
	if err != nil {
		return err
	}

	// Bind queue to exchange
	err = c.client.BindQueue(queue.Name, "notifications_fanout", "")
	if err != nil {
		return err
	}

	log.Printf("Notification queue '%s' bound to notifications_fanout exchange", queue.Name)

	return nil
}

func (c *NotificationConsumer) StartConsuming(handler func(models.StatusUpdateMessage)) error {
	// Get queue name by declaring again (since we used auto-generate)
	queue, err := c.client.DeclareQueue("")
	if err != nil {
		return err
	}

	// Bind to exchange
	err = c.client.BindQueue(queue.Name, "notifications_fanout", "")
	if err != nil {
		return err
	}

	// Start consuming
	msgs, err := c.client.Consume(queue.Name, "notification-consumer")
	if err != nil {
		return err
	}

	log.Printf("Started consuming from queue: %s", queue.Name)

	go func() {
		for msg := range msgs {
			var statusUpdate models.StatusUpdateMessage

			// Parse JSON message
			if err := json.Unmarshal(msg.Body, &statusUpdate); err != nil {
				log.Printf("Error parsing message: %v", err)
				msg.Nack(false, false) // reject and don't requeue
				continue
			}

			// Call handler
			handler(statusUpdate)

			// Acknowledge message
			msg.Ack(false)
		}
	}()

	return nil
}
