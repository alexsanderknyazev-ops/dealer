package grpc

import (
	"context"
	"errors"

	"github.com/dealer/dealer/client-reviews-service/internal/domain"
	"github.com/dealer/dealer/client-reviews-service/internal/service"
	"github.com/dealer/dealer/pkg/clientjwt"
	reviewsv1 "github.com/dealer/dealer/pkg/pb/reviews/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	reviewsv1.UnimplementedReviewsServiceServer
	svc       service.ReviewAPI
	jwtSecret string
}

func NewServer(svc service.ReviewAPI, jwtSecret string) *Server {
	return &Server{svc: svc, jwtSecret: jwtSecret}
}

func (s *Server) CreateReview(ctx context.Context, req *reviewsv1.CreateReviewRequest) (*reviewsv1.CreateReviewResponse, error) {
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	if req.VehicleId == "" {
		return nil, status.Error(codes.InvalidArgument, "vehicle_id required")
	}
	review, err := s.svc.CreateReview(ctx, userID, req.VehicleId, req.Rating, req.Text)
	if err != nil {
		return nil, mapErr(err)
	}
	return &reviewsv1.CreateReviewResponse{Review: toProto(review)}, nil
}

func (s *Server) ListMyReviews(ctx context.Context, _ *reviewsv1.ListMyReviewsRequest) (*reviewsv1.ListMyReviewsResponse, error) {
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	list, err := s.svc.ListMyReviews(ctx, userID)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*reviewsv1.Review, len(list))
	for i, r := range list {
		out[i] = toProto(r)
	}
	return &reviewsv1.ListMyReviewsResponse{Reviews: out}, nil
}

func (s *Server) ListReviewInvitations(ctx context.Context, _ *reviewsv1.ListReviewInvitationsRequest) (*reviewsv1.ListReviewInvitationsResponse, error) {
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	list, err := s.svc.ListReviewInvitations(ctx, userID)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*reviewsv1.ReviewInvitation, len(list))
	for i, inv := range list {
		out[i] = toInvitationProto(inv)
	}
	return &reviewsv1.ListReviewInvitationsResponse{Invitations: out}, nil
}

func (s *Server) DismissReviewInvitation(ctx context.Context, req *reviewsv1.DismissReviewInvitationRequest) (*reviewsv1.DismissReviewInvitationResponse, error) {
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if err := s.svc.DismissReviewInvitation(ctx, userID, req.Id); err != nil {
		return nil, mapErr(err)
	}
	return &reviewsv1.DismissReviewInvitationResponse{}, nil
}

func (s *Server) GetReview(ctx context.Context, req *reviewsv1.GetReviewRequest) (*reviewsv1.GetReviewResponse, error) {
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	review, err := s.svc.GetReview(ctx, userID, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &reviewsv1.GetReviewResponse{Review: toProto(review)}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrNotOwner):
		return status.Error(codes.PermissionDenied, "vehicle not linked to your account")
	case errors.Is(err, service.ErrReviewNotFound), errors.Is(err, service.ErrInvitationNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrDuplicateReview):
		return status.Error(codes.AlreadyExists, "review already exists for this vehicle")
	case errors.Is(err, service.ErrInvalidRating):
		return status.Error(codes.InvalidArgument, "rating must be between 1 and 5")
	case errors.Is(err, service.ErrMissingDealerPoint):
		return status.Error(codes.FailedPrecondition, "vehicle has no dealer point")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toInvitationProto(inv *domain.ReviewInvitation) *reviewsv1.ReviewInvitation {
	if inv == nil {
		return nil
	}
	return &reviewsv1.ReviewInvitation{
		Id:            inv.ID.String(),
		ClientId:      inv.ClientID.String(),
		VehicleId:     inv.VehicleID.String(),
		DealerPointId: inv.DealerPointID.String(),
		SourceType:    inv.SourceType,
		SourceId:      inv.SourceID.String(),
		ServiceKind:   inv.ServiceKind,
		Status:        inv.Status,
		CreatedAt:     inv.CreatedAt.Unix(),
	}
}

func toProto(r *domain.Review) *reviewsv1.Review {
	if r == nil {
		return nil
	}
	return &reviewsv1.Review{
		Id:            r.ID.String(),
		ClientId:      r.ClientID.String(),
		DealerPointId: r.DealerPointID.String(),
		VehicleId:     r.VehicleID.String(),
		Rating:        r.Rating,
		Text:          r.Text,
		Status:        r.Status,
		CreatedAt:     r.CreatedAt.Unix(),
		UpdatedAt:     r.UpdatedAt.Unix(),
	}
}
