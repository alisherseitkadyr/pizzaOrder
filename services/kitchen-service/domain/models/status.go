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
