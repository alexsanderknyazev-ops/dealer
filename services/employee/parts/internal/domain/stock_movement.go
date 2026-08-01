package domain

import (
	"time"

	"github.com/google/uuid"
)

const MovementWorkOrderIssue = "work_order_issue"

type StockMovement struct {
	ID                 uuid.UUID
	PartID             uuid.UUID
	WarehouseID        uuid.UUID
	Quantity           int32
	MovementType       string
	ReferenceType      string
	ReferenceID        *uuid.UUID
	ReferenceLineID    *uuid.UUID
	MovementDocumentID *uuid.UUID
	Notes              string
	CreatedBy          *uuid.UUID
	CreatedAt          time.Time
}
