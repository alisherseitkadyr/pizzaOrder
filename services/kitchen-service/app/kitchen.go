package app

import (
	"context"
	"encoding/json"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	models "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/ports"
	"restaurant-system/services/kitchen-service/utils/logger"
	"time"
)

type KitchenService struct {
	workerSvc *WorkerService
	consumer  ports.MessageConsumer
	publisher ports.MessagePublisher
	Logger    *logger.Logger
}

// Конструктор
func NewKitchenService(workerSvc *WorkerService, consumer ports.MessageConsumer, publisher ports.MessagePublisher, serviceName string) *KitchenService {
	return &KitchenService{
		workerSvc: workerSvc,
		consumer:  consumer,
		publisher: publisher,
		Logger:    logger.New(serviceName),
	}
}

// Запуск слушателя очереди
func (s *KitchenService) Start(ctx context.Context) error {
	return s.consumer.ConsumeOrders(ctx, func(message []byte) error {
		return s.handleOrderMessage(ctx, message)
	})
}

func (s *KitchenService) handleOrderMessage(ctx context.Context, message []byte) error {
	var event domain.OrderCreated
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal OrderCreated: %w", err)
	}

	// найти работника
	worker, err := s.workerSvc.GetAvailableWorker(ctx)
	if err != nil {
		return err
	}

	// отправить статус "cooking"
	status := domain.OrderStatusUpdated{
		ID:          event.ID,
		Status:      string(models.StatusCooking),
		ProcessedBy: worker.Name,
	}
	if err := s.publisher.PublishStatusUpdate(ctx, status); err != nil {
		return err
	}

	// отметить заказ как обработанный у воркера
	if err := s.workerSvc.AddProcessedOrder(ctx, &worker); err != nil {
		return err
	}

	// конкурентно симулируем готовку
	go s.simulateCooking(ctx, event.ID, worker.Name)

	return nil
}

func (s *KitchenService) simulateCooking(ctx context.Context, orderID, workerName string) {
	time.Sleep(8 * time.Second)
	update := domain.OrderStatusUpdated{
		ID:          orderID,
		Status:      string(models.StatusReady),
		ProcessedBy: workerName,
	}
	if err := s.publisher.PublishStatusUpdate(ctx, update); err != nil {
		fmt.Printf("⚠️ failed to publish ready status for order %s: %v\n", orderID, err)
	} else {
		fmt.Printf("✅ Order %s marked as ready (after 8s)\n", orderID)
	}
}
