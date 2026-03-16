package domain

import (
	"time"

	"github.com/google/uuid"
)

type PartFolder struct {
	ID        uuid.UUID
	Name      string
	ParentID  *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Part struct {
	ID        uuid.UUID
	SKU       string
	Name      string
	Category  string
	FolderID       *uuid.UUID
	BrandID        *uuid.UUID
	DealerPointID  *uuid.UUID
	LegalEntityID  *uuid.UUID
	WarehouseID    *uuid.UUID
	Quantity       int32
	Unit      string
	Price     string
	Location  string
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PartStock — остаток запчасти на конкретном складе (одна запчасть может быть на нескольких складах)
type PartStock struct {
	PartID      uuid.UUID
	WarehouseID uuid.UUID
	Quantity    int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
