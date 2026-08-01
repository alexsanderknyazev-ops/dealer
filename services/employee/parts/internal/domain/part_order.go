package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OrderStatusDraft     = "draft"
	OrderStatusLinked    = "linked"
	OrderStatusFulfilled = "fulfilled"
	OrderStatusCancelled = "cancelled"

	RefSupplierOrder = "supplier_order"
	RefCustomerOrder = "customer_order"
)

type PartOrderLine struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	PartID    uuid.UUID
	Quantity  int32
	UnitPrice string
	Notes     string
	SortOrder int32
	CreatedAt time.Time
}

type PartOrderLineInput struct {
	PartID    uuid.UUID
	Quantity  int32
	UnitPrice string
	Notes     string
	SortOrder int32
}

type SupplierOrder struct {
	ID                           uuid.UUID
	OrderNumber                  string
	Status                       string
	SupplierID                   uuid.UUID
	ReceiptWarehouseID           uuid.UUID
	CustomerOrderID              *uuid.UUID
	FulfillmentMovementDocumentID *uuid.UUID
	FulfillmentWorkOrderID        *uuid.UUID
	Notes                        string
	CreatedBy                    *uuid.UUID
	Lines                        []PartOrderLine
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

type CustomerOrder struct {
	ID                           uuid.UUID
	OrderNumber                  string
	Status                       string
	CustomerID                   uuid.UUID
	VehicleID                    *uuid.UUID
	IssueWarehouseID             uuid.UUID
	FulfillmentMovementDocumentID *uuid.UUID
	FulfillmentWorkOrderID        *uuid.UUID
	Notes                        string
	CreatedBy                    *uuid.UUID
	Lines                        []PartOrderLine
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

type CreateSupplierOrderInput struct {
	SupplierID         uuid.UUID
	ReceiptWarehouseID uuid.UUID
	CustomerOrderID    *uuid.UUID
	Notes              string
	CreatedBy          *uuid.UUID
	Lines              []PartOrderLineInput
}

type UpdateSupplierOrderInput struct {
	SupplierID         *uuid.UUID
	ReceiptWarehouseID *uuid.UUID
	CustomerOrderID    *uuid.UUID
	ClearCustomerOrder bool
	Notes              *string
	Lines              []PartOrderLineInput
	ReplaceLines       bool
}

type CreateCustomerOrderInput struct {
	CustomerID       uuid.UUID
	VehicleID        *uuid.UUID
	VehicleVIN       string
	IssueWarehouseID uuid.UUID
	Notes            string
	CreatedBy        *uuid.UUID
	Lines            []PartOrderLineInput
}

type UpdateCustomerOrderInput struct {
	CustomerID       *uuid.UUID
	VehicleID          *uuid.UUID
	VehicleVIN         *string
	ClearVehicle       bool
	IssueWarehouseID   *uuid.UUID
	Notes              *string
	Lines              []PartOrderLineInput
	ReplaceLines       bool
}
