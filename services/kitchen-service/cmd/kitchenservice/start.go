package kitchenservice

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"restaurant-system/services/kitchen-service/adapters/postgres"
	"restaurant-system/services/kitchen-service/adapters/rabbitmq"
	"restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/service"
	"syscall"
	"time"
)

func KitchenService(workerName *string, prefetch *int,heartbeatInterval *int, orderTypes *string) {
	// Parse command line flags
	flag.Parse()

	if *workerName == "" {
		log.Fatal("worker-name is required")
	}

	// Connect to PostgreSQL
	dbPool, err := postgres.NewPostgresPool()
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer dbPool.Close()

	// Connect to RabbitMQ
	rabbitClient, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitClient.Close()

	// Set prefetch count
	if err := rabbitClient.SetPrefetchCount(*prefetch); err != nil {
		log.Fatal("Failed to set prefetch count:", err)
	}

	// Ensure exchanges exist
	if err := rabbitClient.DeclareExchange("orders_topic", "topic"); err != nil {
		log.Fatal("Failed to declare orders_topic exchange:", err)
	}
	if err := rabbitClient.DeclareExchange("notifications_fanout", "fanout"); err != nil {
		log.Fatal("Failed to declare notifications_fanout exchange:", err)
	}

	// Initialize repositories
	workerRepo := &postgres.PostgresWorkerRepository{DB: dbPool}
	orderRepo := &postgres.PostgresOrderRepository{DB: dbPool}
	statusLogRepo := &postgres.PostgresOrderStatusLogRepository{DB: dbPool}
	rabbitPublisher := &rabbitmq.RabbitMQPublisher{Client: rabbitClient}

	// Initialize kitchen service
	kitchenService := service.NewKitchenService(
		workerRepo,
		orderRepo,
		statusLogRepo,
		rabbitPublisher,
		*workerName,
		*orderTypes,
		*prefetch,
	)

	// Register worker
	if err := kitchenService.RegisterWorker(); err != nil {
		log.Fatal("Failed to register worker:", err)
	}
	log.Printf("Worker %s registered successfully", *workerName)

	// Initialize consumer
	consumer := rabbitmq.NewKitchenConsumer(rabbitClient)

	// Start consuming orders
	go func() {
		if err := consumer.ConsumeOrders("kitchen_queue", *workerName, func(orderMsg models.OrderMessage) {
			// Convert OrderMessage directly (you don’t need Marshal/Unmarshal here actually)
			// If you still want JSON conversion, use orderMsg, not order
			jsonData, _ := json.Marshal(orderMsg)
			var orderMessage models.OrderMessage
			if err := json.Unmarshal(jsonData, &orderMessage); err != nil {
				log.Printf("Failed to parse order message: %v", err)
				return
			}
	
			// Process order
			if err := kitchenService.ProcessOrder(orderMessage); err != nil {
				log.Printf("Failed to process order: %v", err)
				// In a real implementation, you would nack the message here
			}
			// In a real implementation, you would ack the message here
		}); err != nil {
			log.Fatal("Failed to start consuming:", err)
		}
	}()
	

	// Start heartbeat routine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(time.Duration(*heartbeatInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := kitchenService.SendHeartbeat(); err != nil {
					log.Printf("Failed to send heartbeat: %v", err)
				} else {
					log.Printf("Heartbeat sent for worker %s", *workerName)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Printf("Kitchen worker %s started. Press Ctrl+C to exit.", *workerName)

	<-sigChan
	log.Println("Shutting down...")

	// Graceful shutdown
	if err := kitchenService.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Worker shutdown completed")
}
