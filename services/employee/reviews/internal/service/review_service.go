package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/employee-reviews-service/internal/domain"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
)

var (
	ErrInvalidID       = errors.New("invalid id")
	ErrNotFound        = errors.New("review not found")
	ErrVehicleNotFound = errors.New("vehicle not found")
)

type reviewRepository interface {
	Insert(ctx context.Context, review *domain.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	List(ctx context.Context, p domain.ReviewListParams) ([]*domain.Review, int64, error)
	Stats(ctx context.Context, clientID, dealerPointID *uuid.UUID) (*domain.ReviewStats, error)
}

type vehicleLookup interface {
	GetByID(ctx context.Context, id string) (*vehiclesv1.Vehicle, error)
}

type ReviewAPI interface {
	RecordReviewPublished(ctx context.Context, reviewID, clientID, userID, dealerPointID, vehicleID uuid.UUID, clientEmail, clientFullName, text, status string, rating int32, occurredAt time.Time) error
	Get(ctx context.Context, id string) (*domain.Review, error)
	ListByClient(ctx context.Context, clientID string, limit, offset int32) ([]*domain.Review, int64, error)
	List(ctx context.Context, p domain.ReviewListParams) ([]*domain.Review, int64, error)
	Stats(ctx context.Context, clientID, dealerPointID string) (*domain.ReviewStats, error)
}

type ReviewService struct {
	repo     reviewRepository
	vehicles vehicleLookup
}

func NewReviewService(repo reviewRepository, vehicles vehicleLookup) *ReviewService {
	return &ReviewService{repo: repo, vehicles: vehicles}
}

func (s *ReviewService) RecordReviewPublished(
	ctx context.Context,
	reviewID, clientID, userID, dealerPointID, vehicleID uuid.UUID,
	clientEmail, clientFullName, text, status string,
	rating int32,
	occurredAt time.Time,
) error {
	if s.vehicles == nil {
		return ErrVehicleNotFound
	}
	veh, err := s.vehicles.GetByID(ctx, vehicleID.String())
	if err != nil {
		return ErrVehicleNotFound
	}
	if veh == nil || veh.Id != vehicleID.String() {
		return ErrVehicleNotFound
	}

	now := time.Now().UTC()
	if occurredAt.IsZero() {
		occurredAt = now
	}
	return s.repo.Insert(ctx, &domain.Review{
		ReviewID:       reviewID,
		ClientID:       clientID,
		UserID:         userID,
		ClientEmail:    clientEmail,
		ClientFullName: clientFullName,
		DealerPointID:  dealerPointID,
		VehicleID:      vehicleID,
		VehicleVIN:     veh.Vin,
		VehicleMake:    veh.Make,
		VehicleModel:   veh.Model,
		VehicleYear:    veh.Year,
		Rating:         rating,
		Text:           text,
		Status:         status,
		OccurredAt:     occurredAt,
		CreatedAt:      now,
	})
}

func (s *ReviewService) Get(ctx context.Context, id string) (*domain.Review, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	review, err := s.repo.GetByID(ctx, parsed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return review, nil
}

func (s *ReviewService) ListByClient(ctx context.Context, clientID string, limit, offset int32) ([]*domain.Review, int64, error) {
	id, err := uuid.Parse(clientID)
	if err != nil {
		return nil, 0, ErrInvalidID
	}
	return s.repo.List(ctx, domain.ReviewListParams{ClientID: &id, Limit: limit, Offset: offset})
}

func (s *ReviewService) List(ctx context.Context, p domain.ReviewListParams) ([]*domain.Review, int64, error) {
	if p.ClientID != nil && *p.ClientID == uuid.Nil {
		return nil, 0, ErrInvalidID
	}
	if p.DealerPointID != nil && *p.DealerPointID == uuid.Nil {
		return nil, 0, ErrInvalidID
	}
	return s.repo.List(ctx, p)
}

func (s *ReviewService) Stats(ctx context.Context, clientID, dealerPointID string) (*domain.ReviewStats, error) {
	var cid, did *uuid.UUID
	if clientID != "" {
		id, err := uuid.Parse(clientID)
		if err != nil {
			return nil, ErrInvalidID
		}
		cid = &id
	}
	if dealerPointID != "" {
		id, err := uuid.Parse(dealerPointID)
		if err != nil {
			return nil, ErrInvalidID
		}
		did = &id
	}
	return s.repo.Stats(ctx, cid, did)
}

func ParseOptionalUUID(raw string) (*uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, ErrInvalidID
	}
	return &id, nil
}
