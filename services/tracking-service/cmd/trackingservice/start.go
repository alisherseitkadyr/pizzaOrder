package trackingservice

import (
	"context"
	"fmt"
	"net/http"
	"restaurant-system/services/tracking-service/adapters/postgres"
	"restaurant-system/services/tracking-service/adapters/web"
	"restaurant-system/services/tracking-service/config"
	"restaurant-system/services/tracking-service/domain/service"
	"restaurant-system/services/tracking-service/utils/logger"
	"time"
)

type Config struct {
	Port int
}

func Start(ctx context.Context, cfg Config) error {
	serviceName := "tracking-service"
	logger := logger.New(serviceName)
	logger.Info("service_starting", "Tracking service starting", "")

	// Load configuration
	appConfig, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed_to_load_config", "Failed to load config", "", err)
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to PostgreSQL
	dbPool, err := postgres.NewPostgresPool(appConfig.Database, serviceName)
	if err != nil {
		logger.Error("failed_to_connect_db", "Failed to connect to PostgreSQL", "", err)
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer dbPool.Close()

	logger.Info("db_connected", "Connected to PostgreSQL database", "")

	// Initialize repositories
	orderRepo := postgres.NewPostgresOrderRepository(dbPool, serviceName)
	workerRepo := postgres.NewPostgresWorkerRepository(dbPool, serviceName)

	// Initialize tracking service
	trackingService := service.NewTrackingService(orderRepo, workerRepo)

	// Initialize web handler
	webHandler := web.NewWebHandler(trackingService)
	router := web.NewRouter(webHandler)

	// Start HTTP server
	port := cfg.Port
	if port == 0 {
		port = 3002
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("service_started", fmt.Sprintf("Tracking service started on port %d", port), "")

	// Start the server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server_start_failed", "Failed to start server", "", err)
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
