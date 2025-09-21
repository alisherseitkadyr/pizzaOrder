package kitchenservice

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"restaurant-system/services/kitchen-service/adapters/postgre"
	"restaurant-system/services/kitchen-service/adapters/rabbitmq"
	"restaurant-system/services/kitchen-service/app"
	"restaurant-system/services/kitchen-service/config"
	"restaurant-system/services/kitchen-service/utils/logger"
	"syscall"
	"time"
)

type Config struct {
	WorkerName        string
	OrderType         string
	Prefetch          int
	HeartbeatInterval int
}

func Start(ctx context.Context, cfg Config) error {
	// Инициализация логгера
	serviceName := "kitchen-worker"
	log := logger.New(serviceName)
	log.Info("service_starting", "Kitchen worker starting", "")

	// Создаем cancellable context для graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Загрузка конфигурации
	appConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Подключение к PostgreSQL
	dbPool, err := postgre.NewPostgresPool(appConfig.Database, serviceName)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer dbPool.Close()

	// Подключение к RabbitMQ
	rabbitClient, err := rabbitmq.NewClient(appConfig.RabbitMQ, serviceName)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer rabbitClient.Close()

	// Декларация обменников
	if err := rabbitClient.DeclareExchange("orders_topic", "topic"); err != nil {
		return fmt.Errorf("failed to declare orders_topic exchange: %w", err)
	}
	if err := rabbitClient.DeclareExchange("notifications_fanout", "fanout"); err != nil {
		return fmt.Errorf("failed to declare notifications_fanout exchange: %w", err)
	}

	// Инициализация репозиториев и сервисов
	workerRepo := postgre.NewPostgresWorkerRepo(dbPool, serviceName)
	kitchenRepo := postgre.NewPostgresKitchenRepo(dbPool, serviceName)

	// Создание потребителя
	consumer, err := rabbitmq.NewKitchenConsumer(rabbitClient, cfg.Prefetch, cfg.OrderType)
	if err != nil {
		return fmt.Errorf("failed to create kitchen consumer: %w", err)
	}

	// Создание издателя
	publisher := rabbitmq.NewNotificationPublisher(rabbitClient, serviceName)

	// Создание сервисов
	workerSvc := app.NewWorkerService(workerRepo, serviceName)
	kitchenSvc := app.NewKitchenService(workerSvc, consumer, publisher, kitchenRepo, cfg.WorkerName, serviceName)

	// Регистрация воркера
	if err := workerSvc.RegisterWorker(ctx, cfg.WorkerName, cfg.OrderType); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	// Запускаем heartbeat в отдельной goroutine
	heartbeatCtx, heartbeatCancel := context.WithCancel(ctx)
	defer heartbeatCancel()

	go workerSvc.StartHeartbeat(heartbeatCtx, cfg.WorkerName, time.Duration(cfg.HeartbeatInterval)*time.Second)

	// Канал для ошибок из kitchen service
	serviceErr := make(chan error, 1)

	// Запускаем обработку заказов в отдельной goroutine
	go func() {
		log.Info("service_started", fmt.Sprintf("Worker %s started processing %s orders", cfg.WorkerName, cfg.OrderType), "")
		if err := kitchenSvc.Start(ctx); err != nil {
			serviceErr <- fmt.Errorf("kitchen service failed: %w", err)
		}
	}()

	// Обработка сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ждем либо сигнала завершения, либо ошибки сервиса
	select {
	case sig := <-sigChan:
		log.Info("shutdown_signal", fmt.Sprintf("Received signal: %s, initiating graceful shutdown", sig), "")

		// Инициируем graceful shutdown
		cancel()

		// Даем время на завершение обработки
		shutdownTimeout := 30 * time.Second
		select {
		case <-time.After(shutdownTimeout):
			log.Info("shutdown_timeout", "Shutdown timeout reached, forcing exit", "")
		case err := <-serviceErr:
			if err != nil {
				log.Error("shutdown_error", "Service error during shutdown", "", err)
			}
		}

	case err := <-serviceErr:
		if err != nil {
			log.Error("service_error", "Service stopped with error", "", err)
			cancel() // Отменяем контекст при ошибке
			return err
		}
	}

	// Final cleanup - отмечаем воркера как offline
	log.Info("shutdown_cleanup", "Performing final cleanup", "")
	if err := workerSvc.SetWorkerOffline(ctx, cfg.WorkerName); err != nil {
		log.Error("cleanup_error", "Failed to set worker offline during cleanup", "", err)
	}

	log.Info("service_stopped", "Kitchen worker stopped gracefully", "")
	return nil
}
