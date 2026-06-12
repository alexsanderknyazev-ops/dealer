package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dealer/dealer/client-reviews-service/internal/domain"
	"github.com/dealer/dealer/client-reviews-service/internal/repository"
	"github.com/dealer/dealer/pkg/obslog"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
)

var (
	ErrNotOwner              = repository.ErrNotOwner
	ErrReviewNotFound        = errors.New("review not found")
	ErrInvitationNotFound    = errors.New("review invitation not found")
	ErrDuplicateReview       = errors.New("review already exists for this vehicle")
	ErrInvalidRating         = errors.New("rating must be between 1 and 5")
	ErrMissingDealerPoint    = errors.New("vehicle has no dealer point")
)

type reviewRepository interface {
	ClientVehicle(ctx context.Context, userID, vehicleID uuid.UUID) (uuid.UUID, error)
	ClientProfile(ctx context.Context, clientID uuid.UUID) (email, fullName string, err error)
	Create(ctx context.Context, review *domain.Review) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Review, error)
	GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*domain.Review, error)
	ListPendingInvitationsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.ReviewInvitation, error)
	DismissInvitationForUser(ctx context.Context, id, userID uuid.UUID) error
	CompleteInvitationsForVehicle(ctx context.Context, clientID, vehicleID uuid.UUID) error
}

type vehicleLookup interface {
	GetByID(ctx context.Context, id string) (*vehiclesv1.Vehicle, error)
}

type ReviewAPI interface {
	CreateReview(ctx context.Context, userID uuid.UUID, vehicleID string, rating int32, text string) (*domain.Review, error)
	ListMyReviews(ctx context.Context, userID uuid.UUID) ([]*domain.Review, error)
	GetReview(ctx context.Context, userID uuid.UUID, id string) (*domain.Review, error)
	ListReviewInvitations(ctx context.Context, userID uuid.UUID) ([]*domain.ReviewInvitation, error)
	DismissReviewInvitation(ctx context.Context, userID uuid.UUID, id string) error
}

type reviewPublisher interface {
	Publish(ctx context.Context, review *domain.Review, clientEmail, clientFullName string) error
}

type ReviewService struct {
	repo      reviewRepository
	vehicles  vehicleLookup
	publisher reviewPublisher
}

func NewReviewService(repo reviewRepository, vehicles vehicleLookup, publisher reviewPublisher) *ReviewService {
	return &ReviewService{repo: repo, vehicles: vehicles, publisher: publisher}
}

func (s *ReviewService) CreateReview(ctx context.Context, userID uuid.UUID, vehicleID string, rating int32, text string) (*domain.Review, error) {
	if rating < 1 || rating > 5 {
		return nil, ErrInvalidRating
	}
	vid, err := uuid.Parse(strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, err
	}

	clientID, err := s.repo.ClientVehicle(ctx, userID, vid)
	if err != nil {
		return nil, err
	}

	veh, err := s.vehicles.GetByID(ctx, vid.String())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotOwner
		}
		return nil, err
	}
	if veh.DealerPointId == "" {
		return nil, ErrMissingDealerPoint
	}
	dealerPointID, err := uuid.Parse(veh.DealerPointId)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	review := &domain.Review{
		ID:            uuid.New(),
		ClientID:      clientID,
		UserID:        userID,
		DealerPointID: dealerPointID,
		VehicleID:     vid,
		Rating:        rating,
		Text:          strings.TrimSpace(text),
		Status:        "published",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.Create(ctx, review); err != nil {
		if repository.IsDuplicateReview(err) {
			return nil, ErrDuplicateReview
		}
		return nil, err
	}
	if s.publisher != nil {
		email, fullName, _ := s.repo.ClientProfile(ctx, clientID)
		if err := s.publisher.Publish(ctx, review, email, fullName); err != nil {
			obslog.Default.Warn("kafka publish failed", "event", "review.published", "review_id", review.ID.String(), "err", err)
		}
	}
	_ = s.repo.CompleteInvitationsForVehicle(ctx, clientID, vid)
	return review, nil
}

func (s *ReviewService) ListMyReviews(ctx context.Context, userID uuid.UUID) ([]*domain.Review, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *ReviewService) GetReview(ctx context.Context, userID uuid.UUID, id string) (*domain.Review, error) {
	reviewID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	review, err := s.repo.GetByIDForUser(ctx, reviewID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrReviewNotFound
	}
	return review, err
}

func (s *ReviewService) ListReviewInvitations(ctx context.Context, userID uuid.UUID) ([]*domain.ReviewInvitation, error) {
	return s.repo.ListPendingInvitationsByUserID(ctx, userID)
}

func (s *ReviewService) DismissReviewInvitation(ctx context.Context, userID uuid.UUID, id string) error {
	invID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return err
	}
	err = s.repo.DismissInvitationForUser(ctx, invID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvitationNotFound
	}
	return err
}
