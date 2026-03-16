package domain

import (
	"time"

	"github.com/google/uuid"
)

type Vehicle struct {
	ID        uuid.UUID
	VIN       string
	Make      string
	Model     string
	Year      int32
	MileageKm int64
	Price     string
	Status    string // available, sold, reserved
	Color     string
	Notes           string
	BrandID         *uuid.UUID
	DealerPointID   *uuid.UUID
	LegalEntityID   *uuid.UUID
	WarehouseID     *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt time.Time
}
