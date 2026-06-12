package grpc

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	appointmentsv1 "github.com/dealer/dealer/pkg/pb/appointments/v1"
	"github.com/dealer/dealer/services/appointments/internal/domain"
	"github.com/dealer/dealer/services/appointments/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReferenceDisplayer interface {
	CustomerName(ctx context.Context, id uuid.UUID) string
	VehicleDisplay(ctx context.Context, id uuid.UUID) (vin, label string)
	PartDisplay(ctx context.Context, id uuid.UUID) (name, sku string)
	WarehouseName(ctx context.Context, id uuid.UUID) string
	WorkDisplay(ctx context.Context, id uuid.UUID) (code, name string)
}

type WorkOrderNamer interface {
	GetOrderNumber(ctx context.Context, id string) string
}

type Server struct {
	appointmentsv1.UnimplementedRepairAppointmentsServiceServer
	svc        *service.RepairAppointmentService
	refs       ReferenceDisplayer
	workOrders WorkOrderNamer
}

func NewServer(svc *service.RepairAppointmentService, refs ReferenceDisplayer, workOrders WorkOrderNamer) *Server {
	return &Server{svc: svc, refs: refs, workOrders: workOrders}
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrCustomerNotFound), errors.Is(err, service.ErrVehicleNotFound),
		errors.Is(err, service.ErrWarehouseNotFound), errors.Is(err, service.ErrPartNotFound),
		errors.Is(err, service.ErrWorkNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrSlotUnavailable), errors.Is(err, service.ErrNotEditable),
		errors.Is(err, service.ErrWorkOrderExists), errors.Is(err, service.ErrWorkOrdersUnavailable):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.ErrInvalidTimeRange):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		return status.Error(codes.Internal, "internal error")
	}
}

func (s *Server) toProto(ctx context.Context, a *domain.RepairAppointment) *appointmentsv1.RepairAppointment {
	if a == nil {
		return nil
	}
	out := &appointmentsv1.RepairAppointment{
		Id: a.ID.String(), AppointmentNumber: a.AppointmentNumber,
		CustomerId: a.CustomerID.String(), VehicleId: a.VehicleID.String(),
		ScheduledStart: a.ScheduledStart.Unix(), ScheduledEnd: a.ScheduledEnd.Unix(),
		Status: a.Status, Complaint: a.Complaint, Notes: a.Notes,
		CreatedAt: a.CreatedAt.Unix(), UpdatedAt: a.UpdatedAt.Unix(),
	}
	if s.refs != nil {
		out.CustomerName = s.refs.CustomerName(ctx, a.CustomerID)
		out.VehicleVin, out.VehicleLabel = s.refs.VehicleDisplay(ctx, a.VehicleID)
	}
	if a.DealerPointID != nil {
		out.DealerPointId = a.DealerPointID.String()
	}
	if a.WarehouseID != nil {
		out.WarehouseId = a.WarehouseID.String()
	}
	if a.WorkOrderID != nil {
		out.WorkOrderId = a.WorkOrderID.String()
		if s.workOrders != nil {
			out.WorkOrderNumber = s.workOrders.GetOrderNumber(ctx, a.WorkOrderID.String())
		}
	}
	if a.CreatedBy != nil {
		out.CreatedBy = a.CreatedBy.String()
	}
	out.Labor = make([]*appointmentsv1.RepairAppointmentLabor, len(a.Labor))
	for i, l := range a.Labor {
		item := &appointmentsv1.RepairAppointmentLabor{
			Id: l.ID.String(), Description: l.Description, Quantity: l.Quantity,
			UnitPrice: l.UnitPrice, SortOrder: l.SortOrder,
		}
		if l.WorkID != nil {
			item.WorkId = l.WorkID.String()
			if s.refs != nil {
				item.WorkCode, item.WorkName = s.refs.WorkDisplay(ctx, *l.WorkID)
			}
		}
		out.Labor[i] = item
	}
	out.Parts = make([]*appointmentsv1.RepairAppointmentPart, len(a.Parts))
	for i, p := range a.Parts {
		item := &appointmentsv1.RepairAppointmentPart{
			Id: p.ID.String(), PartId: p.PartID.String(), WarehouseId: p.WarehouseID.String(),
			Quantity: service.FormatQuantity(p.Quantity), UnitPrice: p.UnitPrice,
			Notes: p.Notes, SortOrder: p.SortOrder,
		}
		if s.refs != nil {
			item.PartName, item.PartSku = s.refs.PartDisplay(ctx, p.PartID)
			item.WarehouseName = s.refs.WarehouseName(ctx, p.WarehouseID)
		}
		out.Parts[i] = item
	}
	return out
}

func parseOptionalUUIDField(raw string) *uuid.UUID {
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func laborInputs(items []*appointmentsv1.RepairAppointmentLaborInput) []domain.LaborInput {
	out := make([]domain.LaborInput, len(items))
	for i, it := range items {
		out[i] = domain.LaborInput{
			WorkID: it.WorkId, Description: it.Description, Quantity: it.Quantity,
			UnitPrice: it.UnitPrice, SortOrder: it.SortOrder,
		}
	}
	return out
}

func partInputs(items []*appointmentsv1.RepairAppointmentPartInput) []domain.PartInput {
	out := make([]domain.PartInput, len(items))
	for i, it := range items {
		qty := int32(1)
		if it.Quantity != "" {
			if v, err := strconv.ParseInt(it.Quantity, 10, 32); err == nil {
				qty = int32(v)
			}
		}
		out[i] = domain.PartInput{
			PartID: it.PartId, WarehouseID: it.WarehouseId, Quantity: qty,
			UnitPrice: it.UnitPrice, Notes: it.Notes, SortOrder: it.SortOrder,
		}
	}
	return out
}

func (s *Server) ListRepairAppointmentSlots(ctx context.Context, req *appointmentsv1.ListRepairAppointmentSlotsRequest) (*appointmentsv1.ListRepairAppointmentSlotsResponse, error) {
	slots, err := s.svc.ListSlots(ctx, req.Date)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*appointmentsv1.RepairAppointmentSlot, len(slots))
	for i, sl := range slots {
		out[i] = &appointmentsv1.RepairAppointmentSlot{
			StartAt: sl.Start.Unix(), EndAt: sl.End.Unix(), Available: sl.Available, Label: sl.Label,
		}
	}
	return &appointmentsv1.ListRepairAppointmentSlotsResponse{Slots: out}, nil
}

func (s *Server) CreateRepairAppointment(ctx context.Context, req *appointmentsv1.CreateRepairAppointmentRequest) (*appointmentsv1.CreateRepairAppointmentResponse, error) {
	cid, err := uuid.Parse(req.CustomerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid customer_id")
	}
	vid, err := uuid.Parse(req.VehicleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vehicle_id")
	}
	var createdBy *uuid.UUID
	if req.CreatedBy != "" {
		if id, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = &id
		}
	}
	a, err := s.svc.Create(ctx, domain.CreateInput{
		CustomerID: cid, VehicleID: vid,
		DealerPointID: parseOptionalUUIDField(req.DealerPointId),
		WarehouseID:   parseOptionalUUIDField(req.WarehouseId),
		ScheduledStart: time.Unix(req.ScheduledStart, 0).UTC(),
		ScheduledEnd:   time.Unix(req.ScheduledEnd, 0).UTC(),
		Complaint: req.Complaint, Notes: req.Notes, CreatedBy: createdBy,
		Labor: laborInputs(req.Labor), Parts: partInputs(req.Parts),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &appointmentsv1.CreateRepairAppointmentResponse{Appointment: s.toProto(ctx, a)}, nil
}

func (s *Server) GetRepairAppointment(ctx context.Context, req *appointmentsv1.GetRepairAppointmentRequest) (*appointmentsv1.GetRepairAppointmentResponse, error) {
	a, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &appointmentsv1.GetRepairAppointmentResponse{Appointment: s.toProto(ctx, a)}, nil
}

func (s *Server) UpdateRepairAppointment(ctx context.Context, req *appointmentsv1.UpdateRepairAppointmentRequest) (*appointmentsv1.UpdateRepairAppointmentResponse, error) {
	in := domain.UpdateInput{ReplaceLines: req.ReplaceLines}
	if req.CustomerId != nil {
		if id, err := uuid.Parse(req.GetCustomerId()); err == nil {
			in.CustomerID = &id
		}
	}
	if req.VehicleId != nil {
		if id, err := uuid.Parse(req.GetVehicleId()); err == nil {
			in.VehicleID = &id
		}
	}
	if req.DealerPointId != nil {
		in.DealerPointID = parseOptionalUUIDField(req.GetDealerPointId())
	}
	if req.WarehouseId != nil {
		in.WarehouseID = parseOptionalUUIDField(req.GetWarehouseId())
	}
	if req.ScheduledStart != nil {
		t := time.Unix(req.GetScheduledStart(), 0).UTC()
		in.ScheduledStart = &t
	}
	if req.ScheduledEnd != nil {
		t := time.Unix(req.GetScheduledEnd(), 0).UTC()
		in.ScheduledEnd = &t
	}
	if req.Complaint != nil {
		v := req.GetComplaint()
		in.Complaint = &v
	}
	if req.Notes != nil {
		v := req.GetNotes()
		in.Notes = &v
	}
	if req.ReplaceLines {
		in.Labor = laborInputs(req.Labor)
		in.Parts = partInputs(req.Parts)
	}
	a, err := s.svc.Update(ctx, req.Id, in)
	if err != nil {
		return nil, mapErr(err)
	}
	return &appointmentsv1.UpdateRepairAppointmentResponse{Appointment: s.toProto(ctx, a)}, nil
}

func (s *Server) ListRepairAppointments(ctx context.Context, req *appointmentsv1.ListRepairAppointmentsRequest) (*appointmentsv1.ListRepairAppointmentsResponse, error) {
	var from, to *time.Time
	if req.DateFrom > 0 {
		t := time.Unix(req.DateFrom, 0).UTC()
		from = &t
	}
	if req.DateTo > 0 {
		t := time.Unix(req.DateTo, 0).UTC()
		to = &t
	}
	list, total, err := s.svc.List(ctx, req.Limit, req.Offset, req.Status, from, to)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*appointmentsv1.RepairAppointment, len(list))
	for i, a := range list {
		full, err := s.svc.Get(ctx, a.ID.String())
		if err != nil {
			out[i] = s.toProto(ctx, a)
		} else {
			out[i] = s.toProto(ctx, full)
		}
	}
	return &appointmentsv1.ListRepairAppointmentsResponse{Appointments: out, Total: total}, nil
}

func (s *Server) CancelRepairAppointment(ctx context.Context, req *appointmentsv1.CancelRepairAppointmentRequest) (*appointmentsv1.CancelRepairAppointmentResponse, error) {
	a, err := s.svc.Cancel(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &appointmentsv1.CancelRepairAppointmentResponse{Appointment: s.toProto(ctx, a)}, nil
}

func (s *Server) CreateWorkOrderFromRepairAppointment(ctx context.Context, req *appointmentsv1.CreateWorkOrderFromRepairAppointmentRequest) (*appointmentsv1.CreateWorkOrderFromRepairAppointmentResponse, error) {
	a, woID, woNum, err := s.svc.CreateWorkOrder(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &appointmentsv1.CreateWorkOrderFromRepairAppointmentResponse{
		WorkOrderId: woID, WorkOrderNumber: woNum, Appointment: s.toProto(ctx, a),
	}, nil
}
