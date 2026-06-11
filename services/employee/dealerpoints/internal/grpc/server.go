package grpc

import (
	"context"
	"errors"

	dealerpointsv1 "github.com/dealer/dealer/pkg/pb/dealerpoints/v1"
	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
	"github.com/dealer/dealer/services/dealerpoints/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	dealerpointsv1.UnimplementedDealerPointsServiceServer
	svc *service.DealerPointsService
}

func NewServer(svc *service.DealerPointsService) *Server {
	return &Server{svc: svc}
}

func toDealerPointProto(d *domain.DealerPoint) *dealerpointsv1.DealerPoint {
	if d == nil {
		return nil
	}
	return &dealerpointsv1.DealerPoint{
		Id:        d.ID.String(),
		Name:      d.Name,
		Address:   d.Address,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
	}
}

func toLegalEntityProto(e *domain.LegalEntity) *dealerpointsv1.LegalEntity {
	if e == nil {
		return nil
	}
	return &dealerpointsv1.LegalEntity{
		Id:        e.ID.String(),
		Name:      e.Name,
		Inn:       e.INN,
		Address:   e.Address,
		CreatedAt: e.CreatedAt.Unix(),
		UpdatedAt: e.UpdatedAt.Unix(),
	}
}

func toWarehouseProto(w *domain.Warehouse) *dealerpointsv1.Warehouse {
	if w == nil {
		return nil
	}
	return &dealerpointsv1.Warehouse{
		Id:             w.ID.String(),
		DealerPointId:  w.DealerPointID.String(),
		LegalEntityId:  w.LegalEntityID.String(),
		Type:           w.Type,
		Name:           w.Name,
		CreatedAt:      w.CreatedAt.Unix(),
		UpdatedAt:      w.UpdatedAt.Unix(),
	}
}

func errToStatus(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, service.ErrDealerPointNotFound), errors.Is(err, service.ErrLegalEntityNotFound), errors.Is(err, service.ErrWarehouseNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

// Dealer points
func (s *Server) CreateDealerPoint(ctx context.Context, req *dealerpointsv1.CreateDealerPointRequest) (*dealerpointsv1.CreateDealerPointResponse, error) {
	d, err := s.svc.CreateDealerPoint(ctx, req.Name, req.Address)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.CreateDealerPointResponse{DealerPoint: toDealerPointProto(d)}, nil
}

func (s *Server) GetDealerPoint(ctx context.Context, req *dealerpointsv1.GetDealerPointRequest) (*dealerpointsv1.GetDealerPointResponse, error) {
	d, err := s.svc.GetDealerPoint(ctx, req.Id)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.GetDealerPointResponse{DealerPoint: toDealerPointProto(d)}, nil
}

func (s *Server) ListDealerPoints(ctx context.Context, req *dealerpointsv1.ListDealerPointsRequest) (*dealerpointsv1.ListDealerPointsResponse, error) {
	list, total, err := s.svc.ListDealerPoints(ctx, req.Limit, req.Offset, req.Search)
	if err != nil {
		return nil, errToStatus(err)
	}
	out := make([]*dealerpointsv1.DealerPoint, len(list))
	for i, d := range list {
		out[i] = toDealerPointProto(d)
	}
	return &dealerpointsv1.ListDealerPointsResponse{DealerPoints: out, Total: total}, nil
}

func (s *Server) UpdateDealerPoint(ctx context.Context, req *dealerpointsv1.UpdateDealerPointRequest) (*dealerpointsv1.UpdateDealerPointResponse, error) {
	var name, address *string
	if req.Name != nil {
		v := req.GetName()
		name = &v
	}
	if req.Address != nil {
		v := req.GetAddress()
		address = &v
	}
	d, err := s.svc.UpdateDealerPoint(ctx, req.Id, name, address)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.UpdateDealerPointResponse{DealerPoint: toDealerPointProto(d)}, nil
}

func (s *Server) DeleteDealerPoint(ctx context.Context, req *dealerpointsv1.DeleteDealerPointRequest) (*dealerpointsv1.DeleteDealerPointResponse, error) {
	if err := s.svc.DeleteDealerPoint(ctx, req.Id); err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.DeleteDealerPointResponse{}, nil
}

// Legal entities
func (s *Server) CreateLegalEntity(ctx context.Context, req *dealerpointsv1.CreateLegalEntityRequest) (*dealerpointsv1.CreateLegalEntityResponse, error) {
	e, err := s.svc.CreateLegalEntity(ctx, req.Name, req.Inn, req.Address)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.CreateLegalEntityResponse{LegalEntity: toLegalEntityProto(e)}, nil
}

func (s *Server) GetLegalEntity(ctx context.Context, req *dealerpointsv1.GetLegalEntityRequest) (*dealerpointsv1.GetLegalEntityResponse, error) {
	e, err := s.svc.GetLegalEntity(ctx, req.Id)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.GetLegalEntityResponse{LegalEntity: toLegalEntityProto(e)}, nil
}

func (s *Server) ListLegalEntities(ctx context.Context, req *dealerpointsv1.ListLegalEntitiesRequest) (*dealerpointsv1.ListLegalEntitiesResponse, error) {
	list, total, err := s.svc.ListLegalEntities(ctx, req.Limit, req.Offset, req.Search)
	if err != nil {
		return nil, errToStatus(err)
	}
	out := make([]*dealerpointsv1.LegalEntity, len(list))
	for i, e := range list {
		out[i] = toLegalEntityProto(e)
	}
	return &dealerpointsv1.ListLegalEntitiesResponse{LegalEntities: out, Total: total}, nil
}

func (s *Server) UpdateLegalEntity(ctx context.Context, req *dealerpointsv1.UpdateLegalEntityRequest) (*dealerpointsv1.UpdateLegalEntityResponse, error) {
	var name, inn, address *string
	if req.Name != nil {
		v := req.GetName()
		name = &v
	}
	if req.Inn != nil {
		v := req.GetInn()
		inn = &v
	}
	if req.Address != nil {
		v := req.GetAddress()
		address = &v
	}
	e, err := s.svc.UpdateLegalEntity(ctx, req.Id, name, inn, address)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.UpdateLegalEntityResponse{LegalEntity: toLegalEntityProto(e)}, nil
}

func (s *Server) DeleteLegalEntity(ctx context.Context, req *dealerpointsv1.DeleteLegalEntityRequest) (*dealerpointsv1.DeleteLegalEntityResponse, error) {
	if err := s.svc.DeleteLegalEntity(ctx, req.Id); err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.DeleteLegalEntityResponse{}, nil
}

func (s *Server) LinkLegalEntityToDealerPoint(ctx context.Context, req *dealerpointsv1.LinkLegalEntityRequest) (*dealerpointsv1.LinkLegalEntityResponse, error) {
	if err := s.svc.LinkLegalEntityToDealerPoint(ctx, req.DealerPointId, req.LegalEntityId); err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.LinkLegalEntityResponse{}, nil
}

func (s *Server) UnlinkLegalEntityFromDealerPoint(ctx context.Context, req *dealerpointsv1.UnlinkLegalEntityRequest) (*dealerpointsv1.UnlinkLegalEntityResponse, error) {
	if err := s.svc.UnlinkLegalEntityFromDealerPoint(ctx, req.DealerPointId, req.LegalEntityId); err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.UnlinkLegalEntityResponse{}, nil
}

func (s *Server) ListLegalEntitiesByDealerPoint(ctx context.Context, req *dealerpointsv1.ListLegalEntitiesByDealerPointRequest) (*dealerpointsv1.ListLegalEntitiesResponse, error) {
	list, total, err := s.svc.ListLegalEntitiesByDealerPoint(ctx, req.DealerPointId, req.Limit, req.Offset)
	if err != nil {
		return nil, errToStatus(err)
	}
	out := make([]*dealerpointsv1.LegalEntity, len(list))
	for i, e := range list {
		out[i] = toLegalEntityProto(e)
	}
	return &dealerpointsv1.ListLegalEntitiesResponse{LegalEntities: out, Total: total}, nil
}

// Warehouses
func (s *Server) CreateWarehouse(ctx context.Context, req *dealerpointsv1.CreateWarehouseRequest) (*dealerpointsv1.CreateWarehouseResponse, error) {
	w, err := s.svc.CreateWarehouse(ctx, req.DealerPointId, req.LegalEntityId, req.Type, req.Name)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.CreateWarehouseResponse{Warehouse: toWarehouseProto(w)}, nil
}

func (s *Server) GetWarehouse(ctx context.Context, req *dealerpointsv1.GetWarehouseRequest) (*dealerpointsv1.GetWarehouseResponse, error) {
	w, err := s.svc.GetWarehouse(ctx, req.Id)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.GetWarehouseResponse{Warehouse: toWarehouseProto(w)}, nil
}

func (s *Server) ListWarehouses(ctx context.Context, req *dealerpointsv1.ListWarehousesRequest) (*dealerpointsv1.ListWarehousesResponse, error) {
	list, total, err := s.svc.ListWarehouses(ctx, req.Limit, req.Offset, req.DealerPointId, req.LegalEntityId, req.Type)
	if err != nil {
		return nil, errToStatus(err)
	}
	out := make([]*dealerpointsv1.Warehouse, len(list))
	for i, w := range list {
		out[i] = toWarehouseProto(w)
	}
	return &dealerpointsv1.ListWarehousesResponse{Warehouses: out, Total: total}, nil
}

func (s *Server) UpdateWarehouse(ctx context.Context, req *dealerpointsv1.UpdateWarehouseRequest) (*dealerpointsv1.UpdateWarehouseResponse, error) {
	var name *string
	if req.Name != nil {
		v := req.GetName()
		name = &v
	}
	w, err := s.svc.UpdateWarehouse(ctx, req.Id, name)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.UpdateWarehouseResponse{Warehouse: toWarehouseProto(w)}, nil
}

func (s *Server) DeleteWarehouse(ctx context.Context, req *dealerpointsv1.DeleteWarehouseRequest) (*dealerpointsv1.DeleteWarehouseResponse, error) {
	if err := s.svc.DeleteWarehouse(ctx, req.Id); err != nil {
		return nil, errToStatus(err)
	}
	return &dealerpointsv1.DeleteWarehouseResponse{}, nil
}
