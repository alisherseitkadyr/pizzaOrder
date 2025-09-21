package notificationservice

import (
	"log"
	"os"
	"os/signal"
	"restaurant-system/services/notification-service/adapters/rabbitmq"
	"restaurant-system/services/notification-service/domain/service"
	"syscall"
)

func NotificationService() {
	// Connect to RabbitMQ
	rabbitURL := "amqp://guest:guest@localhost:5672/"
	client, err := rabbitmq.NewClient(rabbitURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer client.Close()

	log.Println("Connected to RabbitMQ")

	// Create notification consumer
	consumer := rabbitmq.NewNotificationConsumer(client)

	// Setup exchange and queue
	if err := consumer.Setup(); err != nil {
		log.Fatal("Failed to setup RabbitMQ:", err)
	}

	log.Println("RabbitMQ setup completed")

	// Create notification service
	notificationService := service.NewNotificationService()

	// Start consuming messages
	if err := consumer.StartConsuming(notificationService.HandleStatusUpdate); err != nil {
		log.Fatal("Failed to start consuming:", err)
	}

	log.Println("Notification service started. Waiting for messages...")
	log.Println("Press Ctrl+C to exit")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Keep the service running
	select {
	case <-sigChan:
		log.Println("Shutting down notification service...")
	}
}
