package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
	"github.com/dealer/dealer/services/vehicles/internal/domain"
	"github.com/dealer/dealer/services/vehicles/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	vehiclesv1.UnimplementedVehiclesServiceServer
	svc *service.VehicleService
}

func NewServer(svc *service.VehicleService) *Server {
	return &Server{svc: svc}
}

func toProto(v *domain.Vehicle) *vehiclesv1.Vehicle {
	if v == nil {
		return nil
	}
	out := &vehiclesv1.Vehicle{
		Id:        v.ID.String(),
		Vin:       v.VIN,
		Make:      v.Make,
		Model:     v.Model,
		Year:      v.Year,
		MileageKm: v.MileageKm,
		Price:     v.Price,
		Status:    v.Status,
		Color:     v.Color,
		Notes:     v.Notes,
		CreatedAt: v.CreatedAt.Unix(),
		UpdatedAt: v.UpdatedAt.Unix(),
	}
	if v.BrandID != nil {
		out.BrandId = v.BrandID.String()
	}
	if v.DealerPointID != nil {
		out.DealerPointId = v.DealerPointID.String()
	}
	if v.LegalEntityID != nil {
		out.LegalEntityId = v.LegalEntityID.String()
	}
	if v.WarehouseID != nil {
		out.WarehouseId = v.WarehouseID.String()
	}
	return out
}

func parseUUIDOptional(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	uid, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &uid
}

func (s *Server) CreateVehicle(ctx context.Context, req *vehiclesv1.CreateVehicleRequest) (*vehiclesv1.CreateVehicleResponse, error) {
	v, err := s.svc.Create(ctx, req.Vin, req.Make, req.Model, req.Year, req.MileageKm, req.Price, req.Status, req.Color, req.Notes,
		parseUUIDOptional(req.BrandId), parseUUIDOptional(req.DealerPointId), parseUUIDOptional(req.LegalEntityId), parseUUIDOptional(req.WarehouseId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &vehiclesv1.CreateVehicleResponse{Vehicle: toProto(v)}, nil
}

func (s *Server) GetVehicle(ctx context.Context, req *vehiclesv1.GetVehicleRequest) (*vehiclesv1.GetVehicleResponse, error) {
	v, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "vehicle not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &vehiclesv1.GetVehicleResponse{Vehicle: toProto(v)}, nil
}

func (s *Server) ListVehicles(ctx context.Context, req *vehiclesv1.ListVehiclesRequest) (*vehiclesv1.ListVehiclesResponse, error) {
	list, total, err := s.svc.List(ctx, req.Limit, req.Offset, req.Search, req.Status,
		parseUUIDOptional(req.BrandId), parseUUIDOptional(req.DealerPointId), parseUUIDOptional(req.LegalEntityId), parseUUIDOptional(req.WarehouseId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*vehiclesv1.Vehicle, len(list))
	for i, v := range list {
		out[i] = toProto(v)
	}
	return &vehiclesv1.ListVehiclesResponse{Vehicles: out, Total: total}, nil
}

func (s *Server) UpdateVehicle(ctx context.Context, req *vehiclesv1.UpdateVehicleRequest) (*vehiclesv1.UpdateVehicleResponse, error) {
	var brandID, dealerPointID, legalEntityID, warehouseID *uuid.UUID
	clearBrand, clearDealerPoint, clearLegalEntity, clearWarehouse := false, false, false, false
	if req.BrandId != nil {
		if s := req.GetBrandId(); s == "" {
			clearBrand = true
		} else {
			brandID = parseUUIDOptional(s)
		}
	}
	if req.DealerPointId != nil {
		if s := req.GetDealerPointId(); s == "" {
			clearDealerPoint = true
		} else {
			dealerPointID = parseUUIDOptional(s)
		}
	}
	if req.LegalEntityId != nil {
		if s := req.GetLegalEntityId(); s == "" {
			clearLegalEntity = true
		} else {
			legalEntityID = parseUUIDOptional(s)
		}
	}
	if req.WarehouseId != nil {
		if s := req.GetWarehouseId(); s == "" {
			clearWarehouse = true
		} else {
			warehouseID = parseUUIDOptional(s)
		}
	}
	v, err := s.svc.Update(ctx, req.Id,
		req.Vin, req.Make, req.Model, req.Year, req.MileageKm,
		req.Price, req.Status, req.Color, req.Notes, brandID, clearBrand,
		dealerPointID, legalEntityID, warehouseID, clearDealerPoint, clearLegalEntity, clearWarehouse,
	)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "vehicle not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &vehiclesv1.UpdateVehicleResponse{Vehicle: toProto(v)}, nil
}

func (s *Server) DeleteVehicle(ctx context.Context, req *vehiclesv1.DeleteVehicleRequest) (*vehiclesv1.DeleteVehicleResponse, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "vehicle not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &vehiclesv1.DeleteVehicleResponse{}, nil
}
