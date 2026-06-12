package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/workorders/internal/domain"
)

type memWORepo struct {
	byID      map[uuid.UUID]*domain.WorkOrder
	nextNum   int
	err       error
	updateErr error
}

type memWORefs struct {
	err         error
	custOK      *bool
	vehOK       *bool
	workOK      *bool
	employeeOK  *bool
	partOK      *bool
	warehouseOK *bool
	workName    string
	workHours   string
	workPrice   string
	moveDocID   string
	moveDocErr  error
}

func (m *memWORefs) CustomerExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.custOK == nil {
		return true, nil
	}
	return *m.custOK, nil
}

func (m *memWORefs) VehicleExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.vehOK == nil {
		return true, nil
	}
	return *m.vehOK, nil
}

func (m *memWORefs) DealerPointExists(_ context.Context, _ uuid.UUID) (bool, error) {
	return true, nil
}

func (m *memWORefs) WarehouseExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.warehouseOK == nil {
		return true, nil
	}
	return *m.warehouseOK, nil
}

func (m *memWORefs) PartExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.partOK == nil {
		return true, nil
	}
	return *m.partOK, nil
}

func (m *memWORefs) WorkExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.workOK == nil {
		return true, nil
	}
	return *m.workOK, nil
}

func (m *memWORefs) ResolveWork(_ context.Context, _ uuid.UUID) (string, string, string, error) {
	if m.workName == "" {
		return "Diagnostic", "1", "2000", nil
	}
	return m.workName, m.workHours, m.workPrice, nil
}

func (m *memWORefs) EmployeeExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.employeeOK == nil {
		return true, nil
	}
	return *m.employeeOK, nil
}

func (m *memWORefs) EmployeeFullName(_ context.Context, _ uuid.UUID) string {
	return "QA Master"
}

func (m *memWORefs) CreateMovementDocument(_ context.Context, _ uuid.UUID, _ string, _ []domain.MovementDocumentLineInput, _ string) (string, error) {
	if m.moveDocErr != nil {
		return "", m.moveDocErr
	}
	if m.moveDocID != "" {
		return m.moveDocID, nil
	}
	return uuid.New().String(), nil
}

func (m *memWORepo) NextOrderNumber(_ context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.nextNum++
	return "WO-TEST-0001", nil
}

func (m *memWORepo) Create(_ context.Context, wo *domain.WorkOrder) error {
	if m.err != nil {
		return m.err
	}
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.WorkOrder)
	}
	cp := *wo
	m.byID[wo.ID] = &cp
	return nil
}

func (m *memWORepo) GetByID(_ context.Context, id uuid.UUID) (*domain.WorkOrder, error) {
	if m.err != nil {
		return nil, m.err
	}
	wo, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *wo
	return &cp, nil
}

func (m *memWORepo) List(_ context.Context, _, _ int32, _, _, _, _ string) ([]*domain.WorkOrder, int32, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var out []*domain.WorkOrder
	for _, wo := range m.byID {
		cp := *wo
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}

func (m *memWORepo) Update(_ context.Context, wo *domain.WorkOrder, _ bool) error {
	if m.err != nil {
		return m.err
	}
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.byID[wo.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *wo
	m.byID[wo.ID] = &cp
	return nil
}

func (m *memWORepo) MarkPartsIssued(_ context.Context, workOrderID uuid.UUID, lineIDs []uuid.UUID, issuedAt time.Time) error {
	wo, ok := m.byID[workOrderID]
	if !ok {
		return pgx.ErrNoRows
	}
	issued := make(map[uuid.UUID]struct{}, len(lineIDs))
	for _, id := range lineIDs {
		issued[id] = struct{}{}
	}
	for i := range wo.Parts {
		if _, ok := issued[wo.Parts[i].ID]; ok {
			wo.Parts[i].Issued = true
		}
	}
	wo.PartsIssued = true
	wo.PartsIssuedAt = &issuedAt
	return nil
}

func (m *memWORepo) SetMovementDocument(_ context.Context, workOrderID, documentID uuid.UUID, status string) error {
	wo, ok := m.byID[workOrderID]
	if !ok {
		return pgx.ErrNoRows
	}
	wo.MovementDocumentID = &documentID
	wo.MovementDocumentStatus = status
	return nil
}

func (m *memWORepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.byID, id)
	return nil
}

func TestWorkOrderService_Create_DefaultsAndCosts(t *testing.T) {
	repo := &memWORepo{byID: map[uuid.UUID]*domain.WorkOrder{}}
	s := NewWorkOrderService(repo, &memWORefs{})
	cid, vid, wid, pid, whid := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()

	wo, err := s.Create(context.Background(), CreateInput{
		CustomerID: cid.String(),
		VehicleID:  vid.String(),
		Labor: []LaborInput{{
			WorkID:      wid.String(),
			Description: "Diagnostic",
			Quantity:    "1",
			UnitPrice:   "2000",
		}},
		Parts: []PartInput{{
			PartID:      pid.String(),
			WarehouseID: whid.String(),
			Quantity:    "2",
			UnitPrice:   "850",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if wo.Status != domain.StatusDraft || wo.RepairType != domain.RepairCommercial {
		t.Fatalf("status=%q repair=%q", wo.Status, wo.RepairType)
	}
	if wo.LaborCost != "2000.00" || wo.PartsCost != "1700.00" || wo.TotalCost != "3700.00" {
		t.Fatalf("costs labor=%s parts=%s total=%s", wo.LaborCost, wo.PartsCost, wo.TotalCost)
	}
	if wo.Labor[0].WorkID == nil || wo.Labor[0].WorkID.String() != wid.String() {
		t.Fatalf("work_id not set: %+v", wo.Labor[0].WorkID)
	}
}

func TestWorkOrderService_Create_ResolveWorkFromCatalog(t *testing.T) {
	repo := &memWORepo{byID: map[uuid.UUID]*domain.WorkOrder{}}
	s := NewWorkOrderService(repo, &memWORefs{workName: "Oil change", workHours: "0.5", workPrice: "2500"})
	cid, vid, wid := uuid.New(), uuid.New(), uuid.New()

	wo, err := s.Create(context.Background(), CreateInput{
		CustomerID: cid.String(),
		VehicleID:  vid.String(),
		Labor:      []LaborInput{{WorkID: wid.String()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if wo.Labor[0].Description != "Oil change" || wo.Labor[0].Quantity != "0.5" || wo.Labor[0].UnitPrice != "2500" {
		t.Fatalf("%+v", wo.Labor[0])
	}
}

func TestWorkOrderService_Create_InvalidWorkID(t *testing.T) {
	s := NewWorkOrderService(&memWORepo{byID: map[uuid.UUID]*domain.WorkOrder{}}, &memWORefs{})
	_, err := s.Create(context.Background(), CreateInput{
		CustomerID: uuid.New().String(),
		VehicleID:  uuid.New().String(),
		Labor:      []LaborInput{{WorkID: "not-uuid"}},
	})
	if !errors.Is(err, ErrWorkNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestWorkOrderService_Create_CustomerNotFound(t *testing.T) {
	missing := false
	s := NewWorkOrderService(&memWORepo{byID: map[uuid.UUID]*domain.WorkOrder{}}, &memWORefs{custOK: &missing})
	_, err := s.Create(context.Background(), CreateInput{
		CustomerID: uuid.New().String(),
		VehicleID:  uuid.New().String(),
	})
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestWorkOrderService_MovePartsToWork(t *testing.T) {
	repo := &memWORepo{byID: map[uuid.UUID]*domain.WorkOrder{}}
	docID := uuid.New().String()
	s := NewWorkOrderService(repo, &memWORefs{moveDocID: docID})
	cid, vid, wid, pid, whid := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()

	wo, err := s.Create(context.Background(), CreateInput{
		CustomerID: cid.String(),
		VehicleID:  vid.String(),
		Labor:      []LaborInput{{WorkID: wid.String(), Description: "X", Quantity: "1", UnitPrice: "100"}},
		Parts:      []PartInput{{PartID: pid.String(), WarehouseID: whid.String(), Quantity: "2", UnitPrice: "50"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := s.MovePartsToWork(context.Background(), wo.ID.String(), uuid.New().String())
	if err != nil {
		t.Fatal(err)
	}
	if updated.MovementDocumentID == nil || updated.MovementDocumentStatus != "draft" {
		t.Fatalf("%+v", updated)
	}
}

func TestWorkOrderService_MovePartsToWork_NoParts(t *testing.T) {
	repo := &memWORepo{byID: map[uuid.UUID]*domain.WorkOrder{}}
	s := NewWorkOrderService(repo, &memWORefs{})
	cid, vid, wid := uuid.New(), uuid.New(), uuid.New()

	wo, err := s.Create(context.Background(), CreateInput{
		CustomerID: cid.String(),
		VehicleID:  vid.String(),
		Labor:      []LaborInput{{WorkID: wid.String(), Description: "X", Quantity: "1", UnitPrice: "100"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.MovePartsToWork(context.Background(), wo.ID.String(), uuid.New().String())
	if !errors.Is(err, ErrNoPartsToIssue) {
		t.Fatalf("got %v", err)
	}
}

func TestWorkOrderService_ApplyMovementDocument_Confirmed(t *testing.T) {
	repo := &memWORepo{byID: map[uuid.UUID]*domain.WorkOrder{}}
	s := NewWorkOrderService(repo, &memWORefs{})
	cid, vid, wid, pid, whid := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	docID := uuid.New()

	wo, err := s.Create(context.Background(), CreateInput{
		CustomerID: cid.String(),
		VehicleID:  vid.String(),
		Labor:      []LaborInput{{WorkID: wid.String(), Description: "X", Quantity: "1", UnitPrice: "100"}},
		Parts:      []PartInput{{PartID: pid.String(), WarehouseID: whid.String(), Quantity: "1", UnitPrice: "10"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	stored := repo.byID[wo.ID]
	stored.MovementDocumentID = &docID
	stored.MovementDocumentStatus = "draft"

	updated, err := s.ApplyMovementDocument(context.Background(), wo.ID.String(), docID.String(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	if !updated.PartsIssued {
		t.Fatal("expected parts_issued=true")
	}
}

func TestWorkOrderService_Get_NotFound(t *testing.T) {
	s := NewWorkOrderService(&memWORepo{byID: map[uuid.UUID]*domain.WorkOrder{}}, nil)
	_, err := s.Get(context.Background(), uuid.New().String())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}
