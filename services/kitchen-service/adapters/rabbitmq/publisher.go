package rabbitmq

import (
	"context"
	"encoding/json"
	"restaurant-system/services/kitchen-service/domain/models"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (c *Client) Publish(exchange, routingKey string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

type RabbitMQPublisher struct {
	Client *Client
}

func (p *RabbitMQPublisher) PublishStatusUpdate(update models.StatusUpdateMessage) error {
	messageBody, err := json.Marshal(update)
	if err != nil {
		return err
	}

	return p.Client.Publish("notifications_fanout", "", messageBody)
}