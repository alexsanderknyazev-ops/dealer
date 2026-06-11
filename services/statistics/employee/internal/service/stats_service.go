package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/employee-statistics-service/internal/domain"
)

type statsRepository interface {
	InsertDealEvent(ctx context.Context, dealID, customerID, vehicleID uuid.UUID, amount, stage string, occurredAt time.Time) error
	GetOverview(ctx context.Context) (*domain.Overview, error)
}

type StatsAPI interface {
	RecordDealCompleted(ctx context.Context, dealID, customerID, vehicleID uuid.UUID, amount, stage string, occurredAt time.Time) error
	GetOverview(ctx context.Context) (*domain.Overview, error)
}

type StatsService struct {
	repo statsRepository
}

func NewStatsService(repo statsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) RecordDealCompleted(ctx context.Context, dealID, customerID, vehicleID uuid.UUID, amount, stage string, occurredAt time.Time) error {
	return s.repo.InsertDealEvent(ctx, dealID, customerID, vehicleID, amount, stage, occurredAt)
}

func (s *StatsService) GetOverview(ctx context.Context) (*domain.Overview, error) {
	return s.repo.GetOverview(ctx)
}
