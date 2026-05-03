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

// CreateDealInput is the payload for Create (keeps DealAPI arity within Sonar limits).
type CreateDealInput struct {
	CustomerID, VehicleID, Amount, Stage, AssignedTo, Notes string
}

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
	Create(ctx context.Context, in CreateDealInput) (*domain.Deal, error)
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

func (s *DealService) Create(ctx context.Context, in CreateDealInput) (*domain.Deal, error) {
	stage := in.Stage
	if stage == "" {
		stage = "draft"
	}
	cid, err := uuid.Parse(in.CustomerID)
	if err != nil {
		return nil, errors.New("invalid customer_id")
	}
	vid, err := uuid.Parse(in.VehicleID)
	if err != nil {
		return nil, errors.New("invalid vehicle_id")
	}
	var assigned *uuid.UUID
	if in.AssignedTo != "" {
		if a, err := uuid.Parse(in.AssignedTo); err == nil {
			assigned = &a
		}
	}
	now := time.Now().UTC()
	d := &domain.Deal{
		ID:         uuid.New(),
		CustomerID: cid,
		VehicleID:  vid,
		Amount:     in.Amount,
		Stage:      stage,
		AssignedTo: assigned,
		Notes:      in.Notes,
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

func applyDealCustomerIDIfValid(d *domain.Deal, s *string) {
	if s == nil {
		return
	}
	cid, err := uuid.Parse(*s)
	if err != nil {
		return
	}
	d.CustomerID = cid
}

func applyDealVehicleIDIfValid(d *domain.Deal, s *string) {
	if s == nil {
		return
	}
	vid, err := uuid.Parse(*s)
	if err != nil {
		return
	}
	d.VehicleID = vid
}

func applyDealAssignedTo(d *domain.Deal, s *string) {
	if s == nil {
		return
	}
	if *s == "" {
		d.AssignedTo = nil
		return
	}
	a, err := uuid.Parse(*s)
	if err != nil {
		return
	}
	d.AssignedTo = &a
}

func mergeDealUpdateInput(d *domain.Deal, in UpdateDealInput) {
	applyDealCustomerIDIfValid(d, in.CustomerID)
	applyDealVehicleIDIfValid(d, in.VehicleID)
	if in.Amount != nil {
		d.Amount = *in.Amount
	}
	if in.Stage != nil {
		d.Stage = *in.Stage
	}
	applyDealAssignedTo(d, in.AssignedTo)
	if in.Notes != nil {
		d.Notes = *in.Notes
	}
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
	mergeDealUpdateInput(existing, in)
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
