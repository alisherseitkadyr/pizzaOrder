package service

import (
	"errors"
	"fmt"
	"log"
	"restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/ports"
	"strings"
	"time"
)

type KitchenService struct {
	WorkerRepo            ports.WorkerRepository
	OrderRepo             ports.OrderRepository
	StatusLogRepo         ports.OrderStatusLogRepository
	RabbitMQPublisher     ports.RabbitMQPublisher
	WorkerName            string
	SpecializedOrderTypes map[string]bool
	PrefetchCount         int
}

func NewKitchenService(
	workerRepo ports.WorkerRepository,
	orderRepo ports.OrderRepository,
	statusLogRepo ports.OrderStatusLogRepository,
	rabbitMQPublisher ports.RabbitMQPublisher,
	workerName string,
	orderTypes string,
	prefetchCount int,
) *KitchenService {

	specializedTypes := make(map[string]bool)
	if orderTypes != "" {
		for _, ot := range strings.Split(orderTypes, ",") {
			specializedTypes[strings.TrimSpace(ot)] = true
		}
	}

	return &KitchenService{
		WorkerRepo:            workerRepo,
		OrderRepo:             orderRepo,
		StatusLogRepo:         statusLogRepo,
		RabbitMQPublisher:     rabbitMQPublisher,
		WorkerName:            workerName,
		SpecializedOrderTypes: specializedTypes,
		PrefetchCount:         prefetchCount,
	}
}

func (s *KitchenService) RegisterWorker() error {
	log.Printf("Registering worker: %s", s.WorkerName)
	return s.WorkerRepo.RegisterWorker(s.WorkerName, "chef")
}

func (s *KitchenService) CanHandleOrder(orderType string) bool {
	if len(s.SpecializedOrderTypes) == 0 {
		return true // Handle all types if no specialization
	}
	return s.SpecializedOrderTypes[orderType]
}

func (s *KitchenService) ProcessOrder(order models.OrderMessage) error {
	log.Printf("Processing order: %s (type: %s)", order.OrderNumber, order.OrderType)

	// Check if we can handle this order type
	if !s.CanHandleOrder(order.OrderType) {
		log.Printf("Worker %s cannot handle order type: %s", s.WorkerName, order.OrderType)
		return errors.New("cannot handle this order type")
	}

	// Check if order is already being processed
	currentStatus, err := s.OrderRepo.GetOrderStatus(order.OrderNumber)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	// If order is already cooking, skip processing but still acknowledge
	if currentStatus == "cooking" {
		log.Printf("Order %s is already being processed", order.OrderNumber)
		return nil
	}

	// Start transaction-like processing
	if err := s.startCooking(order.OrderNumber); err != nil {
		return fmt.Errorf("failed to start cooking: %w", err)
	}

	// Simulate cooking time
	cookingTime := s.getCookingTime(order.OrderType)
	log.Printf("Cooking order %s for %d seconds", order.OrderNumber, cookingTime/time.Second)
	time.Sleep(cookingTime)

	// Mark order as ready
	if err := s.finishCooking(order.OrderNumber); err != nil {
		return fmt.Errorf("failed to finish cooking: %w", err)
	}

	// Update worker stats
	if err := s.WorkerRepo.IncrementOrdersProcessed(s.WorkerName); err != nil {
		log.Printf("Failed to increment orders processed: %v", err)
	}

	log.Printf("Order %s processed successfully", order.OrderNumber)
	return nil
}

func (s *KitchenService) startCooking(orderNumber string) error {
	// Update order status to cooking
	if err := s.OrderRepo.UpdateOrderStatus(orderNumber, "cooking", s.WorkerName); err != nil {
		return err
	}

	// Log status change
	if err := s.StatusLogRepo.LogStatusChange(orderNumber, "cooking", s.WorkerName); err != nil {
		return err
	}

	// Publish status update
	estimatedReady := time.Now().Add(s.getCookingTime("")).Format(time.RFC3339)
	update := models.StatusUpdateMessage{
		OrderNumber:    orderNumber,
		OldStatus:      "received",
		NewStatus:      "cooking",
		ChangedBy:      s.WorkerName,
		Timestamp:      time.Now().Format(time.RFC3339),
		EstimatedReady: &estimatedReady,
	}

	return s.RabbitMQPublisher.PublishStatusUpdate(update)
}

func (s *KitchenService) finishCooking(orderNumber string) error {
	// Update order status to ready
	if err := s.OrderRepo.SetOrderCompleted(orderNumber); err != nil {
		return err
	}

	// Log status change
	if err := s.StatusLogRepo.LogStatusChange(orderNumber, "ready", s.WorkerName); err != nil {
		return err
	}

	// Publish status update
	update := models.StatusUpdateMessage{
		OrderNumber: orderNumber,
		OldStatus:   "cooking",
		NewStatus:   "ready",
		ChangedBy:   s.WorkerName,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	return s.RabbitMQPublisher.PublishStatusUpdate(update)
}

func (s *KitchenService) getCookingTime(orderType string) time.Duration {
	switch orderType {
	case "dine_in":
		return 8 * time.Second
	case "takeout":
		return 10 * time.Second
	case "delivery":
		return 12 * time.Second
	default:
		return 10 * time.Second
	}
}

func (s *KitchenService) SendHeartbeat() error {
	return s.WorkerRepo.UpdateWorkerHeartbeat(s.WorkerName)
}

func (s *KitchenService) Shutdown() error {
	log.Printf("Shutting down worker: %s", s.WorkerName)
	return s.WorkerRepo.UpdateWorkerStatus(s.WorkerName, "offline")
}
