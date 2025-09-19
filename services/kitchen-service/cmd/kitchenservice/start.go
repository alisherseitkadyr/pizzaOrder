package kitchenservice

import (
	"context"
	"fmt"
	"restaurant-system/services/kitchen-service/adapters/postgre"
	"restaurant-system/services/kitchen-service/adapters/rabbitmq"
	"restaurant-system/services/kitchen-service/app"
	"restaurant-system/services/kitchen-service/config"
	"restaurant-system/services/kitchen-service/utils/logger"
	"strings"
	"time"
)

type Config struct {
	WorkerName        string
	OrderTypes        string
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

	// Parse order types
	var orderTypes []string
	if cfg.OrderTypes != "" {
		orderTypes = strings.Split(cfg.OrderTypes, ",")
		for i := range orderTypes {
			orderTypes[i] = strings.TrimSpace(orderTypes[i])
		}
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

	// Create consumer for kitchen orders
	consumer, err := rabbitmq.NewKitchenConsumer(rabbitClient, cfg.Prefetch, orderTypes)
	if err != nil {
		return fmt.Errorf("failed to create kitchen consumer: %w", err)
	}

	// Create publisher for notifications
	publisher := rabbitmq.NewNotificationPublisher(rabbitClient, serviceName)

	// Create worker service
	workerSvc := app.NewWorkerService(workerRepo, serviceName)

	// Create kitchen service
	kitchenSvc := app.NewKitchenService(workerSvc, consumer, publisher, serviceName)

	// Register worker
	if err := workerSvc.RegisterWorker(ctx, cfg.WorkerName, orderTypes); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
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
				if err := workerSvc.SendHeartbeat(ctx, cfg.WorkerName); err != nil {
					logger.Error("heartbeat_failed", "Failed to send heartbeat", cfg.WorkerName, err)
				}
			}
		}
	}()
	// Start processing orders
	logger.Info("worker_registered", fmt.Sprintf("Worker %s started processing orders", cfg.WorkerName), "")

	if err := kitchenSvc.Start(ctx); err != nil {
		return fmt.Errorf("kitchen service failed: %w", err)
	}

	return nil
}
