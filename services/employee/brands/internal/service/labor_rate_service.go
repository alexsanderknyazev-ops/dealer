package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/brands/internal/domain"
)

var ErrLaborRateNotFound = errors.New("brand labor rate not found")

type LaborRateAPI interface {
	List(ctx context.Context, limit, offset int32, brandID, dealerPointID string) ([]*domain.BrandLaborRate, int32, error)
	Update(ctx context.Context, brandID, dealerPointID, warrantyPrice, commercialPrice string) (*domain.BrandLaborRate, error)
	Delete(ctx context.Context, id string) error
	Resolve(ctx context.Context, brandID, dealerPointID, repairType string) (warranty, commercial, hour string, found bool, err error)
}

type laborRateRepository interface {
	GetByBrandAndDealerPoint(ctx context.Context, brandID, dealerPointID uuid.UUID) (*domain.BrandLaborRate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BrandLaborRate, error)
	List(ctx context.Context, limit, offset int32, brandID, dealerPointID *uuid.UUID) ([]*domain.BrandLaborRate, int32, error)
	Upsert(ctx context.Context, rate *domain.BrandLaborRate) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type LaborRateService struct {
	rates laborRateRepository
	brands brandRepository
}

func NewLaborRateService(rates laborRateRepository, brands brandRepository) *LaborRateService {
	return &LaborRateService{rates: rates, brands: brands}
}

var _ LaborRateAPI = (*LaborRateService)(nil)

func parseOptionalUUID(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (s *LaborRateService) List(ctx context.Context, limit, offset int32, brandID, dealerPointID string) ([]*domain.BrandLaborRate, int32, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	bid, err := parseOptionalUUID(brandID)
	if err != nil {
		return nil, 0, ErrNotFound
	}
	dpid, err := parseOptionalUUID(dealerPointID)
	if err != nil {
		return nil, 0, ErrNotFound
	}
	return s.rates.List(ctx, limit, offset, bid, dpid)
}

func (s *LaborRateService) Update(ctx context.Context, brandID, dealerPointID, warrantyPrice, commercialPrice string) (*domain.BrandLaborRate, error) {
	bid, err := uuid.Parse(brandID)
	if err != nil {
		return nil, ErrNotFound
	}
	dpid, err := uuid.Parse(dealerPointID)
	if err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.brands.GetByID(ctx, bid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	existing, err := s.rates.GetByBrandAndDealerPoint(ctx, bid, dpid)
	now := time.Now().UTC()
	rate := &domain.BrandLaborRate{
		ID:                  uuid.New(),
		BrandID:             bid,
		DealerPointID:       dpid,
		WarrantyHourPrice:   warrantyPrice,
		CommercialHourPrice: commercialPrice,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err == nil {
		rate.ID = existing.ID
		rate.CreatedAt = existing.CreatedAt
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err := s.rates.Upsert(ctx, rate); err != nil {
		return nil, err
	}
	return s.rates.GetByBrandAndDealerPoint(ctx, bid, dpid)
}

func (s *LaborRateService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrLaborRateNotFound
	}
	if err := s.rates.Delete(ctx, uid); err != nil {
		return err
	}
	return nil
}

func hourPriceForRepairType(repairType, warranty, commercial string) string {
	if repairType == "warranty_manufacturer" {
		return warranty
	}
	return commercial
}

func (s *LaborRateService) Resolve(ctx context.Context, brandID, dealerPointID, repairType string) (warranty, commercial, hour string, found bool, err error) {
	bid, err := uuid.Parse(brandID)
	if err != nil {
		return "", "", "", false, nil
	}
	dpid, err := uuid.Parse(dealerPointID)
	if err != nil {
		return "", "", "", false, nil
	}
	rate, err := s.rates.GetByBrandAndDealerPoint(ctx, bid, dpid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", "", false, nil
		}
		return "", "", "", false, err
	}
	warranty = rate.WarrantyHourPrice
	commercial = rate.CommercialHourPrice
	hour = hourPriceForRepairType(repairType, warranty, commercial)
	return warranty, commercial, hour, true, nil
}
