package grpc

import (
	"context"

	"github.com/dealer/dealer/client-statistics-service/internal/domain"
	"github.com/dealer/dealer/client-statistics-service/internal/service"
	clientstatsv1 "github.com/dealer/dealer/pkg/pb/statistics/client/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	clientstatsv1.UnimplementedClientStatisticsServiceServer
	svc service.StatsAPI
}

func NewServer(svc service.StatsAPI) *Server {
	return &Server{svc: svc}
}

func (s *Server) GetOverview(ctx context.Context, _ *clientstatsv1.GetOverviewRequest) (*clientstatsv1.GetOverviewResponse, error) {
	overview, err := s.svc.GetOverview(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &clientstatsv1.GetOverviewResponse{Overview: toProto(overview)}, nil
}

func toProto(o *domain.Overview) *clientstatsv1.ClientOverview {
	if o == nil {
		return nil
	}
	out := &clientstatsv1.ClientOverview{
		ClientsCount:         o.ClientsCount,
		ClientVehiclesCount:  o.ClientVehiclesCount,
		RegisteredUsersCount: o.RegisteredUsersCount,
		ReviewsCount:         o.ReviewsCount,
		AverageRating:        o.AverageRating,
	}
	for _, item := range o.ReviewsByStatus {
		out.ReviewsByStatus = append(out.ReviewsByStatus, &clientstatsv1.ReviewStatusCount{
			Status: item.Status,
			Count:  item.Count,
		})
	}
	return out
}
