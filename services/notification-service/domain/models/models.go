package models

// StatusUpdateMessage represents the message format sent by Kitchen Workers
type StatusUpdateMessage struct {
	OrderNumber    string  `json:"order_number"`
	OldStatus      string  `json:"old_status"`
	NewStatus      string  `json:"new_status"`
	ChangedBy      string  `json:"changed_by"`
	Timestamp      string  `json:"timestamp"`
	EstimatedReady *string `json:"estimated_ready,omitempty"`
}

// Notification represents a formatted notification for display
type Notification struct {
	OrderNumber string `json:"order_number"`
	OldStatus   string `json:"old_status"`
	NewStatus   string `json:"new_status"`
	ChangedBy   string `json:"changed_by"`
	Message     string `json:"message"`
}
