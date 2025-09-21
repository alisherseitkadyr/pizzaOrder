package notificationservice

import (
	"context"
	"log"
	"os"
	"os/signal"
	"restaurant-system/services/notification-service/adapters/rabbitmq"
	"restaurant-system/services/notification-service/domain/service"
	"syscall"
	"time"
)

func Start(ctx context.Context) error {
	// Подключаемся к RabbitMQ
	rabbitURL := "amqp://guest:guest@localhost:5672/"
	client, err := rabbitmq.NewClient(rabbitURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer client.Close()

	log.Println("Connected to RabbitMQ")

	// Создаем потребителя уведомлений
	consumer := rabbitmq.NewNotificationConsumer(client)

	// Настроим обменник и очередь
	if err := consumer.Setup(); err != nil {
		log.Fatal("Failed to setup RabbitMQ:", err)
	}

	log.Println("RabbitMQ setup completed")

	// Создаем сервис для обработки уведомлений
	notificationService := service.NewNotificationService()

	// Начинаем потреблять сообщения
	if err := consumer.StartConsuming(notificationService.HandleStatusUpdate); err != nil {
		log.Fatal("Failed to start consuming:", err)
	}

	log.Println("Notification service started. Waiting for messages...")
	log.Println("Press Ctrl+C to exit")

	// Создаем канал для получения сигналов завершения работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигнал завершения
	select {
	case <-stop:
		log.Println("Shutting down notification service...")
		// Делаем паузу перед завершением, чтобы успеть завершить текущие задачи
		time.Sleep(2 * time.Second)
	}

	return nil
}
