package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/deals/internal/domain"
)

var ErrNotFound = errors.New("deal not found")

// UpdateDealInput carries optional fields for Update (keeps DealAPI parameter count within limits).
type UpdateDealInput struct {
	CustomerID *string
	VehicleID  *string
	Amount     *string
	Stage      *string
	AssignedTo *string
	Notes      *string
}

type dealRepository interface {
	Create(ctx context.Context, d *domain.Deal) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Deal, error)
	List(ctx context.Context, limit, offset int32, stageFilter, customerID string) ([]*domain.Deal, int32, error)
	Update(ctx context.Context, d *domain.Deal) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// DealAPI — HTTP/gRPC и тесты.
type DealAPI interface {
	Create(ctx context.Context, customerID, vehicleID, amount, stage, assignedTo, notes string) (*domain.Deal, error)
	Get(ctx context.Context, id string) (*domain.Deal, error)
	List(ctx context.Context, limit, offset int32, stageFilter, customerID string) ([]*domain.Deal, int32, error)
	Update(ctx context.Context, id string, in UpdateDealInput) (*domain.Deal, error)
	Delete(ctx context.Context, id string) error
}

type DealService struct {
	repo dealRepository
}

func NewDealService(repo dealRepository) *DealService {
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

func (s *DealService) Update(ctx context.Context, id string, in UpdateDealInput) (*domain.Deal, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrNotFound
	}
	if in.CustomerID != nil {
		if cid, err := uuid.Parse(*in.CustomerID); err == nil {
			existing.CustomerID = cid
		}
	}
	if in.VehicleID != nil {
		if vid, err := uuid.Parse(*in.VehicleID); err == nil {
			existing.VehicleID = vid
		}
	}
	if in.Amount != nil {
		existing.Amount = *in.Amount
	}
	if in.Stage != nil {
		existing.Stage = *in.Stage
	}
	if in.AssignedTo != nil {
		if *in.AssignedTo == "" {
			existing.AssignedTo = nil
		} else if a, err := uuid.Parse(*in.AssignedTo); err == nil {
			existing.AssignedTo = &a
		}
	}
	if in.Notes != nil {
		existing.Notes = *in.Notes
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
