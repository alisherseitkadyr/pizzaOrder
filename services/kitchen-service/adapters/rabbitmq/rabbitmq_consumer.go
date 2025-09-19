package rabbitmq

import (
	"context"
	"fmt"
	"restaurant-system/services/kitchen-service/domain/ports"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func NewRabbitMQConsumer(conn *amqp091.Connection) (ports.MessageConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	return &RabbitMQConsumer{conn: conn, channel: ch}, nil
}

func (c *RabbitMQConsumer) ConsumeOrders(ctx context.Context, handler func(message []byte) error) error {
	msgs, err := c.channel.Consume(
		"orders",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume orders: %w", err)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				_ = handler(msg.Body)
			}
		}
	}()
	return nil
}

func (c *RabbitMQConsumer) ConsumeNotifications(ctx context.Context, handler func(message []byte) error) error {
	q, err := c.channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare temp queue: %w", err)
	}
	err = c.channel.QueueBind(q.Name, "", "status_updates", false, nil)
	if err != nil {
		return fmt.Errorf("failed to bind to status_updates exchange: %w", err)
	}
	msgs, err := c.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume notifications:%w", err)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				_ = handler(msg.Body)
			}
		}
	}()
	return nil
}
