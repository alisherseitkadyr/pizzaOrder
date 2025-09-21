package service

import (
	"encoding/json"
	"log"
	"restaurant-system/services/notification-service/domain/models"
	"time"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) HandleStatusUpdate(update models.StatusUpdateMessage) {
	// Create formatted notification message
	notification := models.Notification{
		OrderNumber: update.OrderNumber,
		OldStatus:   update.OldStatus,
		NewStatus:   update.NewStatus,
		ChangedBy:   update.ChangedBy,
		Message:     s.formatNotificationMessage(update),
	}

	// Print to console (human-readable)
	s.printHumanReadableNotification(notification)

	// Also log in structured JSON format
	s.logStructuredNotification(update)
}

func (s *NotificationService) formatNotificationMessage(update models.StatusUpdateMessage) string {
	message := "Notification for order " + update.OrderNumber +
		": Status changed from '" + update.OldStatus +
		"' to '" + update.NewStatus + "' by " + update.ChangedBy

	if update.EstimatedReady != nil {
		message += ". Estimated ready: " + *update.EstimatedReady
	}

	return message
}

func (s *NotificationService) printHumanReadableNotification(notification models.Notification) {
	log.Println(notification.Message)
}

func (s *NotificationService) logStructuredNotification(update models.StatusUpdateMessage) {
	// Structured JSON log
	logData := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     "DEBUG",
		"service":   "notification-subscriber",
		"action":    "notification_received",
		"message":   "Received status update for order " + update.OrderNumber,
		"details": map[string]interface{}{
			"order_number": update.OrderNumber,
			"new_status":   update.NewStatus,
			"old_status":   update.OldStatus,
			"changed_by":   update.ChangedBy,
		},
	}

	if jsonData, err := json.Marshal(logData); err == nil {
		log.Println(string(jsonData))
	}
}
