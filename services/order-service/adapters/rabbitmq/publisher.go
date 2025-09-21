package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"restaurant-system/services/order-service/domain/models"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Add the Publish method to the Client struct
func (c *Client) Publish(exchange, routingKey, body string) error {
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
			Body:         []byte(body),
			DeliveryMode: amqp.Persistent, // persist on disk
		},
	)
}

type RabbitMQPublisher struct {
	Client *Client
}

func (p *RabbitMQPublisher) PublishOrder(order models.Order) error {
	orderMessage := models.OrderMessage{
		OrderNumber:     order.OrderNumber,
		CustomerName:    order.CustomerName,
		OrderType:       order.OrderType,
		TableNumber:     order.TableNumber,
		DeliveryAddress: order.DeliveryAddress,
		Items:           order.Items,
		TotalAmount:     order.TotalAmount,
		Priority:        order.Priority,
	}

	messageBody, err := json.Marshal(orderMessage)
	if err != nil {
		return err
	}

	routingKey := fmt.Sprintf("kitchen.%s.%d", order.OrderType, order.Priority)
	return p.Client.Publish("orders_topic", routingKey, string(messageBody))
}