package trackingservice

import (
	"log"
	"net/http"
	"os"
	"time"

	"restaurant-system/services/tracking-service/adapters/postgres"
	"restaurant-system/services/tracking-service/adapters/web"
	"restaurant-system/services/tracking-service/domain/service"
)

func TrackingService() {
	// Connect to PostgreSQL
	dbPool, err := postgres.NewPostgresPool()
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer dbPool.Close()

	log.Println("Connected to PostgreSQL database")

	// Initialize repositories
	orderRepo := &postgres.PostgresOrderRepository{DB: dbPool}
	workerRepo := &postgres.PostgresWorkerRepository{DB: dbPool}

	// Initialize tracking service
	trackingService := service.NewTrackingService(orderRepo, workerRepo)

	// Initialize web handler
	webHandler := web.NewWebHandler(trackingService)
	router := web.NewRouter(webHandler)

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Tracking service started on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}