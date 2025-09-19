package kitchenservice

import (
	"context"
	"log"
	"restaurant-system/services/kitchen-service/adapters/postgre"
	"restaurant-system/services/kitchen-service/adapters/rabbitmq"
	"restaurant-system/services/kitchen-service/app"
	"restaurant-system/shared/events"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Start(ctx context.Context, rabbitConn *amqp.Connection, dbPool *pgxpool.Pool) error {
	// === 1. Репозиторий и сервисы ===
	workerRepo := postgre.NewPostgresWorkerRepo(dbPool)
	workerSvc := app.NewWorkerService(workerRepo)

	// === 2. Создаём consumer/publisher на основе общего RabbitMQ клиента ===
	ch, err := rabbitConn.Channel()
	if err != nil {
		return err
	}

	consumer, err := rabbitmq.NewRabbitMQConsumer(rabbitConn)
	if err != nil {
		return err
	}
	publisher := rabbitmq.NewRabbitPublisher(ch, events.ExchangeOrders)

	// === 3. Создаём KitchenService ===
	kitchenSvc := app.NewKitchenService(workerSvc, consumer, publisher)

	// === 4. Запуск ===
	go func() {
		if err := kitchenSvc.Start(ctx); err != nil {
			log.Printf("❌ kitchen service stopped: %v", err)
		}
	}()

	log.Println("✅ Kitchen service started")
	return nil
}
