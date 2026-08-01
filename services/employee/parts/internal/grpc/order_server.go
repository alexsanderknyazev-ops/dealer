package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/dealer/dealer/services/parts/internal/domain"
	"github.com/dealer/dealer/services/parts/internal/service"
	partsv1 "github.com/dealer/dealer/pkg/pb/parts/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func orderErr(err error) error {
	switch {
	case errors.Is(err, service.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrOrderNotDraft), errors.Is(err, service.ErrOrderNotEditable),
		errors.Is(err, service.ErrOrderNoLines), errors.Is(err, service.ErrOrderAlreadyLinked),
		errors.Is(err, service.ErrOrderAlreadyLinkedWO), errors.Is(err, service.ErrWorkOrdersUnavailable):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.ErrVehicleRequiredForWO):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrUnitPriceRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrNotFound), errors.Is(err, service.ErrSupplierNotFound),
		errors.Is(err, service.ErrWarehouseNotFound), errors.Is(err, service.ErrCustomerNotFound),
		errors.Is(err, service.ErrVehicleNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return movementDocumentErr(err)
	}
}

func parseOrderLines(req []*partsv1.PartOrderLineInput) ([]domain.PartOrderLineInput, error) {
	out := make([]domain.PartOrderLineInput, 0, len(req))
	for _, it := range req {
		partID, err := uuid.Parse(it.PartId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid part_id")
		}
		out = append(out, domain.PartOrderLineInput{
			PartID: partID, Quantity: it.Quantity, UnitPrice: it.UnitPrice,
			Notes: it.Notes, SortOrder: it.SortOrder,
		})
	}
	return out, nil
}

func (s *Server) enrichOrderLine(ctx context.Context, l *domain.PartOrderLine) *partsv1.PartOrderLine {
	line := &partsv1.PartOrderLine{
		Id: l.ID.String(), PartId: l.PartID.String(), Quantity: l.Quantity,
		UnitPrice: l.UnitPrice, Notes: l.Notes, SortOrder: l.SortOrder,
	}
	if part, err := s.svc.Get(ctx, l.PartID.String()); err == nil && part != nil {
		line.PartName = part.Name
		line.PartSku = part.SKU
	}
	return line
}

func (s *Server) supplierOrderToProto(ctx context.Context, o *domain.SupplierOrder) *partsv1.SupplierOrder {
	if o == nil {
		return nil
	}
	out := &partsv1.SupplierOrder{
		Id: o.ID.String(), OrderNumber: o.OrderNumber, Status: o.Status,
		SupplierId: o.SupplierID.String(), ReceiptWarehouseId: o.ReceiptWarehouseID.String(),
		Notes: o.Notes, CreatedAt: o.CreatedAt.Unix(), UpdatedAt: o.UpdatedAt.Unix(),
	}
	out.SupplierName = s.svc.SupplierName(ctx, o.SupplierID)
	if s.dealerPoints != nil {
		out.ReceiptWarehouseName = s.dealerPoints.WarehouseName(ctx, o.ReceiptWarehouseID)
	}
	if o.FulfillmentMovementDocumentID != nil {
		out.FulfillmentMovementDocumentId = o.FulfillmentMovementDocumentID.String()
		if doc, err := s.svc.GetMovementDocument(ctx, o.FulfillmentMovementDocumentID.String()); err == nil && doc != nil {
			out.FulfillmentMovementDocumentNumber = doc.DocumentNumber
		}
	}
	if o.FulfillmentWorkOrderID != nil {
		out.FulfillmentWorkOrderId = o.FulfillmentWorkOrderID.String()
		out.FulfillmentWorkOrderNumber = s.svc.WorkOrderNumber(ctx, *o.FulfillmentWorkOrderID)
	}
	if o.CustomerOrderID != nil {
		out.CustomerOrderId = o.CustomerOrderID.String()
		out.CustomerOrderNumber = s.svc.GetCustomerOrderNumber(ctx, *o.CustomerOrderID)
	}
	if o.CreatedBy != nil {
		out.CreatedBy = o.CreatedBy.String()
		if s.employees != nil {
			out.CreatedByName = s.employees.FullName(ctx, *o.CreatedBy)
		}
	}
	out.Lines = make([]*partsv1.PartOrderLine, len(o.Lines))
	for i, l := range o.Lines {
		out.Lines[i] = s.enrichOrderLine(ctx, &l)
	}
	return out
}

func (s *Server) customerOrderToProto(ctx context.Context, o *domain.CustomerOrder) *partsv1.CustomerOrder {
	if o == nil {
		return nil
	}
	out := &partsv1.CustomerOrder{
		Id: o.ID.String(), OrderNumber: o.OrderNumber, Status: o.Status,
		CustomerId: o.CustomerID.String(), IssueWarehouseId: o.IssueWarehouseID.String(),
		Notes: o.Notes, CreatedAt: o.CreatedAt.Unix(), UpdatedAt: o.UpdatedAt.Unix(),
	}
	out.CustomerName = s.svc.CustomerName(ctx, o.CustomerID)
	if o.VehicleID != nil {
		out.VehicleId = o.VehicleID.String()
		vin, label := s.svc.VehicleInfo(ctx, *o.VehicleID)
		out.VehicleVin = vin
		out.VehicleLabel = label
	}
	if s.dealerPoints != nil {
		out.IssueWarehouseName = s.dealerPoints.WarehouseName(ctx, o.IssueWarehouseID)
	}
	if o.FulfillmentMovementDocumentID != nil {
		out.FulfillmentMovementDocumentId = o.FulfillmentMovementDocumentID.String()
		if doc, err := s.svc.GetMovementDocument(ctx, o.FulfillmentMovementDocumentID.String()); err == nil && doc != nil {
			out.FulfillmentMovementDocumentNumber = doc.DocumentNumber
		}
	}
	if o.FulfillmentWorkOrderID != nil {
		out.FulfillmentWorkOrderId = o.FulfillmentWorkOrderID.String()
		out.FulfillmentWorkOrderNumber = s.svc.WorkOrderNumber(ctx, *o.FulfillmentWorkOrderID)
	}
	if o.CreatedBy != nil {
		out.CreatedBy = o.CreatedBy.String()
		if s.employees != nil {
			out.CreatedByName = s.employees.FullName(ctx, *o.CreatedBy)
		}
	}
	out.Lines = make([]*partsv1.PartOrderLine, len(o.Lines))
	for i, l := range o.Lines {
		out.Lines[i] = s.enrichOrderLine(ctx, &l)
	}
	return out
}

func (s *Server) CreateSupplierOrder(ctx context.Context, req *partsv1.CreateSupplierOrderRequest) (*partsv1.CreateSupplierOrderResponse, error) {
	supplierID, err := uuid.Parse(req.SupplierId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid supplier_id")
	}
	whID, err := uuid.Parse(req.ReceiptWarehouseId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid receipt_warehouse_id")
	}
	lines, err := parseOrderLines(req.Lines)
	if err != nil {
		return nil, err
	}
	var createdBy *uuid.UUID
	if req.CreatedBy != "" {
		if id, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = &id
		}
	}
	o, err := s.svc.CreateSupplierOrder(ctx, domain.CreateSupplierOrderInput{
		SupplierID: supplierID, ReceiptWarehouseID: whID, CustomerOrderID: parseOptionalUUIDField(req.CustomerOrderId),
		Notes: req.Notes, CreatedBy: createdBy, Lines: lines,
	})
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.CreateSupplierOrderResponse{Order: s.supplierOrderToProto(ctx, o)}, nil
}

func (s *Server) GetSupplierOrder(ctx context.Context, req *partsv1.GetSupplierOrderRequest) (*partsv1.GetSupplierOrderResponse, error) {
	o, err := s.svc.GetSupplierOrder(ctx, req.Id)
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.GetSupplierOrderResponse{Order: s.supplierOrderToProto(ctx, o)}, nil
}

func (s *Server) UpdateSupplierOrder(ctx context.Context, req *partsv1.UpdateSupplierOrderRequest) (*partsv1.UpdateSupplierOrderResponse, error) {
	in := domain.UpdateSupplierOrderInput{ReplaceLines: req.ReplaceLines, ClearCustomerOrder: req.ClearCustomerOrder}
	if req.SupplierId != nil {
		if id, err := uuid.Parse(req.GetSupplierId()); err == nil {
			in.SupplierID = &id
		}
	}
	if req.ReceiptWarehouseId != nil {
		if id, err := uuid.Parse(req.GetReceiptWarehouseId()); err == nil {
			in.ReceiptWarehouseID = &id
		}
	}
	if req.CustomerOrderId != nil {
		in.CustomerOrderID = parseOptionalUUIDField(req.GetCustomerOrderId())
	}
	if req.Notes != nil {
		v := req.GetNotes()
		in.Notes = &v
	}
	if req.ReplaceLines {
		lines, err := parseOrderLines(req.Lines)
		if err != nil {
			return nil, err
		}
		in.Lines = lines
	}
	o, err := s.svc.UpdateSupplierOrder(ctx, req.Id, in)
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.UpdateSupplierOrderResponse{Order: s.supplierOrderToProto(ctx, o)}, nil
}

func (s *Server) ListSupplierOrders(ctx context.Context, req *partsv1.ListSupplierOrdersRequest) (*partsv1.ListSupplierOrdersResponse, error) {
	list, total, err := s.svc.ListSupplierOrders(ctx, req.Limit, req.Offset, req.Status)
	if err != nil {
		return nil, orderErr(err)
	}
	out := make([]*partsv1.SupplierOrder, len(list))
	for i, o := range list {
		full, err := s.svc.GetSupplierOrder(ctx, o.ID.String())
		if err != nil {
			out[i] = s.supplierOrderToProto(ctx, o)
		} else {
			out[i] = s.supplierOrderToProto(ctx, full)
		}
	}
	return &partsv1.ListSupplierOrdersResponse{Orders: out, Total: total}, nil
}

func (s *Server) CancelSupplierOrder(ctx context.Context, req *partsv1.CancelSupplierOrderRequest) (*partsv1.CancelSupplierOrderResponse, error) {
	o, err := s.svc.CancelSupplierOrder(ctx, req.Id)
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.CancelSupplierOrderResponse{Order: s.supplierOrderToProto(ctx, o)}, nil
}

func (s *Server) CreateReceiptFromSupplierOrder(ctx context.Context, req *partsv1.CreateReceiptFromSupplierOrderRequest) (*partsv1.CreateReceiptFromSupplierOrderResponse, error) {
	var createdBy *uuid.UUID
	if req.CreatedBy != "" {
		if id, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = &id
		}
	}
	doc, err := s.svc.CreateReceiptFromSupplierOrder(ctx, req.Id, createdBy)
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.CreateReceiptFromSupplierOrderResponse{Document: s.documentToProto(ctx, doc)}, nil
}

func (s *Server) CreateCustomerOrder(ctx context.Context, req *partsv1.CreateCustomerOrderRequest) (*partsv1.CreateCustomerOrderResponse, error) {
	customerID, err := uuid.Parse(req.CustomerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid customer_id")
	}
	whID, err := uuid.Parse(req.IssueWarehouseId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid issue_warehouse_id")
	}
	lines, err := parseOrderLines(req.Lines)
	if err != nil {
		return nil, err
	}
	var createdBy *uuid.UUID
	if req.CreatedBy != "" {
		if id, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = &id
		}
	}
	o, err := s.svc.CreateCustomerOrder(ctx, domain.CreateCustomerOrderInput{
		CustomerID: customerID, VehicleID: parseOptionalUUIDField(req.VehicleId),
		VehicleVIN: req.VehicleVin, IssueWarehouseID: whID, Notes: req.Notes, CreatedBy: createdBy, Lines: lines,
	})
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.CreateCustomerOrderResponse{Order: s.customerOrderToProto(ctx, o)}, nil
}

func (s *Server) GetCustomerOrder(ctx context.Context, req *partsv1.GetCustomerOrderRequest) (*partsv1.GetCustomerOrderResponse, error) {
	o, err := s.svc.GetCustomerOrder(ctx, req.Id)
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.GetCustomerOrderResponse{Order: s.customerOrderToProto(ctx, o)}, nil
}

func (s *Server) UpdateCustomerOrder(ctx context.Context, req *partsv1.UpdateCustomerOrderRequest) (*partsv1.UpdateCustomerOrderResponse, error) {
	in := domain.UpdateCustomerOrderInput{ReplaceLines: req.ReplaceLines, ClearVehicle: req.ClearVehicle}
	if req.CustomerId != nil {
		in.CustomerID = parseOptionalUUIDField(req.GetCustomerId())
	}
	if req.VehicleId != nil {
		in.VehicleID = parseOptionalUUIDField(req.GetVehicleId())
	}
	if req.VehicleVin != nil {
		v := req.GetVehicleVin()
		in.VehicleVIN = &v
	}
	if req.IssueWarehouseId != nil {
		if id, err := uuid.Parse(req.GetIssueWarehouseId()); err == nil {
			in.IssueWarehouseID = &id
		}
	}
	if req.Notes != nil {
		v := req.GetNotes()
		in.Notes = &v
	}
	if req.ReplaceLines {
		lines, err := parseOrderLines(req.Lines)
		if err != nil {
			return nil, err
		}
		in.Lines = lines
	}
	o, err := s.svc.UpdateCustomerOrder(ctx, req.Id, in)
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.UpdateCustomerOrderResponse{Order: s.customerOrderToProto(ctx, o)}, nil
}

func (s *Server) ListCustomerOrders(ctx context.Context, req *partsv1.ListCustomerOrdersRequest) (*partsv1.ListCustomerOrdersResponse, error) {
	list, total, err := s.svc.ListCustomerOrders(ctx, req.Limit, req.Offset, req.Status)
	if err != nil {
		return nil, orderErr(err)
	}
	out := make([]*partsv1.CustomerOrder, len(list))
	for i, o := range list {
		full, err := s.svc.GetCustomerOrder(ctx, o.ID.String())
		if err != nil {
			out[i] = s.customerOrderToProto(ctx, o)
		} else {
			out[i] = s.customerOrderToProto(ctx, full)
		}
	}
	return &partsv1.ListCustomerOrdersResponse{Orders: out, Total: total}, nil
}

func (s *Server) CancelCustomerOrder(ctx context.Context, req *partsv1.CancelCustomerOrderRequest) (*partsv1.CancelCustomerOrderResponse, error) {
	o, err := s.svc.CancelCustomerOrder(ctx, req.Id)
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.CancelCustomerOrderResponse{Order: s.customerOrderToProto(ctx, o)}, nil
}

func (s *Server) CreateSaleFromCustomerOrder(ctx context.Context, req *partsv1.CreateSaleFromCustomerOrderRequest) (*partsv1.CreateSaleFromCustomerOrderResponse, error) {
	var createdBy *uuid.UUID
	if req.CreatedBy != "" {
		if id, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = &id
		}
	}
	doc, err := s.svc.CreateSaleFromCustomerOrder(ctx, req.Id, createdBy)
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.CreateSaleFromCustomerOrderResponse{Document: s.documentToProto(ctx, doc)}, nil
}

func (s *Server) CreateWorkOrderFromSupplierOrder(ctx context.Context, req *partsv1.CreateWorkOrderFromSupplierOrderRequest) (*partsv1.CreateWorkOrderFromOrderResponse, error) {
	customerID, err := uuid.Parse(req.CustomerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid customer_id")
	}
	wo, err := s.svc.CreateWorkOrderFromSupplierOrder(ctx, req.Id, service.CreateWorkOrderFromSupplierOrderInput{
		CustomerID: customerID, VehicleID: parseOptionalUUIDField(req.VehicleId),
		VehicleVIN: req.VehicleVin, Notes: req.Notes,
	})
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.CreateWorkOrderFromOrderResponse{WorkOrderId: wo.ID, WorkOrderNumber: wo.OrderNumber}, nil
}

func (s *Server) CreateWorkOrderFromCustomerOrder(ctx context.Context, req *partsv1.CreateWorkOrderFromCustomerOrderRequest) (*partsv1.CreateWorkOrderFromOrderResponse, error) {
	wo, err := s.svc.CreateWorkOrderFromCustomerOrder(ctx, req.Id, service.CreateWorkOrderFromCustomerOrderInput{
		VehicleID: parseOptionalUUIDField(req.VehicleId), VehicleVIN: req.VehicleVin, Notes: req.Notes,
	})
	if err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.CreateWorkOrderFromOrderResponse{WorkOrderId: wo.ID, WorkOrderNumber: wo.OrderNumber}, nil
}

func (s *Server) FulfillOrderFromWorkOrder(ctx context.Context, req *partsv1.FulfillOrderFromWorkOrderRequest) (*partsv1.FulfillOrderFromWorkOrderResponse, error) {
	if err := s.svc.FulfillOrderFromWorkOrder(ctx, req.SourceOrderType, req.SourceOrderId); err != nil {
		return nil, orderErr(err)
	}
	return &partsv1.FulfillOrderFromWorkOrderResponse{}, nil
}
