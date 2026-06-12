package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	workordersv1 "github.com/dealer/dealer/pkg/pb/workorders/v1"
	"github.com/dealer/dealer/services/workorders/internal/domain"
	"github.com/dealer/dealer/services/workorders/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EmployeeNamer interface {
	EmployeeFullName(ctx context.Context, id uuid.UUID) string
}

type ReferenceDisplayer interface {
	EmployeeNamer
	CustomerName(ctx context.Context, id uuid.UUID) string
	VehicleDisplay(ctx context.Context, id uuid.UUID) (vin, label string)
	PartDisplay(ctx context.Context, id uuid.UUID) (name, sku string)
	WarehouseName(ctx context.Context, id uuid.UUID) string
	WorkDisplay(ctx context.Context, id uuid.UUID) (code, name, laborHours string)
}

type Server struct {
	workordersv1.UnimplementedWorkOrdersServiceServer
	svc       *service.WorkOrderService
	employees EmployeeNamer
	refs      ReferenceDisplayer
}

func NewServer(svc *service.WorkOrderService, refs ReferenceDisplayer) *Server {
	return &Server{svc: svc, employees: refs, refs: refs}
}

func (s *Server) toProto(ctx context.Context, wo *domain.WorkOrder) *workordersv1.WorkOrder {
	if wo == nil {
		return nil
	}
	out := &workordersv1.WorkOrder{
		Id:            wo.ID.String(),
		OrderNumber:   wo.OrderNumber,
		CustomerId:    wo.CustomerID.String(),
		VehicleId:     wo.VehicleID.String(),
		RepairType:    wo.RepairType,
		Status:        wo.Status,
		Complaint:     wo.Complaint,
		Diagnosis:     wo.Diagnosis,
		MileageKm:     wo.MileageKm,
		LaborCost:     wo.LaborCost,
		PartsCost:     wo.PartsCost,
		TotalCost:     wo.TotalCost,
		PartsIssued:   wo.PartsIssued,
		Notes:         wo.Notes,
		CreatedAt:     wo.CreatedAt.Unix(),
		UpdatedAt:     wo.UpdatedAt.Unix(),
	}
	if wo.DealerPointID != nil {
		out.DealerPointId = wo.DealerPointID.String()
	}
	if wo.WarehouseID != nil {
		out.WarehouseId = wo.WarehouseID.String()
	}
	if s.refs != nil {
		out.CustomerName = s.refs.CustomerName(ctx, wo.CustomerID)
		out.VehicleVin, out.VehicleLabel = s.refs.VehicleDisplay(ctx, wo.VehicleID)
	}
	if wo.ServiceAdvisorID != nil {
		out.ServiceAdvisorId = wo.ServiceAdvisorID.String()
		if s.employees != nil {
			out.ServiceAdvisorName = s.employees.EmployeeFullName(ctx, *wo.ServiceAdvisorID)
		}
	}
	if wo.OpenedAt != nil {
		out.OpenedAt = wo.OpenedAt.Unix()
	}
	if wo.ClosedAt != nil {
		out.ClosedAt = wo.ClosedAt.Unix()
	}
	if wo.PartsIssuedAt != nil {
		out.PartsIssuedAt = wo.PartsIssuedAt.Unix()
	}
	if wo.MovementDocumentID != nil {
		out.MovementDocumentId = wo.MovementDocumentID.String()
	}
	out.MovementDocumentStatus = wo.MovementDocumentStatus
	out.Labor = make([]*workordersv1.WorkOrderLabor, len(wo.Labor))
	for i, l := range wo.Labor {
		item := &workordersv1.WorkOrderLabor{
			Id:          l.ID.String(),
			Description: l.Description,
			Quantity:    l.Quantity,
			UnitPrice:   l.UnitPrice,
			Amount:      l.Amount,
			SortOrder:   l.SortOrder,
		}
		if l.WorkID != nil {
			item.WorkId = l.WorkID.String()
		}
		if l.ExecutorID != nil {
			item.ExecutorId = l.ExecutorID.String()
			if s.employees != nil {
				item.ExecutorName = s.employees.EmployeeFullName(ctx, *l.ExecutorID)
			}
		}
		if l.WorkID != nil && s.refs != nil {
			code, name, laborHours := s.refs.WorkDisplay(ctx, *l.WorkID)
			item.WorkCode = code
			item.WorkName = name
			item.LaborHours = laborHours
			if name != "" && item.Description == "" {
				item.Description = name
			}
		}
		out.Labor[i] = item
	}
	out.Parts = make([]*workordersv1.WorkOrderPart, len(wo.Parts))
	for i, p := range wo.Parts {
		item := &workordersv1.WorkOrderPart{
			Id:          p.ID.String(),
			PartId:      p.PartID.String(),
			WarehouseId: p.WarehouseID.String(),
			Description: p.Description,
			Quantity:    p.Quantity,
			UnitPrice:   p.UnitPrice,
			Amount:      p.Amount,
			Issued:      p.Issued,
			SortOrder:   p.SortOrder,
		}
		if s.refs != nil {
			name, sku := s.refs.PartDisplay(ctx, p.PartID)
			item.PartName = name
			if name != "" && item.Description == "" {
				item.Description = name
			}
			item.PartSku = sku
			item.WarehouseName = s.refs.WarehouseName(ctx, p.WarehouseID)
		}
		out.Parts[i] = item
	}
	return out
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, "work order not found")
	case errors.Is(err, service.ErrCustomerNotFound),
		errors.Is(err, service.ErrVehicleNotFound),
		errors.Is(err, service.ErrDealerPointNotFound),
		errors.Is(err, service.ErrWarehouseNotFound),
		errors.Is(err, service.ErrPartNotFound),
		errors.Is(err, service.ErrWorkNotFound),
		errors.Is(err, service.ErrEmployeeNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrNoPartsToIssue),
		errors.Is(err, service.ErrPartsAlreadyIssued),
		errors.Is(err, service.ErrMovementDocumentExists):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		if err != nil && err.Error() != "" {
			return status.Error(codes.Internal, err.Error())
		}
		return status.Error(codes.Internal, "internal error")
	}
}

func laborInputs(items []*workordersv1.WorkOrderLaborInput) []service.LaborInput {
	out := make([]service.LaborInput, len(items))
	for i, it := range items {
		out[i] = service.LaborInput{
			WorkID:      it.WorkId,
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
			ExecutorID:  it.ExecutorId,
			SortOrder:   it.SortOrder,
		}
	}
	return out
}

func partInputs(items []*workordersv1.WorkOrderPartInput) []service.PartInput {
	out := make([]service.PartInput, len(items))
	for i, it := range items {
		out[i] = service.PartInput{
			PartID:      it.PartId,
			WarehouseID: it.WarehouseId,
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
			SortOrder:   it.SortOrder,
		}
	}
	return out
}

func (s *Server) CreateWorkOrder(ctx context.Context, req *workordersv1.CreateWorkOrderRequest) (*workordersv1.CreateWorkOrderResponse, error) {
	wo, err := s.svc.Create(ctx, service.CreateInput{
		CustomerID: req.CustomerId, VehicleID: req.VehicleId, DealerPointID: req.DealerPointId,
		WarehouseID: req.WarehouseId, RepairType: req.RepairType, Status: req.Status,
		ServiceAdvisorID: req.ServiceAdvisorId, Complaint: req.Complaint, Diagnosis: req.Diagnosis,
		MileageKm: req.MileageKm, OpenedAt: req.OpenedAt, Notes: req.Notes,
		Labor: laborInputs(req.Labor), Parts: partInputs(req.Parts),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &workordersv1.CreateWorkOrderResponse{WorkOrder: s.toProto(ctx, wo)}, nil
}

func (s *Server) GetWorkOrder(ctx context.Context, req *workordersv1.GetWorkOrderRequest) (*workordersv1.GetWorkOrderResponse, error) {
	wo, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &workordersv1.GetWorkOrderResponse{WorkOrder: s.toProto(ctx, wo)}, nil
}

func (s *Server) ListWorkOrders(ctx context.Context, req *workordersv1.ListWorkOrdersRequest) (*workordersv1.ListWorkOrdersResponse, error) {
	list, total, err := s.svc.List(ctx, req.Limit, req.Offset, req.Status, req.RepairType, req.CustomerId, req.VehicleId)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*workordersv1.WorkOrder, len(list))
	for i, wo := range list {
		out[i] = s.toProto(ctx, wo)
	}
	return &workordersv1.ListWorkOrdersResponse{WorkOrders: out, Total: total}, nil
}

func (s *Server) UpdateWorkOrder(ctx context.Context, req *workordersv1.UpdateWorkOrderRequest) (*workordersv1.UpdateWorkOrderResponse, error) {
	in := service.UpdateInput{
		CustomerID: req.CustomerId, VehicleID: req.VehicleId, DealerPointID: req.DealerPointId,
		WarehouseID: req.WarehouseId, RepairType: req.RepairType, Status: req.Status,
		ServiceAdvisorID: req.ServiceAdvisorId, Complaint: req.Complaint, Diagnosis: req.Diagnosis,
		MileageKm: req.MileageKm, OpenedAt: req.OpenedAt, ClosedAt: req.ClosedAt, Notes: req.Notes,
		ReplaceLines: len(req.Labor) > 0 || len(req.Parts) > 0,
		Labor:        laborInputs(req.Labor),
		Parts:        partInputs(req.Parts),
	}
	wo, err := s.svc.Update(ctx, req.Id, in)
	if err != nil {
		return nil, mapErr(err)
	}
	return &workordersv1.UpdateWorkOrderResponse{WorkOrder: s.toProto(ctx, wo)}, nil
}

func (s *Server) DeleteWorkOrder(ctx context.Context, req *workordersv1.DeleteWorkOrderRequest) (*workordersv1.DeleteWorkOrderResponse, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		return nil, mapErr(err)
	}
	return &workordersv1.DeleteWorkOrderResponse{}, nil
}

func (s *Server) MovePartsToWork(ctx context.Context, req *workordersv1.MovePartsToWorkRequest) (*workordersv1.MovePartsToWorkResponse, error) {
	wo, err := s.svc.MovePartsToWork(ctx, req.Id, req.IssuedBy)
	if err != nil {
		return nil, mapErr(err)
	}
	return &workordersv1.MovePartsToWorkResponse{WorkOrder: s.toProto(ctx, wo)}, nil
}

func (s *Server) ApplyMovementDocument(ctx context.Context, req *workordersv1.ApplyMovementDocumentRequest) (*workordersv1.ApplyMovementDocumentResponse, error) {
	wo, err := s.svc.ApplyMovementDocument(ctx, req.WorkOrderId, req.MovementDocumentId, req.MovementDocumentStatus)
	if err != nil {
		return nil, mapErr(err)
	}
	return &workordersv1.ApplyMovementDocumentResponse{WorkOrder: s.toProto(ctx, wo)}, nil
}
