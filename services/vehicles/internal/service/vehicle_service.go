package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/vehicles/internal/domain"
	"github.com/dealer/dealer/services/vehicles/internal/repository"
)

var ErrNotFound = errors.New("vehicle not found")

type VehicleService struct {
	repo *repository.VehicleRepository
}

func NewVehicleService(repo *repository.VehicleRepository) *VehicleService {
	return &VehicleService{repo: repo}
}

func (s *VehicleService) Create(ctx context.Context, vin, make, model string, year int32, mileageKm int64, price, status, color, notes string, brandID, dealerPointID, legalEntityID, warehouseID *uuid.UUID) (*domain.Vehicle, error) {
	if status == "" {
		status = "available"
	}
	now := time.Now().UTC()
	v := &domain.Vehicle{
		ID:            uuid.New(),
		VIN:           vin,
		Make:          make,
		Model:         model,
		Year:          year,
		MileageKm:     mileageKm,
		Price:         price,
		Status:        status,
		Color:         color,
		Notes:         notes,
		BrandID:       brandID,
		DealerPointID: dealerPointID,
		LegalEntityID: legalEntityID,
		WarehouseID:   warehouseID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.Create(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VehicleService) Get(ctx context.Context, id string) (*domain.Vehicle, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	v, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return v, nil
}

func (s *VehicleService) List(ctx context.Context, limit, offset int32, search, statusFilter string, brandID, dealerPointID, legalEntityID, warehouseID *uuid.UUID) ([]*domain.Vehicle, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset, search, statusFilter, brandID, dealerPointID, legalEntityID, warehouseID)
}

func (s *VehicleService) Update(ctx context.Context, id string, vin, make, model *string, year *int32, mileageKm *int64, price, status, color, notes *string, brandID *uuid.UUID, clearBrand bool, dealerPointID, legalEntityID, warehouseID *uuid.UUID, clearDealerPoint, clearLegalEntity, clearWarehouse bool) (*domain.Vehicle, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrNotFound
	}
	if vin != nil {
		existing.VIN = *vin
	}
	if make != nil {
		existing.Make = *make
	}
	if model != nil {
		existing.Model = *model
	}
	if year != nil {
		existing.Year = *year
	}
	if mileageKm != nil {
		existing.MileageKm = *mileageKm
	}
	if price != nil {
		existing.Price = *price
	}
	if status != nil {
		existing.Status = *status
	}
	if color != nil {
		existing.Color = *color
	}
	if notes != nil {
		existing.Notes = *notes
	}
	if clearBrand {
		existing.BrandID = nil
	} else if brandID != nil {
		existing.BrandID = brandID
	}
	if clearDealerPoint {
		existing.DealerPointID = nil
	} else if dealerPointID != nil {
		existing.DealerPointID = dealerPointID
	}
	if clearLegalEntity {
		existing.LegalEntityID = nil
	} else if legalEntityID != nil {
		existing.LegalEntityID = legalEntityID
	}
	if clearWarehouse {
		existing.WarehouseID = nil
	} else if warehouseID != nil {
		existing.WarehouseID = warehouseID
	}
	existing.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *VehicleService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, uid)
}
