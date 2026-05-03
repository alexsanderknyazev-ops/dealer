package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
)

var (
	ErrDealerPointNotFound = errors.New("dealer point not found")
	ErrLegalEntityNotFound = errors.New("legal entity not found")
	ErrWarehouseNotFound   = errors.New("warehouse not found")
)

type dealerPointRepository interface {
	Create(ctx context.Context, d *domain.DealerPoint) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.DealerPoint, error)
	List(ctx context.Context, limit, offset int32, search string) ([]*domain.DealerPoint, int32, error)
	Update(ctx context.Context, d *domain.DealerPoint) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type legalEntityRepository interface {
	Create(ctx context.Context, e *domain.LegalEntity) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.LegalEntity, error)
	List(ctx context.Context, limit, offset int32, search string) ([]*domain.LegalEntity, int32, error)
	ListByDealerPoint(ctx context.Context, dealerPointID uuid.UUID, limit, offset int32) ([]*domain.LegalEntity, int32, error)
	Update(ctx context.Context, e *domain.LegalEntity) error
	Delete(ctx context.Context, id uuid.UUID) error
	LinkToDealerPoint(ctx context.Context, dealerPointID, legalEntityID uuid.UUID) error
	UnlinkFromDealerPoint(ctx context.Context, dealerPointID, legalEntityID uuid.UUID) error
}

type warehouseRepository interface {
	Create(ctx context.Context, w *domain.Warehouse) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error)
	List(ctx context.Context, limit, offset int32, dealerPointID, legalEntityID *uuid.UUID, typeFilter string) ([]*domain.Warehouse, int32, error)
	Update(ctx context.Context, w *domain.Warehouse) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type DealerPointsService struct {
	dpRepo dealerPointRepository
	leRepo legalEntityRepository
	whRepo warehouseRepository
}

func NewDealerPointsService(
	dpRepo dealerPointRepository,
	leRepo legalEntityRepository,
	whRepo warehouseRepository,
) *DealerPointsService {
	return &DealerPointsService{dpRepo: dpRepo, leRepo: leRepo, whRepo: whRepo}
}

// Dealer points
func (s *DealerPointsService) CreateDealerPoint(ctx context.Context, name, address string) (*domain.DealerPoint, error) {
	now := time.Now().UTC()
	d := &domain.DealerPoint{ID: uuid.New(), Name: name, Address: address, CreatedAt: now, UpdatedAt: now}
	if err := s.dpRepo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DealerPointsService) GetDealerPoint(ctx context.Context, id string) (*domain.DealerPoint, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrDealerPointNotFound
	}
	d, err := s.dpRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDealerPointNotFound
		}
		return nil, err
	}
	return d, nil
}

func (s *DealerPointsService) ListDealerPoints(ctx context.Context, limit, offset int32, search string) ([]*domain.DealerPoint, int32, error) {
	return s.dpRepo.List(ctx, limit, offset, search)
}

func (s *DealerPointsService) UpdateDealerPoint(ctx context.Context, id string, name, address *string) (*domain.DealerPoint, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrDealerPointNotFound
	}
	d, err := s.dpRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrDealerPointNotFound
	}
	if name != nil {
		d.Name = *name
	}
	if address != nil {
		d.Address = *address
	}
	d.UpdatedAt = time.Now().UTC()
	if err := s.dpRepo.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DealerPointsService) DeleteDealerPoint(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrDealerPointNotFound
	}
	return s.dpRepo.Delete(ctx, uid)
}

// Legal entities
func (s *DealerPointsService) CreateLegalEntity(ctx context.Context, name, inn, address string) (*domain.LegalEntity, error) {
	now := time.Now().UTC()
	e := &domain.LegalEntity{ID: uuid.New(), Name: name, INN: inn, Address: address, CreatedAt: now, UpdatedAt: now}
	if err := s.leRepo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *DealerPointsService) GetLegalEntity(ctx context.Context, id string) (*domain.LegalEntity, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrLegalEntityNotFound
	}
	e, err := s.leRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLegalEntityNotFound
		}
		return nil, err
	}
	return e, nil
}

func (s *DealerPointsService) ListLegalEntities(ctx context.Context, limit, offset int32, search string) ([]*domain.LegalEntity, int32, error) {
	return s.leRepo.List(ctx, limit, offset, search)
}

func (s *DealerPointsService) ListLegalEntitiesByDealerPoint(ctx context.Context, dealerPointID string, limit, offset int32) ([]*domain.LegalEntity, int32, error) {
	uid, err := uuid.Parse(dealerPointID)
	if err != nil {
		return nil, 0, ErrDealerPointNotFound
	}
	return s.leRepo.ListByDealerPoint(ctx, uid, limit, offset)
}

func (s *DealerPointsService) UpdateLegalEntity(ctx context.Context, id string, name, inn, address *string) (*domain.LegalEntity, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrLegalEntityNotFound
	}
	e, err := s.leRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrLegalEntityNotFound
	}
	if name != nil {
		e.Name = *name
	}
	if inn != nil {
		e.INN = *inn
	}
	if address != nil {
		e.Address = *address
	}
	e.UpdatedAt = time.Now().UTC()
	if err := s.leRepo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *DealerPointsService) DeleteLegalEntity(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrLegalEntityNotFound
	}
	return s.leRepo.Delete(ctx, uid)
}

func (s *DealerPointsService) LinkLegalEntityToDealerPoint(ctx context.Context, dealerPointID, legalEntityID string) error {
	dpID, err := uuid.Parse(dealerPointID)
	if err != nil {
		return ErrDealerPointNotFound
	}
	leID, err := uuid.Parse(legalEntityID)
	if err != nil {
		return ErrLegalEntityNotFound
	}
	return s.leRepo.LinkToDealerPoint(ctx, dpID, leID)
}

func (s *DealerPointsService) UnlinkLegalEntityFromDealerPoint(ctx context.Context, dealerPointID, legalEntityID string) error {
	dpID, err := uuid.Parse(dealerPointID)
	if err != nil {
		return ErrDealerPointNotFound
	}
	leID, err := uuid.Parse(legalEntityID)
	if err != nil {
		return ErrLegalEntityNotFound
	}
	return s.leRepo.UnlinkFromDealerPoint(ctx, dpID, leID)
}

// Warehouses
func (s *DealerPointsService) CreateWarehouse(ctx context.Context, dealerPointID, legalEntityID, typ, name string) (*domain.Warehouse, error) {
	if typ != "cars" && typ != "parts" {
		typ = "parts"
	}
	dpID, err := uuid.Parse(dealerPointID)
	if err != nil {
		return nil, ErrDealerPointNotFound
	}
	leID, err := uuid.Parse(legalEntityID)
	if err != nil {
		return nil, ErrLegalEntityNotFound
	}
	now := time.Now().UTC()
	w := &domain.Warehouse{
		ID: uuid.New(), DealerPointID: dpID, LegalEntityID: leID, Type: typ, Name: name,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.whRepo.Create(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *DealerPointsService) GetWarehouse(ctx context.Context, id string) (*domain.Warehouse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrWarehouseNotFound
	}
	w, err := s.whRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWarehouseNotFound
		}
		return nil, err
	}
	return w, nil
}

func (s *DealerPointsService) ListWarehouses(ctx context.Context, limit, offset int32, dealerPointID, legalEntityID, typeFilter string) ([]*domain.Warehouse, int32, error) {
	var dpID, leID *uuid.UUID
	if dealerPointID != "" {
		if u, err := uuid.Parse(dealerPointID); err == nil {
			dpID = &u
		}
	}
	if legalEntityID != "" {
		if u, err := uuid.Parse(legalEntityID); err == nil {
			leID = &u
		}
	}
	return s.whRepo.List(ctx, limit, offset, dpID, leID, typeFilter)
}

func (s *DealerPointsService) UpdateWarehouse(ctx context.Context, id string, name *string) (*domain.Warehouse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrWarehouseNotFound
	}
	w, err := s.whRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrWarehouseNotFound
	}
	if name != nil {
		w.Name = *name
	}
	w.UpdatedAt = time.Now().UTC()
	if err := s.whRepo.Update(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *DealerPointsService) DeleteWarehouse(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrWarehouseNotFound
	}
	return s.whRepo.Delete(ctx, uid)
}
