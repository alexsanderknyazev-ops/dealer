package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RepairWarrantyManufacturer = "warranty_manufacturer"
	RepairPreSale              = "pre_sale"
	RepairCommercial           = "commercial"
	RepairMaintenance          = "maintenance"

	StatusDraft      = "draft"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusClosed     = "closed"
	StatusPaid       = "paid"
)

type WorkOrderLabor struct {
	ID          uuid.UUID
	WorkOrderID uuid.UUID
	WorkID      *uuid.UUID
	Description string
	Quantity    string
	UnitPrice   string
	Amount      string
	ExecutorID  *uuid.UUID
	SortOrder   int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkOrderPart struct {
	ID          uuid.UUID
	WorkOrderID uuid.UUID
	PartID      uuid.UUID
	WarehouseID uuid.UUID
	Description string
	Quantity    string
	UnitPrice   string
	Amount      string
	Issued      bool
	SortOrder   int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkOrder struct {
	ID               uuid.UUID
	OrderNumber      string
	CustomerID       uuid.UUID
	VehicleID        uuid.UUID
	DealerPointID    *uuid.UUID
	WarehouseID      *uuid.UUID
	RepairType       string
	Status           string
	ServiceAdvisorID *uuid.UUID
	Complaint        string
	Diagnosis        string
	MileageKm        int64
	LaborCost        string
	PartsCost        string
	TotalCost        string
	OpenedAt         *time.Time
	ClosedAt         *time.Time
	PartsIssued            bool
	PartsIssuedAt          *time.Time
	MovementDocumentID     *uuid.UUID
	MovementDocumentStatus string
	SourceOrderType        string
	SourceOrderID          *uuid.UUID
	Notes                  string
	Labor            []WorkOrderLabor
	Parts            []WorkOrderPart
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type MovementDocumentLineInput struct {
	PartID      uuid.UUID
	WarehouseID uuid.UUID
	Quantity    int32
	LineID      uuid.UUID
	Notes       string
	SortOrder   int32
}
