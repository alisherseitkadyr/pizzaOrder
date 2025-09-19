package events

type OrderCreated struct {
	ID           string `json:"id"`
	Number       string `json:"number"`
	CustomerName string `json:"customerName"`
	Type         string `json:"type"` // dine-in / delivery
}

type OrderStatusUpdated struct {
	ID          string `json:"id"`
	Status      string `json:"status"` // cooking / ready / completed
	ProcessedBy string `json:"processedBy,omitempty"`
}
