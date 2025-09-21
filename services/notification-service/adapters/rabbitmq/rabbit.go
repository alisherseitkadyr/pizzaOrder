package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{conn: conn, channel: ch}, nil
}

func (c *Client) DeclareExchange(name, kind string) error {
	return c.channel.ExchangeDeclare(
		name,
		kind,
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	)
}

func (c *Client) DeclareQueue(name string) (amqp.Queue, error) {
	return c.channel.QueueDeclare(
		name,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
}

func (c *Client) BindQueue(queue, exchange, routingKey string) error {
	return c.channel.QueueBind(
		queue,
		routingKey,
		exchange,
		false, // no-wait
		nil,
	)
}

func (c *Client) Consume(queue, consumer string) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		queue,
		consumer,
		false, // auto-ack (false → manual ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
}

func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
