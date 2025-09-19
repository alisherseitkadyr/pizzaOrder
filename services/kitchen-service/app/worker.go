package app

import (
	"context"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/ports"
	"time"
)

type WorkerService struct {
	repo ports.WorkerRepository
}

func NewWorkerService(repo ports.WorkerRepository) *WorkerService {
	return &WorkerService{repo: repo}
}

func (s *WorkerService) RegisterWorker(ctx context.Context, name, workerType string) (*domain.Worker, error) {
	worker := &domain.Worker{
		Name:            name,
		Type:            workerType,
		Status:          domain.WorkerOffline,
		OrdersProcessed: 0,
		LastSeen:        time.Now(),
		CreatedAt:       time.Now(),
	}
	if err := s.repo.Register(ctx, worker); err != nil {
		return nil, fmt.Errorf("failed to register worker: %w", err)
	}
	return worker, nil
}

func (s *WorkerService) Heartbeat(ctx context.Context, worker *domain.Worker) error {
	worker.Heartbeat()
	return s.repo.Update(ctx, worker)
}

func (s *WorkerService) AddProcessedOrder(ctx context.Context, worker *domain.Worker) error {
	worker.ProcessOrder()
	return s.repo.Update(ctx, worker)
}

func (s *WorkerService) GetAllWorkers(ctx context.Context) ([]domain.Worker, error) {
	return s.repo.GetAll(ctx)
}

func (s *WorkerService) GetWorkerByName(ctx context.Context, name string) (*domain.Worker, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *WorkerService) SetOnline(ctx context.Context, worker *domain.Worker) error {
	if err := worker.GoOnline(); err != nil {
		return err
	}
	return s.repo.Update(ctx, worker)
}

func (s *WorkerService) SetOffline(ctx context.Context, worker *domain.Worker) error {
	if err := worker.GoOffline(); err != nil {
		return err
	}
	return s.repo.Update(ctx, worker)
}

func (s *WorkerService) updateLastSeenForOnlineWorkers(ctx context.Context) {
	workers, _ := s.repo.GetAll(ctx)
	now := time.Now()
	for _, worker := range workers {
		if worker.Status == domain.WorkerOnline {
			worker.LastSeen = now
			s.repo.Update(ctx, &worker)
		}
	}
}

func (s *WorkerService) GetAvailableWorker(ctx context.Context) (domain.Worker, error) {
	workers, err := s.GetAllWorkers(ctx)
	if err != nil {
		return domain.Worker{}, err
	}
	for _, worker := range workers {
		if worker.Status == domain.WorkerOnline {
			fmt.Println("found aviable worker", worker.Name)
			return worker, nil
		}
	}
	return domain.Worker{}, fmt.Errorf("not found aviable worker")
}
