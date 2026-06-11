package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID            uuid.UUID
	ClientID      uuid.UUID
	UserID        uuid.UUID
	DealerPointID uuid.UUID
	VehicleID     uuid.UUID
	Rating        int32
	Text          string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
