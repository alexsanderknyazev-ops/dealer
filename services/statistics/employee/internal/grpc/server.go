package grpc

import (
	"context"

	"github.com/dealer/dealer/employee-statistics-service/internal/domain"
	"github.com/dealer/dealer/employee-statistics-service/internal/service"
	employeestatsv1 "github.com/dealer/dealer/pkg/pb/statistics/employee/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	employeestatsv1.UnimplementedEmployeeStatisticsServiceServer
	svc service.StatsAPI
}

func NewServer(svc service.StatsAPI) *Server {
	return &Server{svc: svc}
}

func (s *Server) GetOverview(ctx context.Context, _ *employeestatsv1.GetOverviewRequest) (*employeestatsv1.GetOverviewResponse, error) {
	overview, err := s.svc.GetOverview(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &employeestatsv1.GetOverviewResponse{Overview: toProto(overview)}, nil
}

func toProto(o *domain.Overview) *employeestatsv1.EmployeeOverview {
	if o == nil {
		return nil
	}
	out := &employeestatsv1.EmployeeOverview{
		DealsCount:   o.DealsCount,
		TotalRevenue: o.TotalRevenue,
	}
	for _, item := range o.DealsByStage {
		out.DealsByStage = append(out.DealsByStage, &employeestatsv1.DealStageCount{
			Stage: item.Stage,
			Count: item.Count,
		})
	}
	return out
}
