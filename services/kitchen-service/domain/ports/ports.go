package ports

import (
	"restaurant-system/services/kitchen-service/domain/models"
)

type WorkerRepository interface {
	RegisterWorker(name, workerType string) error
	UpdateWorkerStatus(name, status string) error
	UpdateWorkerHeartbeat(name string) error
	IncrementOrdersProcessed(name string) error
	GetWorkerByName(name string) (models.Worker, error)
}

type OrderRepository interface {
	UpdateOrderStatus(orderNumber, status, processedBy string) error
	SetOrderCompleted(orderNumber string) error
	GetOrderStatus(orderNumber string) (string, error)
}

type OrderStatusLogRepository interface {
	LogStatusChange(orderNumber, status, changedBy string) error
}

type RabbitMQPublisher interface {
	PublishStatusUpdate(update models.StatusUpdateMessage) error
}
