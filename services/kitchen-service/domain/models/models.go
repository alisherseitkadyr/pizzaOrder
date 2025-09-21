package domain

import (
	"errors"
	"time"
)

var (
	ErrWorkerAlreadyOnline  = errors.New("worker already online")
	ErrWorkerAlreadyOffline = errors.New("worker already offline")
)

type WorkerStatus string

const (
	WorkerOnline  WorkerStatus = "online"
	WorkerOffline WorkerStatus = "offline"
)

type Worker struct {
	ID              int64
	Name            string
	Type            string // dine_in, takeout, delivery
	Status          WorkerStatus
	OrdersProcessed int64
	LastSeen        time.Time
	CreatedAt       time.Time
}

// GoOnline переводит работника в статус online
func (w *Worker) GoOnline() error {
	if w.Status == WorkerOnline {
		return ErrWorkerAlreadyOnline
	}
	w.Status = WorkerOnline
	w.LastSeen = time.Now()
	return nil
}

// GoOffline переводит работника в статус offline
func (w *Worker) GoOffline() error {
	if w.Status == WorkerOffline {
		return ErrWorkerAlreadyOffline
	}
	w.Status = WorkerOffline
	w.LastSeen = time.Now()
	return nil
}

// Heartbeat обновляет время последней активности
func (w *Worker) Heartbeat() {
	w.LastSeen = time.Now()
}

// ProcessOrder увеличивает счётчик заказов
func (w *Worker) ProcessOrder() {
	w.OrdersProcessed++
	w.LastSeen = time.Now()
}
