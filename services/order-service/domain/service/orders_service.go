package service

import (
	"context"
	"fmt"
	"regexp"
	"restaurant-system/services/order-service/domain/models"
	"restaurant-system/services/order-service/domain/ports"
	"time"
)

type OrderService struct {
	OrderRepository    ports.OrderRepository
	RabbitMQPublisher  ports.RabbitMQPublisher
	OrderNumberService *OrderNumberService
}

func NewOrderService(repo ports.OrderRepository, publisher ports.RabbitMQPublisher) *OrderService {
	return &OrderService{
		OrderRepository:    repo,
		RabbitMQPublisher:  publisher,
		OrderNumberService: NewOrderNumberService(repo),
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, customerName, orderType string, items []models.OrderItemRequest, tableNumber *int, deliveryAddress *string) (*models.Order, error) {
	// Validate order
	if err := validateOrder(customerName, orderType, items, tableNumber, deliveryAddress); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Calculate total amount and priority
	totalAmount := calculateTotalAmount(items)
	priority := calculatePriority(totalAmount)

	// Generate order number (transactional and daily reset)
	orderNumber, err := s.OrderNumberService.GenerateOrderNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate order number: %w", err)
	}

	// Create order object
	order := &models.Order{
		OrderNumber:     orderNumber,
		CustomerName:    customerName,
		OrderType:       orderType,
		TableNumber:     tableNumber,
		DeliveryAddress: deliveryAddress,
		TotalAmount:     totalAmount,
		Priority:        priority,
	}
	var itemsDb []models.OrderItem
	for _, item := range items {
		var itemDb models.OrderItem
		itemDb.Name = item.Name
		itemDb.Quantity = item.Quantity
		itemDb.Price = item.Price

		itemsDb = append(itemsDb, itemDb)
	}

	// Save order with items and status log in single transaction
	err = s.OrderRepository.SaveOrderWithItems(ctx, order, itemsDb)
	if err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}
	orderMes := &models.OrderMessage{
		OrderNumber:     orderNumber,
		CustomerName:    customerName,
		OrderType:       orderType,
		TableNumber:     tableNumber,
		DeliveryAddress: deliveryAddress,
		TotalAmount:     totalAmount,
		Priority:        priority,
	}
	// Publish to RabbitMQ
	err = s.RabbitMQPublisher.PublishOrder(orderMes)
	if err != nil {
		// Note: Order is already saved, this is a non-critical error
		// We might want to implement retry logic or dead letter queue
		return nil, fmt.Errorf("failed to publish order to RabbitMQ: %w", err)
	}

	return order, nil
}

// OrderNumberService handles transactional order number generation
type OrderNumberService struct {
	repo ports.OrderRepository
}

func NewOrderNumberService(repo ports.OrderRepository) *OrderNumberService {
	return &OrderNumberService{repo: repo}
}

func (s *OrderNumberService) GenerateOrderNumber(ctx context.Context) (string, error) {
	datePrefix := time.Now().UTC().Format("20060102")

	// This should be implemented with database sequence or atomic counter
	// For now using timestamp-based approach
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	sequence := timestamp % 1000

	return fmt.Sprintf("ORD_%s_%03d", datePrefix, sequence), nil
}

// Enhanced validation function
func validateOrder(customerName, orderType string, items []models.OrderItemRequest, tableNumber *int, deliveryAddress *string) error {
	// Validate customer name
	if customerName == "" {
		return fmt.Errorf("customer_name is required")
	}
	if len(customerName) > 100 {
		return fmt.Errorf("customer_name must be 100 characters or less")
	}

	// Validate customer name contains only allowed characters
	validNameRegex := regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ\s\-']+$`)
	if !validNameRegex.MatchString(customerName) {
		return fmt.Errorf("customer_name contains invalid characters")
	}

	// Validate order type
	validOrderTypes := map[string]bool{
		"dine_in":  true,
		"takeout":  true,
		"delivery": true,
	}
	if !validOrderTypes[orderType] {
		return fmt.Errorf("order_type must be one of: dine_in, takeout, delivery")
	}

	// Validate items
	if len(items) == 0 {
		return fmt.Errorf("items cannot be empty")
	}
	if len(items) > 20 {
		return fmt.Errorf("maximum 20 items allowed per order")
	}

	for i, item := range items {
		if item.Name == "" {
			return fmt.Errorf("item[%d].name is required", i)
		}
		if len(item.Name) > 50 {
			return fmt.Errorf("item[%d].name must be 50 characters or less", i)
		}
		if item.Quantity < 1 || item.Quantity > 10 {
			return fmt.Errorf("item[%d].quantity must be between 1 and 10", i)
		}
		if item.Price < 0.01 || item.Price > 999.99 {
			return fmt.Errorf("item[%d].price must be between 0.01 and 999.99", i)
		}
	}

	// Conditional validation based on order type
	switch orderType {
	case "dine_in":
		if tableNumber == nil {
			return fmt.Errorf("table_number is required for dine_in orders")
		}
		if *tableNumber < 1 || *tableNumber > 100 {
			return fmt.Errorf("table_number must be between 1 and 100")
		}
		if deliveryAddress != nil {
			return fmt.Errorf("delivery_address should not be provided for dine_in orders")
		}
	case "delivery":
		if deliveryAddress == nil {
			return fmt.Errorf("delivery_address is required for delivery orders")
		}
		if len(*deliveryAddress) < 10 {
			return fmt.Errorf("delivery_address must be at least 10 characters")
		}
		if tableNumber != nil {
			return fmt.Errorf("table_number should not be provided for delivery orders")
		}
	case "takeout":
		if tableNumber != nil {
			return fmt.Errorf("table_number should not be provided for takeout orders")
		}
		if deliveryAddress != nil {
			return fmt.Errorf("delivery_address should not be provided for takeout orders")
		}
	}

	return nil
}

// Helper functions
func calculateTotalAmount(items []models.OrderItemRequest) float64 {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}

func calculatePriority(totalAmount float64) int {
	switch {
	case totalAmount > 100:
		return 10
	case totalAmount >= 50:
		return 5
	default:
		return 1
	}
}
