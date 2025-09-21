package models

import "time"

type OrderStatusResponse struct {
	OrderNumber          string     `json:"order_number"`
	CurrentStatus        string     `json:"current_status"`
	UpdatedAt            time.Time  `json:"updated_at"`
	EstimatedCompletion  *time.Time `json:"estimated_completion,omitempty"`
	ProcessedBy          *string    `json:"processed_by,omitempty"`
}

type StatusHistory struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	ChangedBy string    `json:"changed_by"`
}

type WorkerStatus struct {
	WorkerName      string    `json:"worker_name"`
	Status          string    `json:"status"`
	OrdersProcessed int       `json:"orders_processed"`
	LastSeen        time.Time `json:"last_seen"`
}