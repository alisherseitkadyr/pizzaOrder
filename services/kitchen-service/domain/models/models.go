package models

import (
	"time"
)

type OrderMessage struct {
	OrderNumber     string       `json:"order_number"`
	CustomerName    string       `json:"customer_name"`
	OrderType       string       `json:"order_type"`
	TableNumber     *int         `json:"table_number,omitempty"`
	DeliveryAddress *string      `json:"delivery_address,omitempty"`
	Items           []OrderItem  `json:"items"`
	TotalAmount     float64      `json:"total_amount"`
	Priority        int          `json:"priority"`
}

type OrderItem struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type StatusUpdateMessage struct {
	OrderNumber        string  `json:"order_number"`
	OldStatus          string  `json:"old_status"`
	NewStatus          string  `json:"new_status"`
	ChangedBy          string  `json:"changed_by"`
	Timestamp          string  `json:"timestamp"`
	EstimatedReady     *string `json:"estimated_ready,omitempty"`
}

type Worker struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	LastSeen        time.Time `json:"last_seen"`
	OrdersProcessed int       `json:"orders_processed"`
	CreatedAt       time.Time `json:"created_at"`
}