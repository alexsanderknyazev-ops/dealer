package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	DocumentStatusDraft      = "draft"
	DocumentStatusInProgress = "in_progress"
	DocumentStatusClosed     = "closed"
	DocumentStatusCancelled  = "cancelled"

	MovementTypeWorkOrderIssue  = "work_order_issue"
	MovementTypeTransfer        = "transfer"
	MovementTypeToProduction    = "to_production"
	MovementTypeFromProduction  = "from_production"

	RefWorkOrder        = "work_order"
	RefMovementDocument = "movement_document"
)

// MovementTypeSupportsExtraction — закрытый документ этого типа можно частично вернуть извлечением.
func MovementTypeSupportsExtraction(movementType string) bool {
	switch movementType {
	case MovementTypeWorkOrderIssue, MovementTypeTransfer, MovementTypeToProduction:
		return true
	default:
		return false
	}
}

type MovementDocumentLine struct {
	ID                     uuid.UUID
	DocumentID             uuid.UUID
	PartID                 uuid.UUID
	WarehouseID            uuid.UUID
	DestinationWarehouseID *uuid.UUID
	Quantity               int32
	ReferenceLineID *uuid.UUID
	Notes           string
	SortOrder       int32
	CreatedAt       time.Time
}

type MovementDocument struct {
	ID               uuid.UUID
	DocumentNumber   string
	Status           string
	MovementType     string
	ReferenceType    string
	ReferenceID      *uuid.UUID
	ParentDocumentID *uuid.UUID
	Notes            string
	CreatedBy      *uuid.UUID
	ConfirmedBy    *uuid.UUID
	Lines          []MovementDocumentLine
	CreatedAt      time.Time
	ConfirmedAt    *time.Time
	UpdatedAt      time.Time
}

type MovementDocumentLineInput struct {
	PartID                 uuid.UUID
	WarehouseID            uuid.UUID
	DestinationWarehouseID *uuid.UUID
	Quantity               int32
	ReferenceLineID *uuid.UUID
	Notes           string
	SortOrder       int32
}

type CreateMovementDocumentInput struct {
	MovementType  string
	ReferenceType string
	ReferenceID   *uuid.UUID
	Notes         string
	CreatedBy     *uuid.UUID
	Lines         []MovementDocumentLineInput
}

type UpdateMovementDocumentInput struct {
	MovementType  *string
	Notes         *string
	Lines         []MovementDocumentLineInput
	ReplaceLines  bool
}
