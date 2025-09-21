package domain

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderStatus string

const (
	StatusReceived  OrderStatus = "received"
	StatusCooking   OrderStatus = "cooking"
	StatusReady     OrderStatus = "ready"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
)

type OrderMessage struct {
	OrderNumber     string
	CustomerName    string
	OrderType       string
	TableNumber     *int
	DeliveryAddress *string
	Items           []OrderItemRequest
	TotalAmount     float64
	Priority        int
	Delivery        amqp.Delivery
}

type OrderItemRequest struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type OrderStatusUpdated struct {
	OrderNumber         string    `json:"order_number"`
	OldStatus           string    `json:"old_status"`
	NewStatus           string    `json:"new_status"`
	ChangedBy           string    `json:"changed_by"`
	Timestamp           time.Time `json:"timestamp"`
	EstimatedCompletion time.Time `json:"estimated_completion,omitempty"`
}

type OrderStatusLog struct {
	ID        int64
	OrderID   int
	Status    OrderStatus
	ChangedBy string
	ChangedAt time.Time
	Notes     *string
	CreatedAt time.Time
}

func (o *OrderMessage) CookingTime() time.Duration {
	switch o.OrderType {
	case "dine_in":
		return 8 * time.Second
	case "takeout":
		return 10 * time.Second
	case "delivery":
		return 12 * time.Second
	default:
		return 10 * time.Second
	}
}
