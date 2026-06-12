package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReviewInvitation struct {
	ID            uuid.UUID
	ClientID      uuid.UUID
	UserID        uuid.UUID
	VehicleID     uuid.UUID
	DealerPointID uuid.UUID
	SourceType    string
	SourceID      uuid.UUID
	ServiceKind   string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
