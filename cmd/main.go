package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	kitchencmd "restaurant-system/services/kitchen-service/cmd/kitchenservice"
	notificationcmd "restaurant-system/services/notification-service/cmd/notificationservice"
	ordercmd "restaurant-system/services/order-service/cmd/orderservice"
	trackingcmd "restaurant-system/services/tracking-service/cmd/trackingservice"
)

func main() {
	// Парсим флаги
	mode := flag.String("mode", "", "Service mode: order-service, kitchen-worker, tracking-service, notification-subscriber")
	port := flag.Int("port", 3000, "HTTP port for services that need it")
	workerName := flag.String("worker-name", "", "Name for kitchen worker")
	orderTypes := flag.String("order-types", "", "Comma-separated order types for kitchen worker")
	prefetch := flag.Int("prefetch", 1, "Prefetch count for RabbitMQ")
	heartbeatInterval := flag.Int("heartbeat-interval", 30, "Heartbeat interval in seconds")
	maxConcurrent := flag.Int("max-concurrent", 50, "Max concurrent orders for order service")

	flag.Parse()

	// Валидация обязательных флагов
	if *mode == "" {
		fmt.Println("Error: --mode flag is required")
		fmt.Println("Available modes: order-service, kitchen-worker, tracking-service, notification-subscriber")
		flag.Usage()
		os.Exit(1)
	}

	if *mode == "kitchen-worker" && *workerName == "" {
		fmt.Println("Error: --worker-name flag is required for kitchen-worker mode")
		flag.Usage()
		os.Exit(1)
	}

	// Контекст и cancel для управления жизненным циклом сервисов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Используем sync.WaitGroup для отслеживания завершения всех горутин
	var wg sync.WaitGroup

	// Запуск выбранного сервиса
	var err error
	switch *mode {
	case "order-service":
		config := ordercmd.Config{
			Port:          *port,
			MaxConcurrent: *maxConcurrent,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = ordercmd.Start(ctx, config)
		}()

	case "kitchen-worker":
		config := kitchencmd.Config{
			WorkerName:        *workerName,
			OrderType:         *orderTypes,
			Prefetch:          *prefetch,
			HeartbeatInterval: *heartbeatInterval,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = kitchencmd.Start(ctx, config)
		}()

	case "tracking-service":
		config := trackingcmd.Config{
			Port: *port,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = trackingcmd.Start(ctx, config)
		}()

	case "notification-subscriber":
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = notificationcmd.Start(ctx)
		}()

	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("Service %s failed: %v", *mode, err)
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Shutdown signal received, shutting down gracefully...")
	cancel() // Прекращаем обработку запросов

	// Ждем завершения всех горутин
	wg.Wait()

	// Даем время для закрытия всех ресурсов (например, базы данных или RabbitMQ)
	time.Sleep(2 * time.Second)
	log.Println("Shutdown completed")
}
