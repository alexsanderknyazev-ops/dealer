package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderNotDraft      = errors.New("order is not draft")
	ErrOrderNotEditable   = errors.New("order cannot be edited")
	ErrOrderNoLines       = errors.New("order has no lines")
	ErrOrderAlreadyLinked   = errors.New("order already linked to fulfillment document")
	ErrOrderAlreadyLinkedWO = errors.New("order already linked to work order")
	ErrWorkOrdersUnavailable = errors.New("workorders service unavailable")
	ErrVehicleRequiredForWO  = errors.New("vehicle required for work order")
	ErrUnitPriceRequired     = errors.New("unit price required for order line")
)

type supplierOrderRepository interface {
	NextOrderNumber(ctx context.Context) (string, error)
	Create(ctx context.Context, o *domain.SupplierOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SupplierOrder, error)
	List(ctx context.Context, limit, offset int32, status string) ([]*domain.SupplierOrder, int32, error)
	Update(ctx context.Context, o *domain.SupplierOrder, replaceLines bool) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, movementDocID *uuid.UUID, updatedAt time.Time) error
	LinkWorkOrder(ctx context.Context, id uuid.UUID, workOrderID uuid.UUID, updatedAt time.Time) error
	MarkFulfilled(ctx context.Context, id uuid.UUID) error
}

type customerOrderRepository interface {
	NextOrderNumber(ctx context.Context) (string, error)
	Create(ctx context.Context, o *domain.CustomerOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomerOrder, error)
	List(ctx context.Context, limit, offset int32, status string) ([]*domain.CustomerOrder, int32, error)
	Update(ctx context.Context, o *domain.CustomerOrder, replaceLines bool) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, movementDocID *uuid.UUID, updatedAt time.Time) error
	LinkWorkOrder(ctx context.Context, id uuid.UUID, workOrderID uuid.UUID, updatedAt time.Time) error
	MarkFulfilled(ctx context.Context, id uuid.UUID) error
}

func parseUnitPrice(raw string) (string, error) {
	return parseUnitCost(raw, true)
}

func (s *PartService) buildOrderLines(
	ctx context.Context,
	orderID uuid.UUID,
	inputs []domain.PartOrderLineInput,
	now time.Time,
) ([]domain.PartOrderLine, error) {
	lines := make([]domain.PartOrderLine, 0, len(inputs))
	for _, in := range inputs {
		if in.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for part %s", in.PartID)
		}
		price, err := parseUnitPrice(in.UnitPrice)
		if err != nil {
			return nil, err
		}
		if _, err := s.repo.GetByID(ctx, in.PartID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		lines = append(lines, domain.PartOrderLine{
			ID: uuid.New(), OrderID: orderID, PartID: in.PartID, Quantity: in.Quantity,
			UnitPrice: price, Notes: in.Notes, SortOrder: in.SortOrder, CreatedAt: now,
		})
	}
	return lines, nil
}

func (s *PartService) CreateSupplierOrder(ctx context.Context, in domain.CreateSupplierOrderInput) (*domain.SupplierOrder, error) {
	if s.suppliers == nil {
		return nil, ErrSupplierNotFound
	}
	sid := in.SupplierID
	if err := s.checkRef(ctx, &sid, s.suppliers.Exists, ErrSupplierNotFound); err != nil {
		return nil, err
	}
	whID := in.ReceiptWarehouseID
	if err := s.checkRef(ctx, &whID, s.dealerPoints.WarehouseExists, ErrWarehouseNotFound); err != nil {
		return nil, err
	}
	if in.CustomerOrderID != nil {
		if _, err := s.GetCustomerOrder(ctx, in.CustomerOrderID.String()); err != nil {
			return nil, err
		}
	}
	number, err := s.supplierOrderRepo.NextOrderNumber(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	orderID := uuid.New()
	var lines []domain.PartOrderLine
	if len(in.Lines) > 0 {
		lines, err = s.buildOrderLines(ctx, orderID, in.Lines, now)
		if err != nil {
			return nil, err
		}
	}
	o := &domain.SupplierOrder{
		ID: orderID, OrderNumber: number, Status: domain.OrderStatusDraft,
		SupplierID: in.SupplierID, ReceiptWarehouseID: in.ReceiptWarehouseID,
		CustomerOrderID: in.CustomerOrderID,
		Notes: in.Notes, CreatedBy: in.CreatedBy, Lines: lines, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.supplierOrderRepo.Create(ctx, o); err != nil {
		return nil, err
	}
	return s.GetSupplierOrder(ctx, orderID.String())
}

func (s *PartService) GetSupplierOrder(ctx context.Context, id string) (*domain.SupplierOrder, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	o, err := s.supplierOrderRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return o, nil
}

func (s *PartService) ListSupplierOrders(ctx context.Context, limit, offset int32, status string) ([]*domain.SupplierOrder, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.supplierOrderRepo.List(ctx, limit, offset, status)
}

func (s *PartService) UpdateSupplierOrder(ctx context.Context, id string, in domain.UpdateSupplierOrderInput) (*domain.SupplierOrder, error) {
	o, err := s.GetSupplierOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if o.Status != domain.OrderStatusDraft {
		return nil, ErrOrderNotEditable
	}
	now := time.Now().UTC()
	if in.SupplierID != nil {
		if err := s.checkRef(ctx, in.SupplierID, s.suppliers.Exists, ErrSupplierNotFound); err != nil {
			return nil, err
		}
		o.SupplierID = *in.SupplierID
	}
	if in.ReceiptWarehouseID != nil {
		if err := s.checkRef(ctx, in.ReceiptWarehouseID, s.dealerPoints.WarehouseExists, ErrWarehouseNotFound); err != nil {
			return nil, err
		}
		o.ReceiptWarehouseID = *in.ReceiptWarehouseID
	}
	if in.ClearCustomerOrder {
		o.CustomerOrderID = nil
	} else if in.CustomerOrderID != nil {
		if _, err := s.GetCustomerOrder(ctx, in.CustomerOrderID.String()); err != nil {
			return nil, err
		}
		o.CustomerOrderID = in.CustomerOrderID
	}
	if in.Notes != nil {
		o.Notes = *in.Notes
	}
	if in.ReplaceLines {
		o.Lines, err = s.buildOrderLines(ctx, o.ID, in.Lines, now)
		if err != nil {
			return nil, err
		}
	}
	o.UpdatedAt = now
	if err := s.supplierOrderRepo.Update(ctx, o, in.ReplaceLines); err != nil {
		return nil, err
	}
	return s.GetSupplierOrder(ctx, id)
}

func (s *PartService) CancelSupplierOrder(ctx context.Context, id string) (*domain.SupplierOrder, error) {
	o, err := s.GetSupplierOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if o.Status != domain.OrderStatusDraft {
		return nil, ErrOrderNotDraft
	}
	now := time.Now().UTC()
	if err := s.supplierOrderRepo.UpdateStatus(ctx, o.ID, domain.OrderStatusCancelled, nil, now); err != nil {
		return nil, err
	}
	return s.GetSupplierOrder(ctx, id)
}

func (s *PartService) CreateReceiptFromSupplierOrder(ctx context.Context, orderID string, createdBy *uuid.UUID) (*domain.MovementDocument, error) {
	o, err := s.GetSupplierOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != domain.OrderStatusDraft {
		return nil, ErrOrderNotDraft
	}
	if o.FulfillmentMovementDocumentID != nil || o.FulfillmentWorkOrderID != nil {
		return nil, ErrOrderAlreadyLinked
	}
	if len(o.Lines) == 0 {
		return nil, ErrOrderNoLines
	}
	refID := o.ID
	lineInputs := make([]domain.MovementDocumentLineInput, len(o.Lines))
	for i, l := range o.Lines {
		refLine := l.ID
		lineInputs[i] = domain.MovementDocumentLineInput{
			PartID: l.PartID, WarehouseID: o.ReceiptWarehouseID, Quantity: l.Quantity,
			UnitCost: l.UnitPrice, ReferenceLineID: &refLine, Notes: l.Notes, SortOrder: l.SortOrder,
		}
	}
	doc, err := s.CreateMovementDocument(ctx, domain.CreateMovementDocumentInput{
		MovementType: domain.MovementTypeReceipt, ReferenceType: domain.RefSupplierOrder, ReferenceID: &refID,
		SupplierID: &o.SupplierID, ReceiptWarehouseID: &o.ReceiptWarehouseID,
		Notes: fmt.Sprintf("Поступление по заказу %s", o.OrderNumber), CreatedBy: createdBy, Lines: lineInputs,
	})
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := s.supplierOrderRepo.UpdateStatus(ctx, o.ID, domain.OrderStatusLinked, &doc.ID, now); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *PartService) CreateCustomerOrder(ctx context.Context, in domain.CreateCustomerOrderInput) (*domain.CustomerOrder, error) {
	customerID := in.CustomerID
	vehicleID, err := s.resolveSaleVehicle(ctx, in.VehicleID, in.VehicleVIN)
	if err != nil {
		return nil, err
	}
	ok, err := s.customers.CustomerExists(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrCustomerNotFound
	}
	if err := s.checkRef(ctx, &in.IssueWarehouseID, s.dealerPoints.WarehouseExists, ErrWarehouseNotFound); err != nil {
		return nil, err
	}
	number, err := s.customerOrderRepo.NextOrderNumber(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	orderID := uuid.New()
	var lines []domain.PartOrderLine
	if len(in.Lines) > 0 {
		lines, err = s.buildOrderLines(ctx, orderID, in.Lines, now)
		if err != nil {
			return nil, err
		}
	}
	o := &domain.CustomerOrder{
		ID: orderID, OrderNumber: number, Status: domain.OrderStatusDraft,
		CustomerID: customerID, VehicleID: vehicleID, IssueWarehouseID: in.IssueWarehouseID,
		Notes: in.Notes, CreatedBy: in.CreatedBy, Lines: lines, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.customerOrderRepo.Create(ctx, o); err != nil {
		return nil, err
	}
	return s.GetCustomerOrder(ctx, orderID.String())
}

func (s *PartService) GetCustomerOrder(ctx context.Context, id string) (*domain.CustomerOrder, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	o, err := s.customerOrderRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return o, nil
}

func (s *PartService) ListCustomerOrders(ctx context.Context, limit, offset int32, status string) ([]*domain.CustomerOrder, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.customerOrderRepo.List(ctx, limit, offset, status)
}

func (s *PartService) UpdateCustomerOrder(ctx context.Context, id string, in domain.UpdateCustomerOrderInput) (*domain.CustomerOrder, error) {
	o, err := s.GetCustomerOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if o.Status != domain.OrderStatusDraft {
		return nil, ErrOrderNotEditable
	}
	now := time.Now().UTC()
	nextCustomer := o.CustomerID
	if in.CustomerID != nil {
		nextCustomer = *in.CustomerID
	}
	nextVehicle := o.VehicleID
	if in.ClearVehicle {
		nextVehicle = nil
	} else if in.VehicleID != nil {
		nextVehicle = in.VehicleID
	}
	vin := ""
	if in.VehicleVIN != nil {
		vin = *in.VehicleVIN
	}
	ok, err := s.customers.CustomerExists(ctx, nextCustomer)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrCustomerNotFound
	}
	vehicleID, err := s.resolveSaleVehicle(ctx, nextVehicle, vin)
	if err != nil {
		return nil, err
	}
	o.CustomerID = nextCustomer
	o.VehicleID = vehicleID
	if in.IssueWarehouseID != nil {
		if err := s.checkRef(ctx, in.IssueWarehouseID, s.dealerPoints.WarehouseExists, ErrWarehouseNotFound); err != nil {
			return nil, err
		}
		o.IssueWarehouseID = *in.IssueWarehouseID
	}
	if in.Notes != nil {
		o.Notes = *in.Notes
	}
	if in.ReplaceLines {
		o.Lines, err = s.buildOrderLines(ctx, o.ID, in.Lines, now)
		if err != nil {
			return nil, err
		}
	}
	o.UpdatedAt = now
	if err := s.customerOrderRepo.Update(ctx, o, in.ReplaceLines); err != nil {
		return nil, err
	}
	return s.GetCustomerOrder(ctx, id)
}

func (s *PartService) CancelCustomerOrder(ctx context.Context, id string) (*domain.CustomerOrder, error) {
	o, err := s.GetCustomerOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if o.Status != domain.OrderStatusDraft {
		return nil, ErrOrderNotDraft
	}
	now := time.Now().UTC()
	if err := s.customerOrderRepo.UpdateStatus(ctx, o.ID, domain.OrderStatusCancelled, nil, now); err != nil {
		return nil, err
	}
	return s.GetCustomerOrder(ctx, id)
}

func (s *PartService) CreateSaleFromCustomerOrder(ctx context.Context, orderID string, createdBy *uuid.UUID) (*domain.MovementDocument, error) {
	o, err := s.GetCustomerOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != domain.OrderStatusDraft {
		return nil, ErrOrderNotDraft
	}
	if o.FulfillmentMovementDocumentID != nil || o.FulfillmentWorkOrderID != nil {
		return nil, ErrOrderAlreadyLinked
	}
	if len(o.Lines) == 0 {
		return nil, ErrOrderNoLines
	}
	refID := o.ID
	lineInputs := make([]domain.MovementDocumentLineInput, len(o.Lines))
	for i, l := range o.Lines {
		refLine := l.ID
		lineInputs[i] = domain.MovementDocumentLineInput{
			PartID: l.PartID, WarehouseID: o.IssueWarehouseID, Quantity: l.Quantity,
			ReferenceLineID: &refLine, Notes: l.Notes, SortOrder: l.SortOrder,
		}
	}
	doc, err := s.CreateMovementDocument(ctx, domain.CreateMovementDocumentInput{
		MovementType: domain.MovementTypeSale, ReferenceType: domain.RefCustomerOrder, ReferenceID: &refID,
		CustomerID: &o.CustomerID, VehicleID: o.VehicleID,
		Notes: fmt.Sprintf("Реализация по заказу %s", o.OrderNumber), CreatedBy: createdBy, Lines: lineInputs,
	})
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := s.customerOrderRepo.UpdateStatus(ctx, o.ID, domain.OrderStatusLinked, &doc.ID, now); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *PartService) workOrderPartsFromLines(lines []domain.PartOrderLine, warehouseID uuid.UUID) []WorkOrderPartLineInput {
	out := make([]WorkOrderPartLineInput, len(lines))
	for i, l := range lines {
		out[i] = WorkOrderPartLineInput{
			PartID: l.PartID.String(), WarehouseID: warehouseID.String(),
			Quantity: strconv.Itoa(int(l.Quantity)), UnitPrice: l.UnitPrice,
			Notes: l.Notes, SortOrder: l.SortOrder,
		}
	}
	return out
}

func (s *PartService) WorkOrderNumber(ctx context.Context, id uuid.UUID) string {
	num, err := s.workOrders.GetWorkOrder(ctx, id.String())
	if err != nil {
		return ""
	}
	return num
}

type CreateWorkOrderFromSupplierOrderInput struct {
	CustomerID uuid.UUID
	VehicleID  *uuid.UUID
	VehicleVIN string
	Notes      string
}

func (s *PartService) CreateWorkOrderFromSupplierOrder(ctx context.Context, orderID string, in CreateWorkOrderFromSupplierOrderInput) (*CreatedWorkOrderRef, error) {
	o, err := s.GetSupplierOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != domain.OrderStatusDraft {
		return nil, ErrOrderNotDraft
	}
	if o.FulfillmentMovementDocumentID != nil || o.FulfillmentWorkOrderID != nil {
		return nil, ErrOrderAlreadyLinked
	}
	if len(o.Lines) == 0 {
		return nil, ErrOrderNoLines
	}
	ok, err := s.customers.CustomerExists(ctx, in.CustomerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrCustomerNotFound
	}
	vehicleID, err := s.resolveSaleVehicle(ctx, in.VehicleID, in.VehicleVIN)
	if err != nil {
		return nil, err
	}
	if vehicleID == nil {
		return nil, ErrVehicleRequiredForWO
	}
	notes := in.Notes
	if notes == "" {
		notes = fmt.Sprintf("Заказ-наряд по заказу поставщику %s", o.OrderNumber)
	}
	wo, err := s.workOrders.CreateWorkOrder(ctx, CreateWorkOrderFromOrderInput{
		CustomerID: in.CustomerID.String(), VehicleID: vehicleID.String(),
		WarehouseID: o.ReceiptWarehouseID.String(),
		SourceOrderType: domain.RefSupplierOrder, SourceOrderID: o.ID.String(),
		Notes: notes, Parts: s.workOrderPartsFromLines(o.Lines, o.ReceiptWarehouseID),
	})
	if err != nil {
		return nil, err
	}
	woID, err := uuid.Parse(wo.ID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := s.supplierOrderRepo.LinkWorkOrder(ctx, o.ID, woID, now); err != nil {
		return nil, err
	}
	return wo, nil
}

type CreateWorkOrderFromCustomerOrderInput struct {
	VehicleID  *uuid.UUID
	VehicleVIN string
	Notes      string
}

func (s *PartService) CreateWorkOrderFromCustomerOrder(ctx context.Context, orderID string, in CreateWorkOrderFromCustomerOrderInput) (*CreatedWorkOrderRef, error) {
	o, err := s.GetCustomerOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != domain.OrderStatusDraft {
		return nil, ErrOrderNotDraft
	}
	if o.FulfillmentMovementDocumentID != nil || o.FulfillmentWorkOrderID != nil {
		return nil, ErrOrderAlreadyLinked
	}
	if len(o.Lines) == 0 {
		return nil, ErrOrderNoLines
	}
	vehicleID := o.VehicleID
	if in.VehicleID != nil {
		vehicleID = in.VehicleID
	}
	vin := in.VehicleVIN
	resolvedVehicle, err := s.resolveSaleVehicle(ctx, vehicleID, vin)
	if err != nil {
		return nil, err
	}
	if resolvedVehicle == nil {
		return nil, ErrVehicleRequiredForWO
	}
	notes := in.Notes
	if notes == "" {
		notes = fmt.Sprintf("Заказ-наряд по заказу покупателя %s", o.OrderNumber)
	}
	wo, err := s.workOrders.CreateWorkOrder(ctx, CreateWorkOrderFromOrderInput{
		CustomerID: o.CustomerID.String(), VehicleID: resolvedVehicle.String(),
		WarehouseID: o.IssueWarehouseID.String(),
		SourceOrderType: domain.RefCustomerOrder, SourceOrderID: o.ID.String(),
		Notes: notes, Parts: s.workOrderPartsFromLines(o.Lines, o.IssueWarehouseID),
	})
	if err != nil {
		return nil, err
	}
	woID, err := uuid.Parse(wo.ID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := s.customerOrderRepo.LinkWorkOrder(ctx, o.ID, woID, now); err != nil {
		return nil, err
	}
	return wo, nil
}

func (s *PartService) FulfillOrderFromWorkOrder(ctx context.Context, sourceOrderType, sourceOrderID string) error {
	id, err := uuid.Parse(sourceOrderID)
	if err != nil {
		return ErrOrderNotFound
	}
	switch sourceOrderType {
	case domain.RefSupplierOrder:
		return s.supplierOrderRepo.MarkFulfilled(ctx, id)
	case domain.RefCustomerOrder:
		return s.customerOrderRepo.MarkFulfilled(ctx, id)
	default:
		return nil
	}
}

func (s *PartService) fulfillLinkedOrder(ctx context.Context, doc *domain.MovementDocument) {
	if doc.ReferenceID == nil {
		return
	}
	switch doc.ReferenceType {
	case domain.RefSupplierOrder:
		if doc.MovementType == domain.MovementTypeReceipt {
			_ = s.supplierOrderRepo.MarkFulfilled(ctx, *doc.ReferenceID)
		}
	case domain.RefCustomerOrder:
		if doc.MovementType == domain.MovementTypeSale {
			_ = s.customerOrderRepo.MarkFulfilled(ctx, *doc.ReferenceID)
		}
	}
}

func (s *PartService) GetSupplierOrderNumber(ctx context.Context, id uuid.UUID) string {
	o, err := s.supplierOrderRepo.GetByID(ctx, id)
	if err != nil || o == nil {
		return ""
	}
	return o.OrderNumber
}

func (s *PartService) GetCustomerOrderNumber(ctx context.Context, id uuid.UUID) string {
	o, err := s.customerOrderRepo.GetByID(ctx, id)
	if err != nil || o == nil {
		return ""
	}
	return o.OrderNumber
}
