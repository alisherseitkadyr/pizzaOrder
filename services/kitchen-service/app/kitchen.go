package app

import (
	"context"
	"encoding/json"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/ports"
	"restaurant-system/shared/events"
	"time"
)

type KitchenService struct {
	workerSvc WorkerService
	consumer  ports.MessageConsumer
	publisher ports.MessagePublisher
}

func (s *KitchenService) Start(ctx context.Context) error {
	return s.consumer.ConsumeOrders(ctx, func(message []byte) error {
		return s.handleOrderMessage(ctx, message)
	})
}

func (s *KitchenService) handleOrderMessage(ctx context.Context, message []byte) error {
	// получаем заказ с ребита
	var event events.OrderCreated
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal OrderCreated: %w", err)
	}
	// найти подходящего работника
	worker, err := s.workerSvc.GetAvailableWorker(ctx)
	if err != nil {
		return err
	}

	// Публикация статуса заказа на готовку
	var status_event events.OrderStatusUpdated
	status_event.ID = event.ID
	status_event.ProcessedBy = worker.Name
	status_event.Status = string(domain.StatusCooking)
	err = s.publisher.PublishStatusUpdate(ctx, status_event)
	if err != nil {
		return err
	}

	err = s.workerSvc.AddProcessedOrder(ctx, &worker)
	if err != nil {
		return err
	}

	go s.simulateCooking(ctx, event.ID, worker.Name)
	return nil
}

func (s *KitchenService) simulateCooking(ctx context.Context, orderID, workerName string) {
	time.Sleep(8 * time.Second)
	update := events.OrderStatusUpdated{
		ID:          orderID,
		Status:      string(domain.StatusReady),
		ProcessedBy: workerName,
	}
	if err := s.publisher.PublishStatusUpdate(ctx, update); err != nil {
		fmt.Printf("⚠️ failed to publish ready status for order %s: %v\n", orderID, err)
	} else {
		fmt.Printf("✅ Order %s marked as ready (after %v)\n", orderID, "8 second")
	}
}
