package kitchenservice

import (
	"context"
	"fmt"
	"os"
	"restaurant-system/services/kitchen-service/adapters/postgre"
	"restaurant-system/services/kitchen-service/adapters/rabbitmq"
	"restaurant-system/services/kitchen-service/app"
	"restaurant-system/services/kitchen-service/config"
	"restaurant-system/services/kitchen-service/utils/logger"
	"time"
)

type Config struct {
	WorkerName        string
	OrderType         string // теперь один тип, а не слайс
	Prefetch          int
	HeartbeatInterval int
}

func Start(ctx context.Context, cfg Config) error {
	// Initialize logger
	serviceName := "kitchen-worker"
	logger := logger.New(serviceName)
	logger.Info("service_starting", "Kitchen worker starting", "")

	// Load configuration
	appConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to PostgreSQL
	dbPool, err := postgre.NewPostgresPool(appConfig.Database, serviceName)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer dbPool.Close()

	// Connect to RabbitMQ
	rabbitClient, err := rabbitmq.NewClient(appConfig.RabbitMQ, serviceName)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer rabbitClient.Close()

	// Declare exchanges
	if err := rabbitClient.DeclareExchange("orders_topic", "topic"); err != nil {
		return fmt.Errorf("failed to declare orders_topic exchange: %w", err)
	}
	if err := rabbitClient.DeclareExchange("notifications_fanout", "fanout"); err != nil {
		return fmt.Errorf("failed to declare notifications_fanout exchange: %w", err)
	}

	// Initialize repositories and services
	workerRepo := postgre.NewPostgresWorkerRepo(dbPool, serviceName)
	kitchenRepo := postgre.NewPostgresKitchenRepo(dbPool, serviceName)

	// Create consumer for kitchen orders (теперь один тип)
	consumer, err := rabbitmq.NewKitchenConsumer(rabbitClient, cfg.Prefetch, cfg.OrderType)
	if err != nil {
		return fmt.Errorf("failed to create kitchen consumer: %w", err)
	}

	// Create publisher for notifications
	publisher := rabbitmq.NewNotificationPublisher(rabbitClient, serviceName)

	// Create worker service
	workerSvc := app.NewWorkerService(workerRepo, serviceName)

	// Create kitchen service
	kitchenSvc := app.NewKitchenService(workerSvc, consumer, publisher, kitchenRepo, cfg.WorkerName, serviceName)

	// Ensure worker registered

	_, err = workerRepo.GetByName(ctx, cfg.WorkerName)
	if err == nil {
		return fmt.Errorf("worker already existed : %w", err)
		os.Exit(1)
	} else {
		err := workerSvc.RegisterWorker(ctx, cfg.WorkerName, cfg.OrderType)
		if err != nil {
			return fmt.Errorf("failed to ensure worker registered: %w", err)
		}
	}

	if err := workerSvc.EnsureRegistered(ctx, cfg.WorkerName, cfg.OrderType); err != nil {
		return fmt.Errorf("failed to ensure worker registered: %w", err)
	}

	// Start heartbeat
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.HeartbeatInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := workerSvc.Heartbeat(ctx, cfg.WorkerName); err != nil {
					logger.Error("heartbeat_failed", "Failed to send heartbeat", cfg.WorkerName, err)
				}
			}
		}
	}()

	// Start processing orders
	logger.Info("worker_registered", fmt.Sprintf("Worker %s started processing %s orders", cfg.WorkerName, cfg.OrderType), "")

	if err := kitchenSvc.Start(ctx); err != nil {
		return fmt.Errorf("kitchen service failed: %w", err)
	}

	return nil
}
