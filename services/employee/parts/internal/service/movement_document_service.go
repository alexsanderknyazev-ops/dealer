package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

var (
	ErrMovementDocumentNotFound      = errors.New("movement document not found")
	ErrMovementDocumentNotDraft      = errors.New("movement document is not draft")
	ErrMovementDocumentNotInProgress = errors.New("movement document is not in progress")
	ErrMovementDocumentNotEditable   = errors.New("movement document cannot be edited")
	ErrMovementDocumentNoLines       = errors.New("movement document has no lines")
	ErrInsufficientStock             = errors.New("insufficient stock")
	ErrParentNotClosed               = errors.New("parent movement document is not closed")
	ErrParentNotExtractable          = errors.New("parent movement document does not support extraction")
	ErrOpenExtractionExists          = errors.New("open production extraction already exists")
	ErrNothingToExtract              = errors.New("nothing left to extract from production")
	ErrExtractionExceedsBalance      = errors.New("extraction quantity exceeds remaining balance")
	ErrDestinationRequired           = errors.New("destination warehouse required")
	ErrSameSourceDestination         = errors.New("source and destination warehouse must differ")
	ErrCustomerRequired              = errors.New("customer required for goods sale")
	ErrCustomerNotFound              = errors.New("customer not found")
	ErrVehicleNotFound               = errors.New("vehicle not found")
	ErrSupplierRequired              = errors.New("supplier required for goods receipt")
	ErrSupplierNotFound              = errors.New("supplier not found")
	ErrReceiptWarehouseRequired      = errors.New("receipt warehouse required")
	ErrUnitCostRequired              = errors.New("unit cost required for receipt line")
)

type stockMovementRepository interface {
	Create(ctx context.Context, m *domain.StockMovement) error
}

type movementDocumentRepository interface {
	NextDocumentNumber(ctx context.Context) (string, error)
	Create(ctx context.Context, doc *domain.MovementDocument) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.MovementDocument, error)
	List(ctx context.Context, limit, offset int32, status, referenceType, referenceID string) ([]*domain.MovementDocument, int32, error)
	UpdateStatus(ctx context.Context, doc *domain.MovementDocument) error
	Update(ctx context.Context, doc *domain.MovementDocument, replaceLines bool) error
	ExtractedQuantityByParentLine(ctx context.Context, parentDocumentID, parentLineID uuid.UUID) (int32, error)
	HasOpenExtractionForParent(ctx context.Context, parentDocumentID uuid.UUID) (bool, error)
}

type WorkOrderPartLineInput struct {
	PartID, WarehouseID, Quantity, UnitPrice, Notes string
	SortOrder                                     int32
}

type CreateWorkOrderFromOrderInput struct {
	CustomerID, VehicleID, WarehouseID string
	SourceOrderType, SourceOrderID     string
	Notes                              string
	Parts                              []WorkOrderPartLineInput
}

type CreatedWorkOrderRef struct {
	ID, OrderNumber string
}

type WorkOrdersNotifier interface {
	ApplyMovementDocument(ctx context.Context, workOrderID, documentID, status string) error
	CreateWorkOrder(ctx context.Context, in CreateWorkOrderFromOrderInput) (*CreatedWorkOrderRef, error)
	GetWorkOrder(ctx context.Context, id string) (orderNumber string, err error)
}

type noopWorkOrdersNotifier struct{}

func (noopWorkOrdersNotifier) ApplyMovementDocument(context.Context, string, string, string) error {
	return nil
}

func (noopWorkOrdersNotifier) CreateWorkOrder(context.Context, CreateWorkOrderFromOrderInput) (*CreatedWorkOrderRef, error) {
	return nil, errors.New("workorders service unavailable")
}

func (noopWorkOrdersNotifier) GetWorkOrder(context.Context, string) (string, error) {
	return "", errors.New("workorders service unavailable")
}

func parseUnitCost(raw string, required bool) (string, error) {
	s := strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	if s == "" {
		if required {
			return "", ErrUnitCostRequired
		}
		return "0", nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v < 0 {
		return "", ErrUnitCostRequired
	}
	if required && v <= 0 {
		return "", ErrUnitCostRequired
	}
	return s, nil
}

func (s *PartService) validateReceiptDocument(
	ctx context.Context,
	movementType string,
	supplierID *uuid.UUID,
	receiptWarehouseID *uuid.UUID,
) (*uuid.UUID, *uuid.UUID, error) {
	if movementType != domain.MovementTypeReceipt {
		return nil, nil, nil
	}
	if supplierID == nil {
		return nil, nil, ErrSupplierRequired
	}
	if s.suppliers == nil {
		return supplierID, receiptWarehouseID, nil
	}
	ok, err := s.suppliers.Exists(ctx, *supplierID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, ErrSupplierNotFound
	}
	if receiptWarehouseID == nil {
		return nil, nil, ErrReceiptWarehouseRequired
	}
	if err := s.checkRef(ctx, receiptWarehouseID, s.dealerPoints.WarehouseExists, ErrWarehouseNotFound); err != nil {
		return nil, nil, err
	}
	return supplierID, receiptWarehouseID, nil
}

func (s *PartService) applyDocumentHeaderFields(
	doc *domain.MovementDocument,
	movementType string,
	customerID, vehicleID, supplierID, receiptWarehouseID *uuid.UUID,
) {
	doc.CustomerID = nil
	doc.VehicleID = nil
	doc.SupplierID = nil
	doc.ReceiptWarehouseID = nil
	switch movementType {
	case domain.MovementTypeSale:
		doc.CustomerID = customerID
		doc.VehicleID = vehicleID
	case domain.MovementTypeReceipt:
		doc.SupplierID = supplierID
		doc.ReceiptWarehouseID = receiptWarehouseID
	}
}

func (s *PartService) buildMovementLines(
	ctx context.Context,
	docID uuid.UUID,
	movementType string,
	receiptWarehouseID *uuid.UUID,
	inputs []domain.MovementDocumentLineInput,
	now time.Time,
) ([]domain.MovementDocumentLine, error) {
	lines := make([]domain.MovementDocumentLine, 0, len(inputs))
	for _, lineIn := range inputs {
		if lineIn.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for part %s", lineIn.PartID)
		}
		if _, err := s.repo.GetByID(ctx, lineIn.PartID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		warehouseID := lineIn.WarehouseID
		if movementType == domain.MovementTypeReceipt {
			if receiptWarehouseID == nil {
				return nil, ErrReceiptWarehouseRequired
			}
			warehouseID = *receiptWarehouseID
		} else if err := s.checkRef(ctx, &lineIn.WarehouseID, s.dealerPoints.WarehouseExists, ErrWarehouseNotFound); err != nil {
			return nil, err
		}
		unitCost, err := parseUnitCost(lineIn.UnitCost, movementType == domain.MovementTypeReceipt)
		if err != nil {
			return nil, err
		}
		var destID *uuid.UUID
		if movementType == domain.MovementTypeTransfer {
			if lineIn.DestinationWarehouseID == nil {
				return nil, ErrDestinationRequired
			}
			if *lineIn.DestinationWarehouseID == lineIn.WarehouseID {
				return nil, ErrSameSourceDestination
			}
			if err := s.checkRef(ctx, lineIn.DestinationWarehouseID, s.dealerPoints.WarehouseExists, ErrWarehouseNotFound); err != nil {
				return nil, err
			}
			destID = lineIn.DestinationWarehouseID
		}
		lines = append(lines, domain.MovementDocumentLine{
			ID:                     uuid.New(),
			DocumentID:             docID,
			PartID:                 lineIn.PartID,
			WarehouseID:            warehouseID,
			DestinationWarehouseID: destID,
			Quantity:               lineIn.Quantity,
			UnitCost:               unitCost,
			ReferenceLineID:        lineIn.ReferenceLineID,
			Notes:                  lineIn.Notes,
			SortOrder:              lineIn.SortOrder,
			CreatedAt:              now,
		})
	}
	return lines, nil
}

func (s *PartService) resolveSaleVehicle(ctx context.Context, vehicleID *uuid.UUID, vehicleVIN string) (*uuid.UUID, error) {
	if vehicleID != nil {
		ok, err := s.customers.VehicleExists(ctx, *vehicleID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrVehicleNotFound
		}
		return vehicleID, nil
	}
	vin := strings.TrimSpace(vehicleVIN)
	if vin == "" {
		return nil, nil
	}
	id, err := s.customers.LookupVehicleIDByVIN(ctx, vin)
	if err != nil {
		return nil, err
	}
	if id == nil {
		return nil, ErrVehicleNotFound
	}
	return id, nil
}

func (s *PartService) validateSaleDocument(
	ctx context.Context,
	movementType string,
	customerID *uuid.UUID,
	vehicleID *uuid.UUID,
	vehicleVIN string,
) (*uuid.UUID, *uuid.UUID, error) {
	if movementType != domain.MovementTypeSale {
		return nil, nil, nil
	}
	if customerID == nil {
		return nil, nil, ErrCustomerRequired
	}
	ok, err := s.customers.CustomerExists(ctx, *customerID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, ErrCustomerNotFound
	}
	resolvedVehicle, err := s.resolveSaleVehicle(ctx, vehicleID, vehicleVIN)
	if err != nil {
		return nil, nil, err
	}
	return customerID, resolvedVehicle, nil
}

func (s *PartService) CreateMovementDocument(ctx context.Context, in domain.CreateMovementDocumentInput) (*domain.MovementDocument, error) {
	if in.MovementType == domain.MovementTypeFromProduction {
		return nil, fmt.Errorf("use CreateProductionExtraction for from_production documents")
	}
	customerID, vehicleID, err := s.validateSaleDocument(ctx, in.MovementType, in.CustomerID, in.VehicleID, in.VehicleVIN)
	if err != nil {
		return nil, err
	}
	supplierID, receiptWarehouseID, err := s.validateReceiptDocument(ctx, in.MovementType, in.SupplierID, in.ReceiptWarehouseID)
	if err != nil {
		return nil, err
	}
	number, err := s.movementDocRepo.NextDocumentNumber(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	docID := uuid.New()
	var lines []domain.MovementDocumentLine
	if len(in.Lines) > 0 {
		lines, err = s.buildMovementLines(ctx, docID, in.MovementType, receiptWarehouseID, in.Lines, now)
		if err != nil {
			return nil, err
		}
	}
	doc := &domain.MovementDocument{
		ID:             docID,
		DocumentNumber: number,
		Status:         domain.DocumentStatusDraft,
		MovementType:   in.MovementType,
		ReferenceType:  in.ReferenceType,
		ReferenceID:    in.ReferenceID,
		Notes:          in.Notes,
		CreatedBy:      in.CreatedBy,
		Lines:          lines,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.applyDocumentHeaderFields(doc, in.MovementType, customerID, vehicleID, supplierID, receiptWarehouseID)
	if err := s.movementDocRepo.Create(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *PartService) GetMovementDocument(ctx context.Context, id string) (*domain.MovementDocument, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrMovementDocumentNotFound
	}
	doc, err := s.movementDocRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMovementDocumentNotFound
		}
		return nil, err
	}
	return doc, nil
}

func (s *PartService) ListMovementDocuments(ctx context.Context, limit, offset int32, status, referenceType, referenceID string) ([]*domain.MovementDocument, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.movementDocRepo.List(ctx, limit, offset, status, referenceType, referenceID)
}

func (s *PartService) UpdateMovementDocument(ctx context.Context, id string, in domain.UpdateMovementDocumentInput) (*domain.MovementDocument, error) {
	doc, err := s.GetMovementDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocumentStatusDraft && doc.Status != domain.DocumentStatusInProgress {
		return nil, ErrMovementDocumentNotEditable
	}
	now := time.Now().UTC()
	if in.MovementType != nil {
		doc.MovementType = *in.MovementType
	}
	nextCustomer := doc.CustomerID
	if in.CustomerID != nil {
		nextCustomer = in.CustomerID
	}
	nextVehicle := doc.VehicleID
	if in.ClearVehicle {
		nextVehicle = nil
	} else if in.VehicleID != nil {
		nextVehicle = in.VehicleID
	}
	vehicleVIN := ""
	if in.VehicleVIN != nil {
		vehicleVIN = *in.VehicleVIN
	}
	customerID, vehicleID, err := s.validateSaleDocument(ctx, doc.MovementType, nextCustomer, nextVehicle, vehicleVIN)
	if err != nil {
		return nil, err
	}
	nextSupplier := doc.SupplierID
	if in.SupplierID != nil {
		nextSupplier = in.SupplierID
	}
	nextReceiptWh := doc.ReceiptWarehouseID
	if in.ReceiptWarehouseID != nil {
		nextReceiptWh = in.ReceiptWarehouseID
	}
	supplierID, receiptWarehouseID, err := s.validateReceiptDocument(ctx, doc.MovementType, nextSupplier, nextReceiptWh)
	if err != nil {
		return nil, err
	}
	s.applyDocumentHeaderFields(doc, doc.MovementType, customerID, vehicleID, supplierID, receiptWarehouseID)
	if in.Notes != nil {
		doc.Notes = *in.Notes
	}
	if in.ReplaceLines {
		doc.Lines, err = s.buildMovementLines(ctx, doc.ID, doc.MovementType, receiptWarehouseID, in.Lines, now)
		if err != nil {
			return nil, err
		}
	}
	doc.UpdatedAt = now
	if err := s.movementDocRepo.Update(ctx, doc, in.ReplaceLines); err != nil {
		return nil, err
	}
	return s.GetMovementDocument(ctx, id)
}

func (s *PartService) StartMovementDocument(ctx context.Context, id string) (*domain.MovementDocument, error) {
	doc, err := s.GetMovementDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocumentStatusDraft {
		return nil, ErrMovementDocumentNotDraft
	}
	if len(doc.Lines) == 0 {
		return nil, ErrMovementDocumentNoLines
	}
	if doc.MovementType == domain.MovementTypeSale {
		if _, _, err := s.validateSaleDocument(ctx, doc.MovementType, doc.CustomerID, doc.VehicleID, ""); err != nil {
			return nil, err
		}
	}
	if doc.MovementType == domain.MovementTypeReceipt {
		if _, _, err := s.validateReceiptDocument(ctx, doc.MovementType, doc.SupplierID, doc.ReceiptWarehouseID); err != nil {
			return nil, err
		}
	}
	doc.Status = domain.DocumentStatusInProgress
	doc.UpdatedAt = time.Now().UTC()
	if err := s.movementDocRepo.UpdateStatus(ctx, doc); err != nil {
		return nil, err
	}
	if doc.ReferenceType == domain.RefWorkOrder && doc.ReferenceID != nil {
		_ = s.workOrders.ApplyMovementDocument(ctx, doc.ReferenceID.String(), doc.ID.String(), doc.Status)
	}
	return s.GetMovementDocument(ctx, id)
}

func (s *PartService) recordStockMovement(
	ctx context.Context,
	doc *domain.MovementDocument,
	line domain.MovementDocumentLine,
	warehouseID uuid.UUID,
	quantity int32,
	refLine *uuid.UUID,
	closedBy *uuid.UUID,
	now time.Time,
) error {
	movement := domain.StockMovement{
		ID:                 uuid.New(),
		PartID:             line.PartID,
		WarehouseID:        warehouseID,
		Quantity:           quantity,
		MovementType:       doc.MovementType,
		ReferenceType:      doc.ReferenceType,
		ReferenceID:        doc.ReferenceID,
		ReferenceLineID:    refLine,
		MovementDocumentID: &doc.ID,
		Notes:              line.Notes,
		CreatedBy:          closedBy,
		CreatedAt:          now,
	}
	return s.movementRepo.Create(ctx, &movement)
}

func (s *PartService) CloseMovementDocument(ctx context.Context, id string, closedBy *uuid.UUID) (*domain.MovementDocument, error) {
	doc, err := s.GetMovementDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocumentStatusInProgress {
		return nil, ErrMovementDocumentNotInProgress
	}
	if len(doc.Lines) == 0 {
		return nil, ErrMovementDocumentNoLines
	}
	if doc.MovementType == domain.MovementTypeFromProduction {
		if err := s.validateExtractionClose(ctx, doc); err != nil {
			return nil, err
		}
	}
	now := time.Now().UTC()
	for _, line := range doc.Lines {
		refLine := line.ReferenceLineID
		switch doc.MovementType {
		case domain.MovementTypeFromProduction:
			if err := s.closeExtractionLine(ctx, doc, line, refLine, closedBy, now); err != nil {
				return nil, err
			}
		case domain.MovementTypeReceipt:
			if err := s.stockRepo.Add(ctx, line.PartID, line.WarehouseID, line.Quantity); err != nil {
				return nil, err
			}
			if err := s.recordStockMovement(ctx, doc, line, line.WarehouseID, line.Quantity, refLine, closedBy, now); err != nil {
				return nil, err
			}
		case domain.MovementTypeTransfer:
			if _, err := s.stockRepo.Deduct(ctx, line.PartID, line.WarehouseID, line.Quantity); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					qty, _ := s.stockRepo.GetQuantity(ctx, line.PartID, line.WarehouseID)
					return nil, fmt.Errorf("%w: part %s on warehouse %s (have %d, need %d)",
						ErrInsufficientStock, line.PartID, line.WarehouseID, qty, line.Quantity)
				}
				return nil, err
			}
			if err := s.recordStockMovement(ctx, doc, line, line.WarehouseID, -line.Quantity, refLine, closedBy, now); err != nil {
				return nil, err
			}
			if line.DestinationWarehouseID == nil {
				return nil, ErrDestinationRequired
			}
			if err := s.stockRepo.Add(ctx, line.PartID, *line.DestinationWarehouseID, line.Quantity); err != nil {
				return nil, err
			}
			if err := s.recordStockMovement(ctx, doc, line, *line.DestinationWarehouseID, line.Quantity, refLine, closedBy, now); err != nil {
				return nil, err
			}
		default:
			if _, err := s.stockRepo.Deduct(ctx, line.PartID, line.WarehouseID, line.Quantity); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					qty, _ := s.stockRepo.GetQuantity(ctx, line.PartID, line.WarehouseID)
					return nil, fmt.Errorf("%w: part %s on warehouse %s (have %d, need %d)",
						ErrInsufficientStock, line.PartID, line.WarehouseID, qty, line.Quantity)
				}
				return nil, err
			}
			if err := s.recordStockMovement(ctx, doc, line, line.WarehouseID, -line.Quantity, refLine, closedBy, now); err != nil {
				return nil, err
			}
		}
	}
	doc.Status = domain.DocumentStatusClosed
	doc.ConfirmedBy = closedBy
	doc.ConfirmedAt = &now
	doc.UpdatedAt = now
	if err := s.movementDocRepo.UpdateStatus(ctx, doc); err != nil {
		return nil, err
	}
	if doc.ReferenceType == domain.RefWorkOrder && doc.ReferenceID != nil {
		if err := s.workOrders.ApplyMovementDocument(ctx, doc.ReferenceID.String(), doc.ID.String(), doc.Status); err != nil {
			return nil, fmt.Errorf("notify work order: %w", err)
		}
	}
	s.fulfillLinkedOrder(ctx, doc)
	return s.GetMovementDocument(ctx, id)
}

// ConfirmMovementDocument — совместимость: закрывает документ (списание при closed).
func (s *PartService) ConfirmMovementDocument(ctx context.Context, id string, confirmedBy *uuid.UUID) (*domain.MovementDocument, error) {
	return s.CloseMovementDocument(ctx, id, confirmedBy)
}

func (s *PartService) closeExtractionLine(
	ctx context.Context,
	doc *domain.MovementDocument,
	line domain.MovementDocumentLine,
	refLine *uuid.UUID,
	closedBy *uuid.UUID,
	now time.Time,
) error {
	returnWh := line.WarehouseID
	// Для извлечения после transfer товар на складе-получателе — списываем оттуда.
	if line.DestinationWarehouseID != nil {
		fromWh := *line.DestinationWarehouseID
		if _, err := s.stockRepo.Deduct(ctx, line.PartID, fromWh, line.Quantity); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				qty, _ := s.stockRepo.GetQuantity(ctx, line.PartID, fromWh)
				return fmt.Errorf("%w: part %s on warehouse %s (have %d, need %d)",
					ErrInsufficientStock, line.PartID, fromWh, qty, line.Quantity)
			}
			return err
		}
		if err := s.recordStockMovement(ctx, doc, line, fromWh, -line.Quantity, refLine, closedBy, now); err != nil {
			return err
		}
	}
	if err := s.stockRepo.Add(ctx, line.PartID, returnWh, line.Quantity); err != nil {
		return err
	}
	return s.recordStockMovement(ctx, doc, line, returnWh, line.Quantity, refLine, closedBy, now)
}

func (s *PartService) validateExtractionClose(ctx context.Context, doc *domain.MovementDocument) error {
	if doc.ParentDocumentID == nil {
		return ErrParentNotExtractable
	}
	parent, err := s.movementDocRepo.GetByID(ctx, *doc.ParentDocumentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMovementDocumentNotFound
		}
		return err
	}
	if parent.Status != domain.DocumentStatusClosed || !domain.MovementTypeSupportsExtraction(parent.MovementType) {
		return ErrParentNotExtractable
	}
	for _, line := range doc.Lines {
		if line.ReferenceLineID == nil {
			return ErrExtractionExceedsBalance
		}
		extracted, err := s.movementDocRepo.ExtractedQuantityByParentLine(ctx, parent.ID, *line.ReferenceLineID)
		if err != nil {
			return err
		}
		var parentQty int32
		for _, pl := range parent.Lines {
			if pl.ID == *line.ReferenceLineID {
				parentQty = pl.Quantity
				break
			}
		}
		if parentQty == 0 || line.Quantity > parentQty-extracted {
			return fmt.Errorf("%w: line %s", ErrExtractionExceedsBalance, line.ID)
		}
	}
	return nil
}

func (s *PartService) buildExtractionLine(parent *domain.MovementDocument, pl domain.MovementDocumentLine, docID uuid.UUID, remaining int32, sortOrder int32, now time.Time) domain.MovementDocumentLine {
	refLineID := pl.ID
	line := domain.MovementDocumentLine{
		ID:              uuid.New(),
		DocumentID:      docID,
		PartID:          pl.PartID,
		WarehouseID:     pl.WarehouseID,
		Quantity:        remaining,
		ReferenceLineID: &refLineID,
		Notes:           fmt.Sprintf("извлечение из %s", parent.DocumentNumber),
		SortOrder:       sortOrder,
		CreatedAt:       now,
	}
	if parent.MovementType == domain.MovementTypeTransfer && pl.DestinationWarehouseID != nil {
		dest := *pl.DestinationWarehouseID
		line.DestinationWarehouseID = &dest
	}
	return line
}

func (s *PartService) CreateProductionExtraction(ctx context.Context, parentID string, createdBy *uuid.UUID) (*domain.MovementDocument, error) {
	parentUID, err := uuid.Parse(parentID)
	if err != nil {
		return nil, ErrMovementDocumentNotFound
	}
	parent, err := s.movementDocRepo.GetByID(ctx, parentUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMovementDocumentNotFound
		}
		return nil, err
	}
	if parent.Status != domain.DocumentStatusClosed {
		return nil, ErrParentNotClosed
	}
	if parent.MovementType == domain.MovementTypeFromProduction || !domain.MovementTypeSupportsExtraction(parent.MovementType) {
		return nil, ErrParentNotExtractable
	}
	open, err := s.movementDocRepo.HasOpenExtractionForParent(ctx, parent.ID)
	if err != nil {
		return nil, err
	}
	if open {
		return nil, ErrOpenExtractionExists
	}
	now := time.Now().UTC()
	docID := uuid.New()
	var lines []domain.MovementDocumentLine
	for i, pl := range parent.Lines {
		extracted, err := s.movementDocRepo.ExtractedQuantityByParentLine(ctx, parent.ID, pl.ID)
		if err != nil {
			return nil, err
		}
		remaining := pl.Quantity - extracted
		if remaining <= 0 {
			continue
		}
		lines = append(lines, s.buildExtractionLine(parent, pl, docID, remaining, int32(i+1), now))
	}
	if len(lines) == 0 {
		return nil, ErrNothingToExtract
	}
	number, err := s.movementDocRepo.NextDocumentNumber(ctx)
	if err != nil {
		return nil, err
	}
	parentRef := parent.ID
	doc := &domain.MovementDocument{
		ID:               docID,
		DocumentNumber:   number,
		Status:           domain.DocumentStatusDraft,
		MovementType:     domain.MovementTypeFromProduction,
		ReferenceType:    domain.RefMovementDocument,
		ReferenceID:      &parentRef,
		ParentDocumentID: &parentRef,
		Notes:            fmt.Sprintf("Извлечение по документу %s", parent.DocumentNumber),
		CreatedBy:        createdBy,
		Lines:            lines,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.movementDocRepo.Create(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *PartService) CancelMovementDocument(ctx context.Context, id string) (*domain.MovementDocument, error) {
	doc, err := s.GetMovementDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocumentStatusDraft && doc.Status != domain.DocumentStatusInProgress {
		return nil, ErrMovementDocumentNotDraft
	}
	now := time.Now().UTC()
	doc.Status = domain.DocumentStatusCancelled
	doc.UpdatedAt = now
	if err := s.movementDocRepo.UpdateStatus(ctx, doc); err != nil {
		return nil, err
	}
	if doc.ReferenceType == domain.RefWorkOrder && doc.ReferenceID != nil {
		_ = s.workOrders.ApplyMovementDocument(ctx, doc.ReferenceID.String(), doc.ID.String(), doc.Status)
	}
	return doc, nil
}
