package rabbitmq

import (
	"encoding/json"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/utils/logger"
)

type NotificationPublisher struct {
	client *Client
	logger *logger.Logger
}

func NewNotificationPublisher(client *Client, serviceName string) *NotificationPublisher {
	return &NotificationPublisher{
		client: client,
		logger: logger.New(serviceName),
	}
}

func (p *NotificationPublisher) PublishStatusUpdate(update domain.OrderStatus) error {
	messageBytes, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal status update: %w", err)
	}

	err = p.client.Publish("notifications_fanout", "", messageBytes)
	if err != nil {
		p.logger.Error("notification_publish_failed", "Failed to publish status update", "", err)
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	p.logger.Debug("notification_published", "Status update published", "")
	return nil
}
