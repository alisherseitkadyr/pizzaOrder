package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	channel  *amqp.Channel
	exchange string
}

func NewRabbitPublisher(ch *amqp.Channel, exchange string) *RabbitPublisher {
	return &RabbitPublisher{channel: ch, exchange: exchange}
}

// публикуем OrderCreated
func (p *RabbitPublisher) PublishOrderCreated(ctx context.Context, event domain.OrderCreated) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal OrderCreated: %w", err)
	}

	return p.channel.PublishWithContext(ctx,
		p.exchange,      // exchange
		"order.created", // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// публикуем OrderStatusUpdated
func (p *RabbitPublisher) PublishStatusUpdate(ctx context.Context, event domain.OrderStatusUpdated) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal OrderStatusUpdated: %w", err)
	}

	return p.channel.PublishWithContext(ctx,
		p.exchange,     // exchange
		"order.status", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
