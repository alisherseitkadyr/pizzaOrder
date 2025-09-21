package app

import (
	"context"
	"fmt"
	domain "restaurant-system/services/kitchen-service/domain/models"
	"restaurant-system/services/kitchen-service/domain/ports"
	"restaurant-system/services/kitchen-service/utils/logger"
	"time"

	"github.com/jackc/pgx/v5"
)

type WorkerService struct {
	repo   ports.WorkerRepository
	Logger *logger.Logger
}

func NewWorkerService(repo ports.WorkerRepository, serviceName string) *WorkerService {
	return &WorkerService{
		repo:   repo,
		Logger: logger.New(serviceName),
	}
}

func (s *WorkerService) RegisterWorker(ctx context.Context, name, workerType string) error {
	worker := &domain.Worker{
		Name:            name,
		Type:            workerType,
		Status:          domain.WorkerOffline,
		OrdersProcessed: 0,
		LastSeen:        time.Now(),
		CreatedAt:       time.Now(),
	}
	if err := s.repo.Register(ctx, worker); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}
	return nil
}

// функция в WorkerService
// Добавляем методы для graceful shutdown
func (s *WorkerService) StartHeartbeat(ctx context.Context, workerName string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Heartbeat(ctx, workerName); err != nil {
				return
			}
		}
	}
}

func (s *WorkerService) SetWorkerOffline(ctx context.Context, workerName string) error {
	worker, err := s.repo.GetByName(ctx, workerName)
	if err != nil {
		return fmt.Errorf("worker not found: %w", err)
	}

	if err := worker.GoOffline(); err != nil {
		return fmt.Errorf("failed to set offline: %w", err)
	}

	return s.repo.Update(ctx, worker)
}

func (s *WorkerService) SetOfflineForAllWorkers(ctx context.Context) error {
	workers, err := s.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workers: %w", err)
	}

	for _, worker := range workers {
		if worker.Status == domain.WorkerOnline {
			worker.Status = domain.WorkerOffline
			worker.LastSeen = time.Now()
			if err := s.repo.Update(ctx, &worker); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *WorkerService) EnsureRegistered(ctx context.Context, name, workerType string) error {
	existing, err := s.repo.GetByName(ctx, name)
	if err != nil && err != pgx.ErrNoRows { // адаптировать под используемый драйвер
		return err
	}
	if existing != nil {
		if existing.Status == domain.WorkerOnline {
			return fmt.Errorf("worker already online")
		}
		// обновляем запись
		existing.Type = workerType
		existing.LastSeen = time.Now()
		existing.Status = domain.WorkerOnline
		return s.repo.Update(ctx, existing)
	}
	// создаём новую запись
	err = s.RegisterWorker(ctx, name, workerType)
	return err
}

func (s *WorkerService) Heartbeat(ctx context.Context, workerName string) error {
	worker, err := s.GetWorkerByName(ctx, workerName)
	if err != nil {
		s.Logger.Error("worker_not_found", "Failed to Heartbeat", "", err)
		return err
	}
	worker.Heartbeat()
	return s.repo.Update(ctx, worker)
}

func (s *WorkerService) AddProcessedOrder(ctx context.Context, workerName string) error {
	worker, err := s.GetWorkerByName(ctx, workerName)
	if err != nil {
		s.Logger.Error("worker_not_found", "Failed to Add Prosses Order", "", err)
		return err
	}
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
