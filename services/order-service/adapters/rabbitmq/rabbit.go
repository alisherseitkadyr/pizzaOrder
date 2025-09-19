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
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewClient(rabbitConfig config.RabbitMQConfig, serviceName string) (*Client, error) {
	log := logger.New(serviceName)
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", rabbitConfig.User, rabbitConfig.Password, rabbitConfig.Host, rabbitConfig.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set quality of service
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}
	log.Info("MessageBrocker", "Connected to RabbitMq database", "")

	return &Client{conn: conn, channel: ch}, nil
}

func (c *Client) DeclareExchange(name, exchangeType string) error {
	err := c.channel.ExchangeDeclare(
		name,         // name
		exchangeType, // type: "topic", "fanout", etc.
		true,         // durable
		false,        // auto-delete
		false,        // internal
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", name, err)
	}
	return nil
}

func (c *Client) DeclareQueue(queueName string) (amqp.Queue, error) {
	queue, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}
	return queue, nil
}

func (c *Client) BindQueue(queueName, exchange, routingKey string) error {
	err := c.channel.QueueBind(
		queueName,
		routingKey,
		exchange,
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s with key %s: %w", queueName, exchange, routingKey, err)
	}
	return nil
}

func (c *Client) Publish(exchange, routingKey string, message []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.channel.PublishWithContext(ctx,
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

func (c *Client) PublishWithPersistentDelivery(exchange, routingKey string, message []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.channel.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent, // Persistent delivery mode
			Priority:     0,               // You can set priority here if needed
		})
}

func (c *Client) Consume(queueName, consumer string) (<-chan amqp.Delivery, error) {
	msgs, err := c.channel.Consume(
		queueName,
		consumer, // consumer
		false,    // auto-ack (set to false for manual ack)
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume from queue %s: %w", queueName, err)
	}
	return msgs, nil
}

func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
