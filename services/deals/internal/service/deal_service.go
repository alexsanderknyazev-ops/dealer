package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/deals/internal/domain"
	"github.com/dealer/dealer/services/deals/internal/repository"
)

var ErrNotFound = errors.New("deal not found")

type DealService struct {
	repo *repository.DealRepository
}

func NewDealService(repo *repository.DealRepository) *DealService {
	return &DealService{repo: repo}
}

func (s *DealService) Create(ctx context.Context, customerID, vehicleID, amount, stage, assignedTo, notes string) (*domain.Deal, error) {
	if stage == "" {
		stage = "draft"
	}
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return nil, errors.New("invalid customer_id")
	}
	vid, err := uuid.Parse(vehicleID)
	if err != nil {
		return nil, errors.New("invalid vehicle_id")
	}
	var assigned *uuid.UUID
	if assignedTo != "" {
		if a, err := uuid.Parse(assignedTo); err == nil {
			assigned = &a
		}
	}
	now := time.Now().UTC()
	d := &domain.Deal{
		ID:         uuid.New(),
		CustomerID: cid,
		VehicleID:  vid,
		Amount:     amount,
		Stage:      stage,
		AssignedTo: assigned,
		Notes:      notes,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DealService) Get(ctx context.Context, id string) (*domain.Deal, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	d, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return d, nil
}

func (s *DealService) List(ctx context.Context, limit, offset int32, stageFilter, customerID string) ([]*domain.Deal, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset, stageFilter, customerID)
}

func (s *DealService) Update(ctx context.Context, id string, customerID, vehicleID, amount, stage, assignedTo, notes *string) (*domain.Deal, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrNotFound
	}
	if customerID != nil {
		if cid, err := uuid.Parse(*customerID); err == nil {
			existing.CustomerID = cid
		}
	}
	if vehicleID != nil {
		if vid, err := uuid.Parse(*vehicleID); err == nil {
			existing.VehicleID = vid
		}
	}
	if amount != nil {
		existing.Amount = *amount
	}
	if stage != nil {
		existing.Stage = *stage
	}
	if assignedTo != nil {
		if *assignedTo == "" {
			existing.AssignedTo = nil
		} else if a, err := uuid.Parse(*assignedTo); err == nil {
			existing.AssignedTo = &a
		}
	}
	if notes != nil {
		existing.Notes = *notes
	}
	existing.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DealService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, uid)
}
