package main

import (
	"flag"
	"fmt"

	// "log"
	"os"
	"restaurant-system/services/order-service/cmd/orderservice"

	"restaurant-system/services/tracking-service/cmd/trackingservice"
	// "restaurant-system/services/kitchen-service/cmd/kitchenservice"
)

func main() {
	mode := flag.String("mode", "", "Which service to run: order | kitchen | tracking")
	flag.Parse()

	switch *mode {
	case "order":
		orderservice.OrderService()
	// case "kitchen":
	// 	kitchenservice.KitchenService()
	case "tracking":
		trackingservice.TrackingService()
	// case "notification":
	// 	notification.NotificationService()
	default:
		fmt.Println("Usage: restaurant-system --mode=order|kitchen|tracking")
		os.Exit(1)
	}
}
