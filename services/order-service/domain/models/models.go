package models

import (
	"time"
)

type OrderCreated struct {
	ID              string         `json:"id"`
	Number          string         `json:"number"`
	CustomerName    string         `json:"customer_name"`
	OrderType       string         `json:"order_type"` // dine-in / delivery
	TableNumber     *int           `json:"table_number,omitempty"`
	DeliveryAddress *string        `json:"delivery_address,omitempty"`
	Items           []OrderItemMes `json:"items"`
	TotalAmount     float64        `json:"total_amount"`
	Priority        int            `json:"priority"`
}

type OrderItemMes struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type OrderStatusUpdated struct {
	ID          string `json:"id"`
	Status      string `json:"status"` // cooking / ready / completed
	ProcessedBy string `json:"processed_by,omitempty"`
}

type Order struct {
	ID              int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	OrderNumber     string
	CustomerName    string
	OrderType       string
	TableNumber     *int
	DeliveryAddress *string
	TotalAmount     float64
	Priority        int
	Status          string
	Items           []OrderItem
}

type OrderItem struct {
	ID        int
	OrderID   int
	Name      string
	Quantity  int
	Price     float64
	CreatedAt time.Time // Add this field
}

type OrderStatusLog struct {
	ID        int
	OrderID   int
	Status    string
	ChangedBy string
	ChangedAt time.Time
	Notes     string
	CreatedAt time.Time // Add this field
}
