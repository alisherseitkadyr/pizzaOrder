package domain

import "time"

type OrderStatus string

const (
	StatusReceived  OrderStatus = "received"
	StatusCooking   OrderStatus = "cooking"
	StatusReady     OrderStatus = "ready"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
)

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
