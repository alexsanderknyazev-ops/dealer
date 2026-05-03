package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

var ErrNotFound = errors.New("part not found")
var ErrFolderNotFound = errors.New("folder not found")

// StockRow — склад и количество для остатков запчасти
type StockRow struct {
	WarehouseID uuid.UUID
	Quantity    int32
}

type partRepository interface {
	Create(ctx context.Context, p *domain.Part) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Part, error)
	List(ctx context.Context, filter domain.PartListFilter) ([]*domain.Part, int32, error)
	Update(ctx context.Context, p *domain.Part) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type folderRepository interface {
	Create(ctx context.Context, f *domain.PartFolder) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PartFolder, error)
	ListByParent(ctx context.Context, parentID *uuid.UUID) ([]*domain.PartFolder, error)
	Update(ctx context.Context, f *domain.PartFolder) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type partStockRepository interface {
	ListByPart(ctx context.Context, partID uuid.UUID) ([]*domain.PartStock, error)
	Upsert(ctx context.Context, partID, warehouseID uuid.UUID, quantity int32) error
	ReplaceForPart(ctx context.Context, partID uuid.UUID, rows []struct {
		WarehouseID uuid.UUID
		Quantity    int32
	}) error
}

type PartService struct {
	repo       partRepository
	folderRepo folderRepository
	stockRepo  partStockRepository
}

func NewPartService(repo partRepository, folderRepo folderRepository, stockRepo partStockRepository) *PartService {
	return &PartService{repo: repo, folderRepo: folderRepo, stockRepo: stockRepo}
}

func (s *PartService) Create(ctx context.Context, sku, name, category string, folderID, brandID, dealerPointID, legalEntityID, warehouseID *uuid.UUID, quantity int32, unit, price, location, notes string, initialStock []StockRow) (*domain.Part, error) {
	if unit == "" {
		unit = "шт"
	}
	now := time.Now().UTC()
	p := &domain.Part{
		ID:            uuid.New(),
		SKU:           sku,
		Name:          name,
		Category:      category,
		FolderID:      folderID,
		BrandID:       brandID,
		DealerPointID: dealerPointID,
		LegalEntityID: legalEntityID,
		WarehouseID:   warehouseID,
		Quantity:      0, // пересчитается из part_stock триггером
		Unit:          unit,
		Price:         price,
		Location:      location,
		Notes:         notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	if len(initialStock) > 0 {
		rows := make([]struct {
			WarehouseID uuid.UUID
			Quantity    int32
		}, len(initialStock))
		for i, row := range initialStock {
			rows[i].WarehouseID = row.WarehouseID
			rows[i].Quantity = row.Quantity
		}
		if err := s.stockRepo.ReplaceForPart(ctx, p.ID, rows); err != nil {
			return nil, err
		}
		// перечитать part — quantity обновлён триггером
		p, _ = s.repo.GetByID(ctx, p.ID)
	} else if warehouseID != nil && quantity > 0 {
		if err := s.stockRepo.Upsert(ctx, p.ID, *warehouseID, quantity); err != nil {
			return nil, err
		}
		p, _ = s.repo.GetByID(ctx, p.ID)
	}
	return p, nil
}

func (s *PartService) Get(ctx context.Context, id string) (*domain.Part, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	p, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *PartService) List(ctx context.Context, limit, offset int32, search, categoryFilter string, folderID, brandID, dealerPointID, legalEntityID, warehouseID *uuid.UUID) ([]*domain.Part, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, domain.PartListFilter{
		Limit: limit, Offset: offset, Search: search, CategoryFilter: categoryFilter,
		FolderID: folderID, BrandID: brandID, DealerPointID: dealerPointID,
		LegalEntityID: legalEntityID, WarehouseID: warehouseID,
	})
}

func (s *PartService) Update(ctx context.Context, id string, sku, name, category *string, folderIDOpt, brandIDOpt, dealerPointIDOpt, legalEntityIDOpt, warehouseIDOpt *string, quantity *int32, unit, price, location, notes *string) (*domain.Part, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrNotFound
	}
	if sku != nil {
		existing.SKU = *sku
	}
	if name != nil {
		existing.Name = *name
	}
	if category != nil {
		existing.Category = *category
	}
	if folderIDOpt != nil {
		if *folderIDOpt == "" {
			existing.FolderID = nil
		} else {
			fid, err := uuid.Parse(*folderIDOpt)
			if err == nil {
				existing.FolderID = &fid
			}
		}
	}
	if brandIDOpt != nil {
		if *brandIDOpt == "" {
			existing.BrandID = nil
		} else {
			bid, err := uuid.Parse(*brandIDOpt)
			if err == nil {
				existing.BrandID = &bid
			}
		}
	}
	if dealerPointIDOpt != nil {
		if *dealerPointIDOpt == "" {
			existing.DealerPointID = nil
		} else {
			did, err := uuid.Parse(*dealerPointIDOpt)
			if err == nil {
				existing.DealerPointID = &did
			}
		}
	}
	if legalEntityIDOpt != nil {
		if *legalEntityIDOpt == "" {
			existing.LegalEntityID = nil
		} else {
			lid, err := uuid.Parse(*legalEntityIDOpt)
			if err == nil {
				existing.LegalEntityID = &lid
			}
		}
	}
	if warehouseIDOpt != nil {
		if *warehouseIDOpt == "" {
			existing.WarehouseID = nil
		} else {
			wid, err := uuid.Parse(*warehouseIDOpt)
			if err == nil {
				existing.WarehouseID = &wid
			}
		}
	}
	// quantity теперь хранится в part_stock и пересчитывается триггером; прямое изменение parts.quantity не делаем
	if unit != nil {
		existing.Unit = *unit
	}
	if price != nil {
		existing.Price = *price
	}
	if location != nil {
		existing.Location = *location
	}
	if notes != nil {
		existing.Notes = *notes
	}
	existing.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *PartService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, uid)
}

// ListStock возвращает остатки запчасти по складам
func (s *PartService) ListStock(ctx context.Context, partID string) ([]*domain.PartStock, error) {
	uid, err := uuid.Parse(partID)
	if err != nil {
		return nil, ErrNotFound
	}
	return s.stockRepo.ListByPart(ctx, uid)
}

// ReplaceStock заменяет все остатки запчасти по складам
func (s *PartService) ReplaceStock(ctx context.Context, partID string, rows []StockRow) error {
	uid, err := uuid.Parse(partID)
	if err != nil {
		return ErrNotFound
	}
	repoRows := make([]struct {
		WarehouseID uuid.UUID
		Quantity    int32
	}, 0, len(rows))
	for _, row := range rows {
		if row.Quantity < 0 {
			continue
		}
		repoRows = append(repoRows, struct {
			WarehouseID uuid.UUID
			Quantity    int32
		}{row.WarehouseID, row.Quantity})
	}
	return s.stockRepo.ReplaceForPart(ctx, uid, repoRows)
}

// Folders

func (s *PartService) CreateFolder(ctx context.Context, name string, parentID *uuid.UUID) (*domain.PartFolder, error) {
	now := time.Now().UTC()
	f := &domain.PartFolder{
		ID:        uuid.New(),
		Name:      name,
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.folderRepo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *PartService) GetFolder(ctx context.Context, id string) (*domain.PartFolder, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	f, err := s.folderRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}
	return f, nil
}

func (s *PartService) ListFolders(ctx context.Context, parentID *uuid.UUID) ([]*domain.PartFolder, error) {
	return s.folderRepo.ListByParent(ctx, parentID)
}

func (s *PartService) UpdateFolder(ctx context.Context, id string, name *string, parentIDOpt *string) (*domain.PartFolder, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	existing, err := s.folderRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	if name != nil {
		existing.Name = *name
	}
	if parentIDOpt != nil {
		if *parentIDOpt == "" {
			existing.ParentID = nil
		} else {
			pid, err := uuid.Parse(*parentIDOpt)
			if err == nil {
				existing.ParentID = &pid
			}
		}
	}
	existing.UpdatedAt = time.Now().UTC()
	if err := s.folderRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *PartService) DeleteFolder(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrFolderNotFound
	}
	return s.folderRepo.Delete(ctx, uid)
}
