package rabbitmq

import (
	"context"
	"fmt"
	"restaurant-system/services/order-service/config"
	"restaurant-system/services/order-service/utils/logger"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewClient(rabbitConfig config.RabbitMQConfig, serviceName string) (*Client, error) {
	log := logger.New(serviceName)

	conn, err := amqp.Dial(rabbitConfig.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchanges
	err = ch.ExchangeDeclare(
		"orders_topic", // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	err = ch.ExchangeDeclare(
		"notifications_fanout", // name
		"fanout",               // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare fanout exchange: %w", err)
	}

	log.Info("rabbitmq_connected", "Connected to RabbitMQ", "")
	return &Client{conn: conn, ch: ch}, nil
}

func (c *Client) Publish(exchange, routingKey string, message []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.ch.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent, // Make message persistent
		})
}

func (c *Client) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
