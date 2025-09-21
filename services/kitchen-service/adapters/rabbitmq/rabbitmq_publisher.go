package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/ports"
	"restaurant-system/services/kitchen-service/utils/logger"
	"time"
)

type NotificationPublisher struct {
	client *Client
	logger *logger.Logger
}

func NewNotificationPublisher(client *Client, serviceName string) ports.StatusPublisher {
	return &NotificationPublisher{
		client: client,
		logger: logger.New(serviceName),
	}
}

func (p *NotificationPublisher) PublishStatusUpdate(ctx context.Context, event domain.OrderStatusUpdated) error {
	messageBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal status update: %w", err)
	}

	return p.client.Publish("notifications_fanout", "", messageBytes)
}

func (p *NotificationPublisher) PublishCookingStarted(ctx context.Context, order domain.OrderMessage, workerName string) error {
	event := domain.OrderStatusUpdated{
		OrderNumber: order.OrderNumber,
		OldStatus:   string(domain.StatusReceived),
		NewStatus:   string(domain.StatusCooking),
		ChangedBy:   workerName,
		Timestamp:   time.Now(),
	}

	return p.PublishStatusUpdate(ctx, event)
}

func (p *NotificationPublisher) PublishOrderReady(ctx context.Context, order domain.OrderMessage, workerName string) error {
	event := domain.OrderStatusUpdated{
		OrderNumber: order.OrderNumber,
		OldStatus:   string(domain.StatusCooking),
		NewStatus:   string(domain.StatusReady),
		ChangedBy:   workerName,
		Timestamp:   time.Now(),
	}

	return p.PublishStatusUpdate(ctx, event)
}
