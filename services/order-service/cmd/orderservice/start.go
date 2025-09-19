package orderservice

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"restaurant-system/services/order-service/adapters/postgres"
	"restaurant-system/services/order-service/adapters/rabbitmq"
	"restaurant-system/services/order-service/adapters/web"
	"restaurant-system/services/order-service/config"
	"restaurant-system/services/order-service/domain/service"
	"restaurant-system/services/order-service/utils/logger"
	"syscall"
	"time"
)

type Config struct {
	Port          int
	MaxConcurrent int
}

func Start(ctx context.Context, cfg Config) error {
	// Initialize logger
	serviceName := "order-service"
	logger := logger.New(serviceName)
	logger.Info("service_starting", "Order service starting", "")

	// Load configuration
	appConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to PostgreSQL
	dbPool, err := postgres.NewPostgresPool(appConfig.Database, serviceName)
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

	// Initialize repositories and services
	orderRepo := postgres.NewPostgresOrderRepository(dbPool, serviceName)
	rabbitPublisher := rabbitmq.NewRabbitMQPublisher(rabbitClient, serviceName)

	orderService := service.NewOrderService(orderRepo, rabbitPublisher)

	// HTTP handler
	webHandler := web.NewWebHandler(orderService, serviceName)
	router := web.NewRouter(webHandler)

	// HTTP server
	port := cfg.Port
	if port == 0 {
		port = 3000
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server_starting", fmt.Sprintf("Starting HTTP server on port %d", port), "")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown_requested", "Shutdown requested via context", "")
	case sig := <-sigChan:
		logger.Info("shutdown_signal", fmt.Sprintf("Received signal: %s", sig), "")
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("service_stopped", "Order service stopped gracefully", "")
	return nil
}
