package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/vehicles/internal/domain"
)

var ErrNotFound = errors.New("vehicle not found")

// CreateVehicleInput — поля для создания ТС (снижает число параметров VehicleAPI.Create).
type CreateVehicleInput struct {
	VIN, Make, Model                                   string
	Year                                               int32
	MileageKm                                          int64
	Price, Status, Color, Notes                        string
	BrandID, DealerPointID, LegalEntityID, WarehouseID *uuid.UUID
}

// UpdateVehicleInput — частичное обновление ТС и флаги сброса ссылок.
type UpdateVehicleInput struct {
	VIN, Make, Model            *string
	Year                        *int32
	MileageKm                   *int64
	Price, Status, Color, Notes *string
	BrandID                     *uuid.UUID
	ClearBrand                  bool
	DealerPointID               *uuid.UUID
	LegalEntityID               *uuid.UUID
	WarehouseID                 *uuid.UUID
	ClearDealerPoint            bool
	ClearLegalEntity            bool
	ClearWarehouse              bool
}

type vehicleRepository interface {
	Create(ctx context.Context, v *domain.Vehicle) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error)
	List(ctx context.Context, f domain.VehicleListFilter) ([]*domain.Vehicle, int32, error)
	Update(ctx context.Context, v *domain.Vehicle) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// VehicleAPI — HTTP/gRPC и тесты.
type VehicleAPI interface {
	Create(ctx context.Context, in CreateVehicleInput) (*domain.Vehicle, error)
	Get(ctx context.Context, id string) (*domain.Vehicle, error)
	List(ctx context.Context, f domain.VehicleListFilter) ([]*domain.Vehicle, int32, error)
	Update(ctx context.Context, id string, in UpdateVehicleInput) (*domain.Vehicle, error)
	Delete(ctx context.Context, id string) error
}

type VehicleService struct {
	repo vehicleRepository
}

func NewVehicleService(repo vehicleRepository) *VehicleService {
	return &VehicleService{repo: repo}
}

func (s *VehicleService) Create(ctx context.Context, in CreateVehicleInput) (*domain.Vehicle, error) {
	status := in.Status
	if status == "" {
		status = "available"
	}
	now := time.Now().UTC()
	v := &domain.Vehicle{
		ID:            uuid.New(),
		VIN:           in.VIN,
		Make:          in.Make,
		Model:         in.Model,
		Year:          in.Year,
		MileageKm:     in.MileageKm,
		Price:         in.Price,
		Status:        status,
		Color:         in.Color,
		Notes:         in.Notes,
		BrandID:       in.BrandID,
		DealerPointID: in.DealerPointID,
		LegalEntityID: in.LegalEntityID,
		WarehouseID:   in.WarehouseID,
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

func (s *VehicleService) List(ctx context.Context, f domain.VehicleListFilter) ([]*domain.Vehicle, int32, error) {
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	ff := f
	ff.Limit = limit
	return s.repo.List(ctx, ff)
}

func (s *VehicleService) Update(ctx context.Context, id string, in UpdateVehicleInput) (*domain.Vehicle, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrNotFound
	}
	if in.VIN != nil {
		existing.VIN = *in.VIN
	}
	if in.Make != nil {
		existing.Make = *in.Make
	}
	if in.Model != nil {
		existing.Model = *in.Model
	}
	if in.Year != nil {
		existing.Year = *in.Year
	}
	if in.MileageKm != nil {
		existing.MileageKm = *in.MileageKm
	}
	if in.Price != nil {
		existing.Price = *in.Price
	}
	if in.Status != nil {
		existing.Status = *in.Status
	}
	if in.Color != nil {
		existing.Color = *in.Color
	}
	if in.Notes != nil {
		existing.Notes = *in.Notes
	}
	if in.ClearBrand {
		existing.BrandID = nil
	} else if in.BrandID != nil {
		existing.BrandID = in.BrandID
	}
	if in.ClearDealerPoint {
		existing.DealerPointID = nil
	} else if in.DealerPointID != nil {
		existing.DealerPointID = in.DealerPointID
	}
	if in.ClearLegalEntity {
		existing.LegalEntityID = nil
	} else if in.LegalEntityID != nil {
		existing.LegalEntityID = in.LegalEntityID
	}
	if in.ClearWarehouse {
		existing.WarehouseID = nil
	} else if in.WarehouseID != nil {
		existing.WarehouseID = in.WarehouseID
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
