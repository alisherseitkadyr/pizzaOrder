package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	db "restaurant-system/shared/db"
	"restaurant-system/shared/rabbitmq"
	"sync"
	"syscall"
	"time"

	kitchencmd "restaurant-system/services/kitchen-service/cmd/kitchenservice"
	ordercmd "restaurant-system/services/order-service/cmd/orderservice"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ---------------------------------------------------
	// 1) Подключение к RabbitMQ (одно соединение на процесс)
	// ---------------------------------------------------
	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	client, err := rabbitmq.NewClient(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	// не закрываем client прямо сейчас — подождём пока сервисы завершат работу
	// client.Close() вызовем в конце main после graceful shutdown

	// ---------------------------------------------------
	// 2) Подключение к БД (пример, если сервисам нужна БД)
	// ---------------------------------------------------
	dbPool, err := db.NewPostgresPool()
	if err != nil {
		client.Close()
		log.Fatalf("failed to connect postgres: %v", err)
	}

	// ---------------------------------------------------
	// 3) Запускаем сервисы параллельно
	// ---------------------------------------------------
	var wg sync.WaitGroup

	// Order service
	wg.Add(1)
	go func() {
		defer wg.Done()
		// ordercmd.Start(ctx, client, dbPool)  -- функция Start должна принимать client и dbPool
		if err := ordercmd.Start(ctx, client, dbPool); err != nil {
			log.Printf("order service stopped with error: %v", err)
			cancel() // опционально: при ошибке в сервисе остановить всё
		}
	}()

	// Kitchen service
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := kitchencmd.Start(ctx, client, dbPool); err != nil {
			log.Printf("kitchen service stopped with error: %v", err)
			cancel()
		}
	}()

	// ---------------------------------------------------
	// 4) Graceful shutdown (CTRL+C)
	// ---------------------------------------------------
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
		log.Println("signal received, shutting down...")
		cancel()
	case <-ctx.Done():
		// если кто-то вызвал cancel() - уйдём
	}

	// даём сервисам время корректно завершиться (они слушают ctx.Done())
	shutdownTimeout := time.Second * 10
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		// все сервисы корректно завершились
	case <-time.After(shutdownTimeout):
		log.Println("shutdown timeout reached")
	}

	// закрываем общие ресурсы
	dbPool.Close()
	client.Close()

	log.Println("main finished")
}
