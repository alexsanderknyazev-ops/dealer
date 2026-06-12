package service

import (
	"context"
)

type invitationRepository interface {
	CreateFromClosedWorkOrders(ctx context.Context, limit int) (int64, error)
	CreateFromCompletedDeals(ctx context.Context, limit int) (int64, error)
}

type RunResult struct {
	WorkOrdersCreated int64
	DealsCreated      int64
}

type SchedulerService struct {
	repo      invitationRepository
	batchSize int
}

func NewSchedulerService(repo invitationRepository, batchSize int) *SchedulerService {
	return &SchedulerService{repo: repo, batchSize: batchSize}
}

func (s *SchedulerService) RunOnce(ctx context.Context) (RunResult, error) {
	wo, err := s.repo.CreateFromClosedWorkOrders(ctx, s.batchSize)
	if err != nil {
		return RunResult{}, err
	}
	deals, err := s.repo.CreateFromCompletedDeals(ctx, s.batchSize)
	if err != nil {
		return RunResult{}, err
	}
	return RunResult{WorkOrdersCreated: wo, DealsCreated: deals}, nil
}
