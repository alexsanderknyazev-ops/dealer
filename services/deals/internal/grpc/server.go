package grpc

import (
	"context"
	"errors"

	dealsv1 "github.com/dealer/dealer/pkg/pb/deals/v1"
	"github.com/dealer/dealer/services/deals/internal/domain"
	"github.com/dealer/dealer/services/deals/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	dealsv1.UnimplementedDealsServiceServer
	svc service.DealAPI
}

func NewServer(svc service.DealAPI) *Server {
	return &Server{svc: svc}
}

func toProto(d *domain.Deal) *dealsv1.Deal {
	if d == nil {
		return nil
	}
	assignedTo := ""
	if d.AssignedTo != nil {
		assignedTo = d.AssignedTo.String()
	}
	return &dealsv1.Deal{
		Id:         d.ID.String(),
		CustomerId: d.CustomerID.String(),
		VehicleId:  d.VehicleID.String(),
		Amount:     d.Amount,
		Stage:      d.Stage,
		AssignedTo: assignedTo,
		Notes:      d.Notes,
		CreatedAt:  d.CreatedAt.Unix(),
		UpdatedAt:  d.UpdatedAt.Unix(),
	}
}

func (s *Server) CreateDeal(ctx context.Context, req *dealsv1.CreateDealRequest) (*dealsv1.CreateDealResponse, error) {
	d, err := s.svc.Create(ctx, req.CustomerId, req.VehicleId, req.Amount, req.Stage, req.AssignedTo, req.Notes)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &dealsv1.CreateDealResponse{Deal: toProto(d)}, nil
}

func (s *Server) GetDeal(ctx context.Context, req *dealsv1.GetDealRequest) (*dealsv1.GetDealResponse, error) {
	d, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "deal not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &dealsv1.GetDealResponse{Deal: toProto(d)}, nil
}

func (s *Server) ListDeals(ctx context.Context, req *dealsv1.ListDealsRequest) (*dealsv1.ListDealsResponse, error) {
	list, total, err := s.svc.List(ctx, req.Limit, req.Offset, req.Stage, req.CustomerId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*dealsv1.Deal, len(list))
	for i, d := range list {
		out[i] = toProto(d)
	}
	return &dealsv1.ListDealsResponse{Deals: out, Total: total}, nil
}

func (s *Server) UpdateDeal(ctx context.Context, req *dealsv1.UpdateDealRequest) (*dealsv1.UpdateDealResponse, error) {
	d, err := s.svc.Update(ctx, req.Id, service.UpdateDealInput{
		CustomerID: req.CustomerId, VehicleID: req.VehicleId, Amount: req.Amount,
		Stage: req.Stage, AssignedTo: req.AssignedTo, Notes: req.Notes,
	})
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "deal not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &dealsv1.UpdateDealResponse{Deal: toProto(d)}, nil
}

func (s *Server) DeleteDeal(ctx context.Context, req *dealsv1.DeleteDealRequest) (*dealsv1.DeleteDealResponse, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "deal not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &dealsv1.DeleteDealResponse{}, nil
}
