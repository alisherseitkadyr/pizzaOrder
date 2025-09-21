package models

import (
	"time"
)

// rabbit
type OrderMessage struct {
	OrderNumber     string
	CustomerName    string
	OrderType       string
	TableNumber     *int
	DeliveryAddress *string
	Items           []OrderItemRequest
	TotalAmount     float64
	Priority        int
}

// принимаем с апи
type OrderCreatedRequest struct {
	CustomerName    string             `json:"customer_name"`
	OrderType       string             `json:"order_type"` // dine-in / delivery
	TableNumber     *int               `json:"table_number,omitempty"`
	DeliveryAddress *string            `json:"delivery_address,omitempty"`
	Items           []OrderItemRequest `json:"items"`
}

//  принимаем с апи
type OrderItemRequest struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// ответ на апи
type CreateOrderResponse struct {
	OrderNumber string
	Status      string
	TotalAmount float64
}

// db
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
	ProcessedBy     *string
	Items           []OrderItem
	CompletedAt     *time.Time
}

// db
type OrderItem struct {
	ID        int
	OrderID   int
	Name      string
	Quantity  int
	Price     float64
	CreatedAt time.Time // Add this field
}

// db
type OrderStatusLog struct {
	ID        int
	OrderID   int
	Status    string
	ChangedBy string
	ChangedAt time.Time
	Notes     *string
	CreatedAt time.Time // Add this field
}
