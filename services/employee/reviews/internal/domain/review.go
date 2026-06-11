package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID             uuid.UUID
	ReviewID       uuid.UUID
	ClientID       uuid.UUID
	UserID         uuid.UUID
	ClientEmail    string
	ClientFullName string
	DealerPointID  uuid.UUID
	VehicleID      uuid.UUID
	VehicleVIN     string
	VehicleMake    string
	VehicleModel   string
	VehicleYear    int32
	Rating         int32
	Text           string
	Status         string
	OccurredAt     time.Time
	CreatedAt      time.Time
}

type ReviewListParams struct {
	ClientID      *uuid.UUID
	DealerPointID *uuid.UUID
	Status        string
	Limit         int32
	Offset        int32
}

type ReviewStats struct {
	TotalCount    int64
	AverageRating float64
	ByStatus      []StatusCount
}

type StatusCount struct {
	Status string
	Count  int64
}
