package app

import (
	"context"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/ports"
	"restaurant-system/services/kitchen-service/utils/logger"
	"time"
)

type KitchenService struct {
	workerService    *WorkerService
	orderConsumer    ports.MessageConsumer
	statusPublisher  ports.StatusPublisher
	kitchenOrderRepo ports.KitchenOrderRepository
	workerName       string
	logger           *logger.Logger
}

func NewKitchenService(
	workerService *WorkerService,
	orderConsumer ports.MessageConsumer,
	statusPublisher ports.StatusPublisher,
	kitchenOrderRepo ports.KitchenOrderRepository,
	workerName string,
	serviceName string,
) *KitchenService {
	return &KitchenService{
		workerService:    workerService,
		orderConsumer:    orderConsumer,
		statusPublisher:  statusPublisher,
		kitchenOrderRepo: kitchenOrderRepo,
		workerName:       workerName,
		logger:           logger.New(serviceName),
	}
}

func (s *KitchenService) Start(ctx context.Context) error {
	s.logger.Info("service_starting", "Starting kitchen service", s.workerName)

	// Начинаем потреблять заказы
	messages, err := s.orderConsumer.ConsumeOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to consume orders: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("service_stopping", "Stopping kitchen service", s.workerName)
			return nil
		case msg, ok := <-messages:
			if !ok {
				return fmt.Errorf("orders channel closed")
			}
			go s.processOrder(ctx, msg)
		}
	}
}

func (s *KitchenService) processOrder(ctx context.Context, msg domain.OrderMessage) {
	orderNumber := msg.OrderNumber
	requestID := fmt.Sprintf("order_%s", orderNumber)

	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("order_panic", fmt.Sprintf("Panic processing order: %v", r), requestID, nil)
			_ = s.orderConsumer.NackMessage(msg, true)
		}
	}()

	s.logger.Info("order_received", fmt.Sprintf("Processing order %s", orderNumber), requestID)

	// cooking started
	if err := s.kitchenOrderRepo.UpdateOrderStatus(ctx, msg.OrderNumber, domain.StatusCooking, s.workerName); err != nil {
		s.logger.Error("update_status_failed", "Failed to update cooking status", requestID, err)
		_ = s.orderConsumer.NackMessage(msg, true)
		return
	}
	if err := s.statusPublisher.PublishCookingStarted(ctx, msg, s.workerName); err != nil {
		s.logger.Error("event_publish_failed", "Failed to publish cooking event", requestID, err)
	}

	// ограничиваем время готовки
	cookingCtx, cancel := context.WithTimeout(ctx, msg.CookingTime())
	defer cancel()

	select {
	case <-cookingCtx.Done():
		if cookingCtx.Err() == context.DeadlineExceeded {
			s.logger.Error("cooking_timeout", "Cooking time exceeded", requestID, cookingCtx.Err())
			_ = s.orderConsumer.NackMessage(msg, true)
			return
		}
	case <-time.After(msg.CookingTime()):
		// готово
	}

	// обновляем статус → ready
	if err := s.kitchenOrderRepo.UpdateOrderStatus(ctx, orderNumber, domain.StatusReady, s.workerName); err != nil {
		s.logger.Error("status_update_failed", "Failed to update order to ready", requestID, err)
		_ = s.orderConsumer.NackMessage(msg, true)
		return
	}

	if err := s.statusPublisher.PublishOrderReady(ctx, msg, s.workerName); err != nil {
		s.logger.Error("event_publish_failed", "Failed to publish ready event", requestID, err)
	}

	// обновляем статистику
	if err := s.workerService.AddProcessedOrder(ctx, s.workerName); err != nil {
		s.logger.Error("worker_update_failed", "Failed to update worker stats", requestID, err)
	}

	// подтверждаем сообщение
	if err := s.orderConsumer.AckMessage(msg); err != nil {
		s.logger.Error("ack_failed", "Failed to ack message", requestID, err)
	}
}
