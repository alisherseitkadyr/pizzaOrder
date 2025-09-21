package ports

import (
	"context"
	domain "restaurant-system/services/kitchen-service/domain/models"
)

type WorkerRepository interface {
	Register(ctx context.Context, worker *domain.Worker) error
	Update(ctx context.Context, worker *domain.Worker) error
	GetAll(ctx context.Context) ([]domain.Worker, error)
	GetByName(ctx context.Context, name string) (*domain.Worker, error)
}
