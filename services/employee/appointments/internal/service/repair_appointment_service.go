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

	"github.com/dealer/dealer/services/appointments/internal/domain"
)

var (
	ErrNotFound           = errors.New("repair appointment not found")
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrVehicleNotFound    = errors.New("vehicle not found")
	ErrWarehouseNotFound  = errors.New("warehouse not found")
	ErrPartNotFound       = errors.New("part not found")
	ErrWorkNotFound       = errors.New("work not found")
	ErrSlotUnavailable    = errors.New("time slot is not available")
	ErrInvalidTimeRange   = errors.New("invalid time range")
	ErrNotEditable        = errors.New("appointment cannot be edited")
	ErrWorkOrderExists    = errors.New("work order already exists")
	ErrWorkOrdersUnavailable = errors.New("workorders service unavailable")
)

type appointmentRepository interface {
	NextNumber(ctx context.Context) (string, error)
	Create(ctx context.Context, a *domain.RepairAppointment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RepairAppointment, error)
	Update(ctx context.Context, a *domain.RepairAppointment, replaceLines bool) error
	List(ctx context.Context, limit, offset int32, status string, from, to *time.Time) ([]*domain.RepairAppointment, int32, error)
	HasOverlap(ctx context.Context, start, end time.Time, excludeID *uuid.UUID) (bool, error)
	ListBusyInRange(ctx context.Context, from, to time.Time) ([]domain.RepairAppointment, error)
	SetWorkOrder(ctx context.Context, id, workOrderID uuid.UUID, updatedAt time.Time) error
}

type ReferenceChecker interface {
	CustomerExists(ctx context.Context, id uuid.UUID) (bool, error)
	VehicleExists(ctx context.Context, id uuid.UUID) (bool, error)
	WarehouseExists(ctx context.Context, id uuid.UUID) (bool, error)
	PartExists(ctx context.Context, id uuid.UUID) (bool, error)
	WorkExists(ctx context.Context, id uuid.UUID) (bool, error)
}

type WorkOrdersCreator interface {
	CreateFromAppointment(ctx context.Context, a *domain.RepairAppointment) (id, number string, err error)
	GetOrderNumber(ctx context.Context, id string) string
}

type noopRefs struct{}

func (noopRefs) CustomerExists(context.Context, uuid.UUID) (bool, error) { return true, nil }
func (noopRefs) VehicleExists(context.Context, uuid.UUID) (bool, error)  { return true, nil }
func (noopRefs) WarehouseExists(context.Context, uuid.UUID) (bool, error) { return true, nil }
func (noopRefs) PartExists(context.Context, uuid.UUID) (bool, error)     { return true, nil }
func (noopRefs) WorkExists(context.Context, uuid.UUID) (bool, error)      { return true, nil }

type RepairAppointmentService struct {
	repo       appointmentRepository
	refs       ReferenceChecker
	workOrders WorkOrdersCreator
}

func NewRepairAppointmentService(repo appointmentRepository, refs ReferenceChecker, workOrders WorkOrdersCreator) *RepairAppointmentService {
	if refs == nil {
		refs = noopRefs{}
	}
	return &RepairAppointmentService{repo: repo, refs: refs, workOrders: workOrders}
}

func parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date required")
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

func slotLabel(start, end time.Time) string {
	return fmt.Sprintf("%02d:%02d–%02d:%02d", start.Hour(), start.Minute(), end.Hour(), end.Minute())
}

func (s *RepairAppointmentService) ListSlots(ctx context.Context, dateStr string) ([]domain.TimeSlot, error) {
	day, err := parseDate(dateStr)
	if err != nil {
		return nil, err
	}
	dayEnd := day.Add(24 * time.Hour)
	busy, err := s.repo.ListBusyInRange(ctx, day, dayEnd)
	if err != nil {
		return nil, err
	}
	slots := make([]domain.TimeSlot, 0, domain.SlotCloseHour-domain.SlotOpenHour)
	for h := domain.SlotOpenHour; h < domain.SlotCloseHour; h++ {
		start := day.Add(time.Duration(h) * time.Hour)
		end := day.Add(time.Duration(h+1) * time.Hour)
		available := true
		for _, b := range busy {
			if start.Before(b.ScheduledEnd) && end.After(b.ScheduledStart) {
				available = false
				break
			}
		}
		slots = append(slots, domain.TimeSlot{
			Start: start, End: end, Available: available, Label: slotLabel(start, end),
		})
	}
	return slots, nil
}

func (s *RepairAppointmentService) validateRefs(ctx context.Context, customerID, vehicleID uuid.UUID, warehouseID *uuid.UUID, labor []domain.LaborInput, parts []domain.PartInput) error {
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
	if warehouseID != nil {
		ok, err = s.refs.WarehouseExists(ctx, *warehouseID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrWarehouseNotFound
		}
	}
	for _, l := range labor {
		if strings.TrimSpace(l.WorkID) != "" {
			wid, err := uuid.Parse(l.WorkID)
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
		}
	}
	for _, p := range parts {
		pid, err := uuid.Parse(p.PartID)
		if err != nil {
			return ErrPartNotFound
		}
		ok, err := s.refs.PartExists(ctx, pid)
		if err != nil {
			return err
		}
		if !ok {
			return ErrPartNotFound
		}
		whID, err := uuid.Parse(p.WarehouseID)
		if err != nil {
			return ErrWarehouseNotFound
		}
		ok, err = s.refs.WarehouseExists(ctx, whID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrWarehouseNotFound
		}
	}
	return nil
}

func (s *RepairAppointmentService) buildLabor(appointmentID uuid.UUID, now time.Time, inputs []domain.LaborInput) []domain.RepairAppointmentLabor {
	out := make([]domain.RepairAppointmentLabor, 0, len(inputs))
	for _, in := range inputs {
		var workID *uuid.UUID
		if strings.TrimSpace(in.WorkID) != "" {
			if id, err := uuid.Parse(in.WorkID); err == nil {
				workID = &id
			}
		}
		qty := strings.TrimSpace(in.Quantity)
		if qty == "" {
			qty = "1"
		}
		price := strings.TrimSpace(in.UnitPrice)
		if price == "" {
			price = "0"
		}
		out = append(out, domain.RepairAppointmentLabor{
			ID: uuid.New(), AppointmentID: appointmentID, WorkID: workID,
			Description: in.Description, Quantity: qty, UnitPrice: price,
			SortOrder: in.SortOrder, CreatedAt: now,
		})
	}
	return out
}

func (s *RepairAppointmentService) buildParts(appointmentID uuid.UUID, now time.Time, inputs []domain.PartInput, defaultWarehouse *uuid.UUID) ([]domain.RepairAppointmentPart, error) {
	out := make([]domain.RepairAppointmentPart, 0, len(inputs))
	for _, in := range inputs {
		partID, err := uuid.Parse(in.PartID)
		if err != nil {
			return nil, ErrPartNotFound
		}
		whID := defaultWarehouse
		if strings.TrimSpace(in.WarehouseID) != "" {
			parsed, err := uuid.Parse(in.WarehouseID)
			if err != nil {
				return nil, ErrWarehouseNotFound
			}
			whID = &parsed
		}
		if whID == nil {
			return nil, ErrWarehouseNotFound
		}
		if in.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity")
		}
		price := strings.TrimSpace(in.UnitPrice)
		if price == "" {
			price = "0"
		}
		out = append(out, domain.RepairAppointmentPart{
			ID: uuid.New(), AppointmentID: appointmentID, PartID: partID, WarehouseID: *whID,
			Quantity: in.Quantity, UnitPrice: price, Notes: in.Notes, SortOrder: in.SortOrder, CreatedAt: now,
		})
	}
	return out, nil
}

func (s *RepairAppointmentService) ensureSlot(ctx context.Context, start, end time.Time, excludeID *uuid.UUID) error {
	if !end.After(start) {
		return ErrInvalidTimeRange
	}
	overlap, err := s.repo.HasOverlap(ctx, start, end, excludeID)
	if err != nil {
		return err
	}
	if overlap {
		return ErrSlotUnavailable
	}
	return nil
}

func (s *RepairAppointmentService) Create(ctx context.Context, in domain.CreateInput) (*domain.RepairAppointment, error) {
	if err := s.ensureSlot(ctx, in.ScheduledStart, in.ScheduledEnd, nil); err != nil {
		return nil, err
	}
	if err := s.validateRefs(ctx, in.CustomerID, in.VehicleID, in.WarehouseID, in.Labor, in.Parts); err != nil {
		return nil, err
	}
	number, err := s.repo.NextNumber(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	id := uuid.New()
	parts, err := s.buildParts(id, now, in.Parts, in.WarehouseID)
	if err != nil {
		return nil, err
	}
	a := &domain.RepairAppointment{
		ID: id, AppointmentNumber: number, CustomerID: in.CustomerID, VehicleID: in.VehicleID,
		DealerPointID: in.DealerPointID, WarehouseID: in.WarehouseID,
		ScheduledStart: in.ScheduledStart, ScheduledEnd: in.ScheduledEnd,
		Status: domain.StatusScheduled, Complaint: in.Complaint, Notes: in.Notes,
		CreatedBy: in.CreatedBy, Labor: s.buildLabor(id, now, in.Labor), Parts: parts,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return s.Get(ctx, id.String())
}

func (s *RepairAppointmentService) Get(ctx context.Context, id string) (*domain.RepairAppointment, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	a, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (s *RepairAppointmentService) List(ctx context.Context, limit, offset int32, status string, from, to *time.Time) ([]*domain.RepairAppointment, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset, status, from, to)
}

func (s *RepairAppointmentService) Update(ctx context.Context, id string, in domain.UpdateInput) (*domain.RepairAppointment, error) {
	a, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Status == domain.StatusCancelled || a.Status == domain.StatusCompleted || a.WorkOrderID != nil {
		return nil, ErrNotEditable
	}
	if in.CustomerID != nil {
		a.CustomerID = *in.CustomerID
	}
	if in.VehicleID != nil {
		a.VehicleID = *in.VehicleID
	}
	if in.DealerPointID != nil {
		a.DealerPointID = in.DealerPointID
	}
	if in.WarehouseID != nil {
		a.WarehouseID = in.WarehouseID
	}
	if in.ScheduledStart != nil {
		a.ScheduledStart = *in.ScheduledStart
	}
	if in.ScheduledEnd != nil {
		a.ScheduledEnd = *in.ScheduledEnd
	}
	if err := s.ensureSlot(ctx, a.ScheduledStart, a.ScheduledEnd, &a.ID); err != nil {
		return nil, err
	}
	if in.Complaint != nil {
		a.Complaint = *in.Complaint
	}
	if in.Notes != nil {
		a.Notes = *in.Notes
	}
	laborIn := in.Labor
	partsIn := in.Parts
	if in.ReplaceLines {
		now := time.Now().UTC()
		a.Labor = s.buildLabor(a.ID, now, in.Labor)
		parts, err := s.buildParts(a.ID, now, in.Parts, a.WarehouseID)
		if err != nil {
			return nil, err
		}
		a.Parts = parts
		laborIn = nil
		partsIn = nil
	}
	if err := s.validateRefs(ctx, a.CustomerID, a.VehicleID, a.WarehouseID, laborIn, partsIn); err != nil {
		return nil, err
	}
	a.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, a, in.ReplaceLines); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *RepairAppointmentService) Cancel(ctx context.Context, id string) (*domain.RepairAppointment, error) {
	a, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Status == domain.StatusCancelled {
		return a, nil
	}
	if a.WorkOrderID != nil {
		return nil, ErrNotEditable
	}
	a.Status = domain.StatusCancelled
	a.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, a, false); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *RepairAppointmentService) CreateWorkOrder(ctx context.Context, id string) (*domain.RepairAppointment, string, string, error) {
	a, err := s.Get(ctx, id)
	if err != nil {
		return nil, "", "", err
	}
	if a.WorkOrderID != nil {
		return nil, "", "", ErrWorkOrderExists
	}
	if a.Status == domain.StatusCancelled {
		return nil, "", "", ErrNotEditable
	}
	if s.workOrders == nil {
		return nil, "", "", ErrWorkOrdersUnavailable
	}
	woID, woNum, err := s.workOrders.CreateFromAppointment(ctx, a)
	if err != nil {
		return nil, "", "", err
	}
	parsed, err := uuid.Parse(woID)
	if err != nil {
		return nil, "", "", err
	}
	now := time.Now().UTC()
	if err := s.repo.SetWorkOrder(ctx, a.ID, parsed, now); err != nil {
		return nil, "", "", err
	}
	updated, err := s.Get(ctx, id)
	if err != nil {
		return nil, woID, woNum, nil
	}
	return updated, woID, woNum, nil
}

func FormatQuantity(q int32) string {
	return strconv.Itoa(int(q))
}
