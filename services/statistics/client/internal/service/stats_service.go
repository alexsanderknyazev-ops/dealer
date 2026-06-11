package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/client-statistics-service/internal/domain"
)

type statsRepository interface {
	InsertClientRegistration(ctx context.Context, userID uuid.UUID, email string, vehicleID *uuid.UUID, occurredAt time.Time) error
	InsertReviewEvent(ctx context.Context, reviewID, clientID, userID, dealerPointID, vehicleID uuid.UUID, rating int32, status string, occurredAt time.Time) error
	GetOverview(ctx context.Context) (*domain.Overview, error)
}

type StatsAPI interface {
	RecordClientRegistration(ctx context.Context, userID uuid.UUID, email string, vehicleID *uuid.UUID, occurredAt time.Time) error
	RecordReviewPublished(ctx context.Context, reviewID, clientID, userID, dealerPointID, vehicleID uuid.UUID, rating int32, status string, occurredAt time.Time) error
	GetOverview(ctx context.Context) (*domain.Overview, error)
}

type StatsService struct {
	repo statsRepository
}

func NewStatsService(repo statsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) RecordClientRegistration(ctx context.Context, userID uuid.UUID, email string, vehicleID *uuid.UUID, occurredAt time.Time) error {
	return s.repo.InsertClientRegistration(ctx, userID, email, vehicleID, occurredAt)
}

func (s *StatsService) RecordReviewPublished(ctx context.Context, reviewID, clientID, userID, dealerPointID, vehicleID uuid.UUID, rating int32, status string, occurredAt time.Time) error {
	return s.repo.InsertReviewEvent(ctx, reviewID, clientID, userID, dealerPointID, vehicleID, rating, status, occurredAt)
}

func (s *StatsService) GetOverview(ctx context.Context) (*domain.Overview, error) {
	return s.repo.GetOverview(ctx)
}
