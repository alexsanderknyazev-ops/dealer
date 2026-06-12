package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusDraft      = "draft"
	StatusScheduled  = "scheduled"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"

	SlotOpenHour  = 8
	SlotCloseHour = 13
)

type RepairAppointmentLabor struct {
	ID          uuid.UUID
	AppointmentID uuid.UUID
	WorkID      *uuid.UUID
	Description string
	Quantity    string
	UnitPrice   string
	SortOrder   int32
	CreatedAt   time.Time
}

type RepairAppointmentPart struct {
	ID            uuid.UUID
	AppointmentID uuid.UUID
	PartID        uuid.UUID
	WarehouseID   uuid.UUID
	Quantity      int32
	UnitPrice     string
	Notes         string
	SortOrder     int32
	CreatedAt     time.Time
}

type RepairAppointment struct {
	ID                uuid.UUID
	AppointmentNumber string
	CustomerID        uuid.UUID
	VehicleID         uuid.UUID
	DealerPointID     *uuid.UUID
	WarehouseID       *uuid.UUID
	ScheduledStart    time.Time
	ScheduledEnd      time.Time
	Status            string
	WorkOrderID       *uuid.UUID
	Complaint         string
	Notes             string
	CreatedBy         *uuid.UUID
	Labor             []RepairAppointmentLabor
	Parts             []RepairAppointmentPart
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TimeSlot struct {
	Start     time.Time
	End       time.Time
	Available bool
	Label     string
}

type LaborInput struct {
	WorkID, Description, Quantity, UnitPrice string
	SortOrder                              int32
}

type PartInput struct {
	PartID, WarehouseID, UnitPrice, Notes string
	Quantity                              int32
	SortOrder                             int32
}

type CreateInput struct {
	CustomerID, VehicleID                  uuid.UUID
	DealerPointID, WarehouseID             *uuid.UUID
	ScheduledStart, ScheduledEnd             time.Time
	Complaint, Notes                       string
	CreatedBy                              *uuid.UUID
	Labor                                  []LaborInput
	Parts                                  []PartInput
}

type UpdateInput struct {
	CustomerID, VehicleID                  *uuid.UUID
	DealerPointID, WarehouseID             *uuid.UUID
	ScheduledStart, ScheduledEnd           *time.Time
	Complaint, Notes                       *string
	Labor                                  []LaborInput
	Parts                                  []PartInput
	ReplaceLines                           bool
}
