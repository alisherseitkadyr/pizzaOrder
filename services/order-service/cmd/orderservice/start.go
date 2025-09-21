package orderservice

import (
	"log"
	"net/http"
	"os"
	"restaurant-system/services/order-service/adapters/postgres"
	"restaurant-system/services/order-service/adapters/rabbitmq"
	"restaurant-system/services/order-service/adapters/web"
	"restaurant-system/services/order-service/domain/service"
	"time"
)
// services/order-service/cmd/orderservice/start.go

func OrderService() {
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

	// Declare exchange
	err = rabbitClient.DeclareExchange("orders_topic")
	if err != nil {
		log.Fatal("Failed to declare exchange:", err)
	}

	// ✅ Declare & bind a queue (example: "kitchen_orders")
	err = rabbitClient.DeclareAndBindQueue(
		"kitchen_orders",   // queue name
		"orders_topic",     // exchange
		"kitchen.*.*",      // routing key pattern
	)
	if err != nil {
		log.Fatal("Failed to declare/bind queue:", err)
	}

	// Initialize repositories
	orderRepo := &postgres.PostgresOrderRepository{DB: dbPool}
	orderItemRepo := &postgres.PostgresOrderItemRepository{DB: dbPool}
	statusLogRepo := &postgres.PostgresOrderStatusLogRepository{DB: dbPool}

	// Initialize RabbitMQ publisher
	rabbitPublisher := &rabbitmq.RabbitMQPublisher{Client: rabbitClient}

	// Initialize order service
	orderService := service.OrderService{
		OrderRepository:   orderRepo,
		OrderItemRepo:     orderItemRepo,
		StatusLogRepo:     statusLogRepo,
		RabbitMQPublisher: rabbitPublisher,
	}

	// Initialize web handler
	webHandler := web.NewWebHandler(orderService)
	router := web.NewRouter(webHandler)

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Order service started on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
