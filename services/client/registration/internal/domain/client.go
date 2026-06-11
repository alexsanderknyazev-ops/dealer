package domain

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Email     string
	FullName  string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ClientVehicle struct {
	ID        uuid.UUID
	ClientID  uuid.UUID
	VehicleID uuid.UUID
	VIN       string
	Make      string
	Model     string
	Year      int32
	AddedAt   time.Time
}
