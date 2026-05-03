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

func copyVehString(dst *string, src *string) {
	if src != nil {
		*dst = *src
	}
}

func copyVehInt32(dst *int32, src *int32) {
	if src != nil {
		*dst = *src
	}
}

func copyVehInt64(dst *int64, src *int64) {
	if src != nil {
		*dst = *src
	}
}

func applyVehUUIDClearOrSet(clear bool, src *uuid.UUID, slot **uuid.UUID) {
	if clear {
		*slot = nil
		return
	}
	if src != nil {
		*slot = src
	}
}

func mergeVehicleUpdateInput(v *domain.Vehicle, in UpdateVehicleInput) {
	copyVehString(&v.VIN, in.VIN)
	copyVehString(&v.Make, in.Make)
	copyVehString(&v.Model, in.Model)
	copyVehInt32(&v.Year, in.Year)
	copyVehInt64(&v.MileageKm, in.MileageKm)
	copyVehString(&v.Price, in.Price)
	copyVehString(&v.Status, in.Status)
	copyVehString(&v.Color, in.Color)
	copyVehString(&v.Notes, in.Notes)
	applyVehUUIDClearOrSet(in.ClearBrand, in.BrandID, &v.BrandID)
	applyVehUUIDClearOrSet(in.ClearDealerPoint, in.DealerPointID, &v.DealerPointID)
	applyVehUUIDClearOrSet(in.ClearLegalEntity, in.LegalEntityID, &v.LegalEntityID)
	applyVehUUIDClearOrSet(in.ClearWarehouse, in.WarehouseID, &v.WarehouseID)
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
	mergeVehicleUpdateInput(existing, in)
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
