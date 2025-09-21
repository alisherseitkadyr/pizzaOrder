package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"restaurant-system/services/tracking-service/cmd/trackingservice"
	"restaurant-system/services/kitchen-service/cmd/kitchenservice"
	"restaurant-system/services/notification-service/cmd/notificationservice"
	"restaurant-system/services/order-service/cmd/orderservice"
)

func main() {
	// Common flag
	mode := flag.String("mode", "", "Which service to run: order | kitchen | kitchen-worker | tracking | notification")

	// Kitchen worker flags
	workerName := flag.String("worker-name", "", "Unique name for the worker (e.g., chef_mario)")
	orderTypes := flag.String("order-types", "", "Comma-separated list of order types the worker can handle (default: all)")
	heartbeatInterval := flag.Int("heartbeat-interval", 30, "Interval (seconds) between heartbeats")
	prefetch := flag.Int("prefetch", 1, "RabbitMQ prefetch count")

	flag.Parse()

	if *mode == "" {
		fmt.Println("Usage: ./restaurant-system --mode=<service>")
		os.Exit(1)
	}

	switch *mode {
	case "order":
		log.Println("Starting Order Service...")
		orderservice.OrderService()

	case "kitchen-service":
		if *workerName == "" {
			log.Fatal("--worker-name is required when running in kitchen-worker mode")
		}
		log.Printf("Starting Kitchen Worker: %s (orderTypes=%s, heartbeat=%ds, prefetch=%d)\n",
			*workerName, *orderTypes, *heartbeatInterval, *prefetch)
		kitchenservice.KitchenService(*workerName, *orderTypes, *heartbeatInterval, *prefetch)

	case "tracking":
		log.Println("Starting Tracking Service...")
		trackingservice.TrackingService()

	case "notification":
		log.Println("Starting Notification Service...")
		notificationservice.NotificationService()

	default:
		log.Fatalf("Unknown mode: %s. Expected one of: order | kitchen | kitchen-worker | tracking | notification", *mode)
	}
}
