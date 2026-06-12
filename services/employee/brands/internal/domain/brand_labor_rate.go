package domain

import (
	"time"

	"github.com/google/uuid"
)

type BrandLaborRate struct {
	ID                    uuid.UUID
	BrandID               uuid.UUID
	DealerPointID         uuid.UUID
	WarrantyHourPrice     string
	CommercialHourPrice   string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
