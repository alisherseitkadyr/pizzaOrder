package domain

import (
	"time"
)

type OrderStatus string

const (
	StatusReceived  OrderStatus = "received"
	StatusCooking   OrderStatus = "cooking"
	StatusReady     OrderStatus = "ready"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
)

type OrderCreated struct {
	OrderNumber     string         `json:"order_number"`
	CustomerName    string         `json:"customer_name"`
	OrderType       string         `json:"order_type"`
	TableNumber     *int           `json:"table_number,omitempty"`
	DeliveryAddress *string        `json:"delivery_address,omitempty"`
	Items           []OrderItemMes `json:"items"`
	TotalAmount     float64        `json:"total_amount"`
	Priority        int            `json:"priority"`
	Status          OrderStatus    `json:"status,omitempty"`
	CreatedAt       time.Time      `json:"created_at,omitempty"`
}

type OrderItemMes struct {
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
	OrderID   int64
	Status    OrderStatus
	ChangedBy string
	ChangedAt time.Time
	Notes     *string
}

func NewStatusLog(orderID int64, status OrderStatus, changedBy string, notes *string) OrderStatusLog {
	return OrderStatusLog{
		OrderID:   orderID,
		Status:    status,
		ChangedBy: changedBy,
		ChangedAt: time.Now(),
		Notes:     notes,
	}
}
