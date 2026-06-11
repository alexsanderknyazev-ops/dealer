package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/dealer/dealer/services/workorders/internal/domain"
)

var (
	ErrNotFound            = errors.New("work order not found")
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrVehicleNotFound     = errors.New("vehicle not found")
	ErrDealerPointNotFound = errors.New("dealer point not found")
	ErrWarehouseNotFound   = errors.New("warehouse not found")
	ErrPartNotFound        = errors.New("part not found")
	ErrWorkNotFound        = errors.New("work not found")
	ErrEmployeeNotFound    = errors.New("employee not found")
	ErrNoPartsToIssue         = errors.New("no parts to issue")
	ErrPartsAlreadyIssued     = errors.New("all parts already issued")
	ErrMovementDocumentExists = errors.New("movement document already exists")
)

type LaborInput struct {
	WorkID      string
	Description string
	Quantity    string
	UnitPrice   string
	ExecutorID  string
	SortOrder   int32
}

type PartInput struct {
	PartID      string
	WarehouseID string
	Description string
	Quantity    string
	UnitPrice   string
	SortOrder   int32
}

type CreateInput struct {
	CustomerID, VehicleID, DealerPointID, WarehouseID, RepairType, Status string
	ServiceAdvisorID, Complaint, Diagnosis, Notes                         string
	MileageKm                                                           int64
	OpenedAt                                                            int64
	Labor                                                               []LaborInput
	Parts                                                               []PartInput
}

type UpdateInput struct {
	CustomerID, VehicleID, DealerPointID, WarehouseID *string
	RepairType, Status, ServiceAdvisorID                *string
	Complaint, Diagnosis, Notes                         *string
	MileageKm, OpenedAt, ClosedAt                       *int64
	Labor                                               []LaborInput
	Parts                                               []PartInput
	ReplaceLines                                        bool
}

type workOrderRepository interface {
	NextOrderNumber(ctx context.Context) (string, error)
	Create(ctx context.Context, wo *domain.WorkOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkOrder, error)
	List(ctx context.Context, limit, offset int32, status, repairType, customerID, vehicleID string) ([]*domain.WorkOrder, int32, error)
	Update(ctx context.Context, wo *domain.WorkOrder, replaceLines bool) error
	MarkPartsIssued(ctx context.Context, workOrderID uuid.UUID, lineIDs []uuid.UUID, issuedAt time.Time) error
	SetMovementDocument(ctx context.Context, workOrderID, documentID uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ReferenceChecker interface {
	CustomerExists(ctx context.Context, id uuid.UUID) (bool, error)
	VehicleExists(ctx context.Context, id uuid.UUID) (bool, error)
	DealerPointExists(ctx context.Context, id uuid.UUID) (bool, error)
	WarehouseExists(ctx context.Context, id uuid.UUID) (bool, error)
	PartExists(ctx context.Context, id uuid.UUID) (bool, error)
	WorkExists(ctx context.Context, id uuid.UUID) (bool, error)
	ResolveWork(ctx context.Context, id uuid.UUID) (name, laborHours, unitPrice string, err error)
	EmployeeExists(ctx context.Context, id uuid.UUID) (bool, error)
	EmployeeFullName(ctx context.Context, id uuid.UUID) string
	CreateMovementDocument(ctx context.Context, workOrderID uuid.UUID, orderNumber string, lines []domain.MovementDocumentLineInput, createdBy string) (string, error)
}

type noopReferenceChecker struct{}

func (noopReferenceChecker) CustomerExists(context.Context, uuid.UUID) (bool, error) { return true, nil }
func (noopReferenceChecker) VehicleExists(context.Context, uuid.UUID) (bool, error)  { return true, nil }
func (noopReferenceChecker) DealerPointExists(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}
func (noopReferenceChecker) WarehouseExists(context.Context, uuid.UUID) (bool, error) { return true, nil }
func (noopReferenceChecker) PartExists(context.Context, uuid.UUID) (bool, error)     { return true, nil }
func (noopReferenceChecker) WorkExists(context.Context, uuid.UUID) (bool, error)      { return true, nil }
func (noopReferenceChecker) ResolveWork(context.Context, uuid.UUID) (string, string, string, error) {
	return "", "", "", errors.New("works service unavailable")
}
func (noopReferenceChecker) EmployeeExists(context.Context, uuid.UUID) (bool, error) { return true, nil }
func (noopReferenceChecker) EmployeeFullName(context.Context, uuid.UUID) string      { return "" }
func (noopReferenceChecker) CreateMovementDocument(context.Context, uuid.UUID, string, []domain.MovementDocumentLineInput, string) (string, error) {
	return "", errors.New("parts service unavailable")
}

type WorkOrderService struct {
	repo workOrderRepository
	refs ReferenceChecker
}

func NewWorkOrderService(repo workOrderRepository, refs ReferenceChecker) *WorkOrderService {
	if refs == nil {
		refs = noopReferenceChecker{}
	}
	return &WorkOrderService{repo: repo, refs: refs}
}

func parseOptionalUUID(s string) *uuid.UUID {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func unixToTime(v int64) *time.Time {
	if v <= 0 {
		return nil
	}
	t := time.Unix(v, 0).UTC()
	return &t
}

func (s *WorkOrderService) validateRefs(ctx context.Context, customerID, vehicleID uuid.UUID, dealerPointID, warehouseID, serviceAdvisorID *uuid.UUID, labor []LaborInput, parts []PartInput) error {
	ok, err := s.refs.CustomerExists(ctx, customerID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCustomerNotFound
	}
	ok, err = s.refs.VehicleExists(ctx, vehicleID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrVehicleNotFound
	}
	if dealerPointID != nil {
		ok, err = s.refs.DealerPointExists(ctx, *dealerPointID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrDealerPointNotFound
		}
	}
	if warehouseID != nil {
		ok, err = s.refs.WarehouseExists(ctx, *warehouseID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrWarehouseNotFound
		}
	}
	if serviceAdvisorID != nil {
		ok, err := s.refs.EmployeeExists(ctx, *serviceAdvisorID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrEmployeeNotFound
		}
	}
	for _, l := range labor {
		wid, err := uuid.Parse(strings.TrimSpace(l.WorkID))
		if err != nil {
			return ErrWorkNotFound
		}
		ok, err := s.refs.WorkExists(ctx, wid)
		if err != nil {
			return err
		}
		if !ok {
			return ErrWorkNotFound
		}
		if eid := parseOptionalUUID(l.ExecutorID); eid != nil {
			ok, err = s.refs.EmployeeExists(ctx, *eid)
			if err != nil {
				return err
			}
			if !ok {
				return ErrEmployeeNotFound
			}
		}
	}
	for _, p := range parts {
		pid, err := uuid.Parse(p.PartID)
		if err != nil {
			return ErrPartNotFound
		}
		ok, err = s.refs.PartExists(ctx, pid)
		if err != nil {
			return err
		}
		if !ok {
			return ErrPartNotFound
		}
		if p.WarehouseID != "" {
			wid, err := uuid.Parse(p.WarehouseID)
			if err != nil {
				return ErrWarehouseNotFound
			}
			ok, err = s.refs.WarehouseExists(ctx, wid)
			if err != nil {
				return err
			}
			if !ok {
				return ErrWarehouseNotFound
			}
		}
	}
	return nil
}

func (s *WorkOrderService) buildLaborRows(ctx context.Context, workOrderID uuid.UUID, now time.Time, inputs []LaborInput) ([]domain.WorkOrderLabor, error) {
	out := make([]domain.WorkOrderLabor, 0, len(inputs))
	for _, in := range inputs {
		workID, err := uuid.Parse(strings.TrimSpace(in.WorkID))
		if err != nil {
			return nil, ErrWorkNotFound
		}
		desc := strings.TrimSpace(in.Description)
		qty := strings.TrimSpace(in.Quantity)
		price := strings.TrimSpace(in.UnitPrice)
		if desc == "" || qty == "" || price == "" {
			name, laborHours, unitPrice, err := s.refs.ResolveWork(ctx, workID)
			if err != nil {
				return nil, ErrWorkNotFound
			}
			if desc == "" {
				desc = name
			}
			if qty == "" {
				qty = laborHours
			}
			if price == "" {
				price = unitPrice
			}
		}
		if qty == "" {
			qty = "1"
		}
		if price == "" {
			price = "0"
		}
		wid := workID
		out = append(out, domain.WorkOrderLabor{
			ID:          uuid.New(),
			WorkOrderID: workOrderID,
			WorkID:      &wid,
			Description: desc,
			Quantity:    qty,
			UnitPrice:   price,
			Amount:      multiplyAmount(qty, price),
			ExecutorID:  parseOptionalUUID(in.ExecutorID),
			SortOrder:   in.SortOrder,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	return out, nil
}

func buildPartRows(workOrderID uuid.UUID, now time.Time, inputs []PartInput, defaultWarehouse *uuid.UUID) ([]domain.WorkOrderPart, error) {
	out := make([]domain.WorkOrderPart, 0, len(inputs))
	for _, in := range inputs {
		partID, err := uuid.Parse(in.PartID)
		if err != nil {
			return nil, ErrPartNotFound
		}
		warehouseID := parseOptionalUUID(in.WarehouseID)
		if warehouseID == nil {
			warehouseID = defaultWarehouse
		}
		if warehouseID == nil {
			return nil, fmt.Errorf("warehouse_id required for part %s", in.PartID)
		}
		qty := strings.TrimSpace(in.Quantity)
		if qty == "" {
			qty = "1"
		}
		price := strings.TrimSpace(in.UnitPrice)
		if price == "" {
			price = "0"
		}
		out = append(out, domain.WorkOrderPart{
			ID:          uuid.New(),
			WorkOrderID: workOrderID,
			PartID:      partID,
			WarehouseID: *warehouseID,
			Description: in.Description,
			Quantity:    qty,
			UnitPrice:   price,
			Amount:      multiplyAmount(qty, price),
			Issued:      false,
			SortOrder:   in.SortOrder,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	return out, nil
}

func recalcCosts(wo *domain.WorkOrder) {
	laborAmounts := make([]string, len(wo.Labor))
	for i, l := range wo.Labor {
		laborAmounts[i] = l.Amount
	}
	partAmounts := make([]string, len(wo.Parts))
	for i, p := range wo.Parts {
		partAmounts[i] = p.Amount
	}
	wo.LaborCost = sumAmounts(laborAmounts...)
	wo.PartsCost = sumAmounts(partAmounts...)
	wo.TotalCost = sumAmounts(wo.LaborCost, wo.PartsCost)
}

func (s *WorkOrderService) Create(ctx context.Context, in CreateInput) (*domain.WorkOrder, error) {
	customerID, err := uuid.Parse(in.CustomerID)
	if err != nil {
		return nil, errors.New("invalid customer_id")
	}
	vehicleID, err := uuid.Parse(in.VehicleID)
	if err != nil {
		return nil, errors.New("invalid vehicle_id")
	}
	repairType := strings.TrimSpace(in.RepairType)
	if repairType == "" {
		repairType = domain.RepairCommercial
	}
	statusVal := strings.TrimSpace(in.Status)
	if statusVal == "" {
		statusVal = domain.StatusDraft
	}
	dealerPointID := parseOptionalUUID(in.DealerPointID)
	warehouseID := parseOptionalUUID(in.WarehouseID)
	if err := s.validateRefs(ctx, customerID, vehicleID, dealerPointID, warehouseID, parseOptionalUUID(in.ServiceAdvisorID), in.Labor, in.Parts); err != nil {
		return nil, err
	}
	orderNumber, err := s.repo.NextOrderNumber(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	woID := uuid.New()
	labor, err := s.buildLaborRows(ctx, woID, now, in.Labor)
	if err != nil {
		return nil, err
	}
	parts, err := buildPartRows(woID, now, in.Parts, warehouseID)
	if err != nil {
		return nil, err
	}
	wo := &domain.WorkOrder{
		ID:               woID,
		OrderNumber:      orderNumber,
		CustomerID:       customerID,
		VehicleID:        vehicleID,
		DealerPointID:    dealerPointID,
		WarehouseID:      warehouseID,
		RepairType:       repairType,
		Status:           statusVal,
		ServiceAdvisorID: parseOptionalUUID(in.ServiceAdvisorID),
		Complaint:        in.Complaint,
		Diagnosis:        in.Diagnosis,
		MileageKm:        in.MileageKm,
		OpenedAt:         unixToTime(in.OpenedAt),
		Notes:            in.Notes,
		Labor:            labor,
		Parts:            parts,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	recalcCosts(wo)
	if err := s.repo.Create(ctx, wo); err != nil {
		return nil, err
	}
	return wo, nil
}

func (s *WorkOrderService) Get(ctx context.Context, id string) (*domain.WorkOrder, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	wo, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return wo, nil
}

func (s *WorkOrderService) List(ctx context.Context, limit, offset int32, status, repairType, customerID, vehicleID string) ([]*domain.WorkOrder, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset, status, repairType, customerID, vehicleID)
}

func (s *WorkOrderService) Update(ctx context.Context, id string, in UpdateInput) (*domain.WorkOrder, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if in.CustomerID != nil {
		if cid, err := uuid.Parse(*in.CustomerID); err == nil {
			existing.CustomerID = cid
		}
	}
	if in.VehicleID != nil {
		if vid, err := uuid.Parse(*in.VehicleID); err == nil {
			existing.VehicleID = vid
		}
	}
	if in.DealerPointID != nil {
		existing.DealerPointID = parseOptionalUUID(*in.DealerPointID)
	}
	if in.WarehouseID != nil {
		existing.WarehouseID = parseOptionalUUID(*in.WarehouseID)
	}
	if in.RepairType != nil {
		existing.RepairType = *in.RepairType
	}
	if in.Status != nil {
		existing.Status = *in.Status
	}
	if in.ServiceAdvisorID != nil {
		existing.ServiceAdvisorID = parseOptionalUUID(*in.ServiceAdvisorID)
	}
	if in.Complaint != nil {
		existing.Complaint = *in.Complaint
	}
	if in.Diagnosis != nil {
		existing.Diagnosis = *in.Diagnosis
	}
	if in.MileageKm != nil {
		existing.MileageKm = *in.MileageKm
	}
	if in.OpenedAt != nil {
		existing.OpenedAt = unixToTime(*in.OpenedAt)
	}
	if in.ClosedAt != nil {
		existing.ClosedAt = unixToTime(*in.ClosedAt)
	}
	if in.Notes != nil {
		existing.Notes = *in.Notes
	}
	partsInput := in.Parts
	if in.ReplaceLines {
		now := time.Now().UTC()
		laborRows, err := s.buildLaborRows(ctx, existing.ID, now, in.Labor)
		if err != nil {
			return nil, err
		}
		existing.Labor = laborRows
		newParts, err := buildPartRows(existing.ID, now, in.Parts, existing.WarehouseID)
		if err != nil {
			return nil, err
		}
		issuedParts := make([]domain.WorkOrderPart, 0)
		for _, p := range existing.Parts {
			if p.Issued {
				issuedParts = append(issuedParts, p)
			}
		}
		existing.Parts = append(issuedParts, newParts...)
		partsInput = nil
	}
	laborInput := in.Labor
	if !in.ReplaceLines {
		laborInput = nil
	}
	serviceAdvisorID := existing.ServiceAdvisorID
	if in.ServiceAdvisorID != nil {
		serviceAdvisorID = parseOptionalUUID(*in.ServiceAdvisorID)
	}
	if err := s.validateRefs(ctx, existing.CustomerID, existing.VehicleID, existing.DealerPointID, existing.WarehouseID, serviceAdvisorID, laborInput, partsInput); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now().UTC()
	recalcCosts(existing)
	if err := s.repo.Update(ctx, existing, in.ReplaceLines); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *WorkOrderService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, uid)
}

func (s *WorkOrderService) MovePartsToWork(ctx context.Context, id, createdBy string) (*domain.WorkOrder, error) {
	wo, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if wo.MovementDocumentID != nil && wo.MovementDocumentStatus == "draft" {
		return nil, ErrMovementDocumentExists
	}
	lines := make([]domain.MovementDocumentLineInput, 0)
	for _, p := range wo.Parts {
		if p.Issued {
			continue
		}
		qty, err := quantityToUnits(p.Quantity)
		if err != nil {
			return nil, fmt.Errorf("part line %s: %w", p.ID, err)
		}
		lines = append(lines, domain.MovementDocumentLineInput{
			PartID: p.PartID, WarehouseID: p.WarehouseID, Quantity: qty, LineID: p.ID,
			Notes: fmt.Sprintf("Заказ-наряд %s", wo.OrderNumber), SortOrder: p.SortOrder,
		})
	}
	if len(lines) == 0 {
		if wo.PartsIssued {
			return nil, ErrPartsAlreadyIssued
		}
		return nil, ErrNoPartsToIssue
	}
	docID, err := s.refs.CreateMovementDocument(ctx, wo.ID, wo.OrderNumber, lines, createdBy)
	if err != nil {
		return nil, err
	}
	documentUUID, err := uuid.Parse(docID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SetMovementDocument(ctx, wo.ID, documentUUID, "draft"); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *WorkOrderService) ApplyMovementDocument(ctx context.Context, workOrderID, documentID, documentStatus string) (*domain.WorkOrder, error) {
	woUID, err := uuid.Parse(workOrderID)
	if err != nil {
		return nil, ErrNotFound
	}
	docUID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, errors.New("invalid movement document id")
	}
	wo, err := s.repo.GetByID(ctx, woUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if wo.MovementDocumentID == nil || *wo.MovementDocumentID != docUID {
		return nil, errors.New("movement document mismatch")
	}
	wo.MovementDocumentStatus = documentStatus
	wo.UpdatedAt = time.Now().UTC()
	if documentStatus == "confirmed" {
		lineIDs := make([]uuid.UUID, 0)
		for _, p := range wo.Parts {
			if !p.Issued {
				lineIDs = append(lineIDs, p.ID)
			}
		}
		now := time.Now().UTC()
		if err := s.repo.MarkPartsIssued(ctx, wo.ID, lineIDs, now); err != nil {
			return nil, err
		}
		return s.Get(ctx, workOrderID)
	}
	if err := s.repo.SetMovementDocument(ctx, wo.ID, docUID, documentStatus); err != nil {
		return nil, err
	}
	return s.Get(ctx, workOrderID)
}
