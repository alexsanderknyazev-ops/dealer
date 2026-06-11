package grpc

import (
	"context"
	"errors"

	"github.com/dealer/dealer/employee-reviews-service/internal/domain"
	"github.com/dealer/dealer/employee-reviews-service/internal/service"
	reviewsv1 "github.com/dealer/dealer/pkg/pb/reviews/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	reviewsv1.UnimplementedEmployeeReviewsServiceServer
	svc service.ReviewAPI
}

func NewServer(svc service.ReviewAPI) *Server {
	return &Server{svc: svc}
}

func (s *Server) ListReviewsByClient(ctx context.Context, req *reviewsv1.ListReviewsByClientRequest) (*reviewsv1.ListReviewsByClientResponse, error) {
	list, total, err := s.svc.ListByClient(ctx, req.ClientId, req.Limit, req.Offset)
	if err != nil {
		return nil, mapErr(err)
	}
	return &reviewsv1.ListReviewsByClientResponse{Reviews: toProtoList(list), Total: total}, nil
}

func (s *Server) ListReviews(ctx context.Context, req *reviewsv1.ListReviewsRequest) (*reviewsv1.ListReviewsResponse, error) {
	clientID, err := service.ParseOptionalUUID(req.ClientId)
	if err != nil {
		return nil, mapErr(err)
	}
	dealerPointID, err := service.ParseOptionalUUID(req.DealerPointId)
	if err != nil {
		return nil, mapErr(err)
	}
	list, total, err := s.svc.List(ctx, domain.ReviewListParams{
		Limit:         req.Limit,
		Offset:        req.Offset,
		ClientID:      clientID,
		DealerPointID: dealerPointID,
		Status:        req.Status,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &reviewsv1.ListReviewsResponse{Reviews: toProtoList(list), Total: total}, nil
}

func (s *Server) GetReviewStats(ctx context.Context, req *reviewsv1.GetReviewStatsRequest) (*reviewsv1.GetReviewStatsResponse, error) {
	stats, err := s.svc.Stats(ctx, req.ClientId, req.DealerPointId)
	if err != nil {
		return nil, mapErr(err)
	}
	out := &reviewsv1.GetReviewStatsResponse{
		TotalCount:    stats.TotalCount,
		AverageRating: stats.AverageRating,
	}
	for _, sc := range stats.ByStatus {
		out.ByStatus = append(out.ByStatus, &reviewsv1.ReviewStatusCount{Status: sc.Status, Count: sc.Count})
	}
	return out, nil
}

func mapErr(err error) error {
	if errors.Is(err, service.ErrInvalidID) {
		return status.Error(codes.InvalidArgument, "invalid id")
	}
	return status.Error(codes.Internal, err.Error())
}

func toProtoList(list []*domain.Review) []*reviewsv1.EmployeeReview {
	out := make([]*reviewsv1.EmployeeReview, len(list))
	for i, r := range list {
		out[i] = toProto(r)
	}
	return out
}

func toProto(r *domain.Review) *reviewsv1.EmployeeReview {
	if r == nil {
		return nil
	}
	return &reviewsv1.EmployeeReview{
		Id:             r.ID.String(),
		ReviewId:       r.ReviewID.String(),
		ClientId:       r.ClientID.String(),
		UserId:         r.UserID.String(),
		ClientEmail:    r.ClientEmail,
		ClientFullName: r.ClientFullName,
		DealerPointId:  r.DealerPointID.String(),
		VehicleId:      r.VehicleID.String(),
		VehicleVin:     r.VehicleVIN,
		VehicleMake:    r.VehicleMake,
		VehicleModel:   r.VehicleModel,
		VehicleYear:    r.VehicleYear,
		Rating:         r.Rating,
		Text:           r.Text,
		Status:         r.Status,
		OccurredAt:     r.OccurredAt.Unix(),
		CreatedAt:      r.CreatedAt.Unix(),
	}
}
