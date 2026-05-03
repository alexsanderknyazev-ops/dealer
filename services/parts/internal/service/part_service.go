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

// StockRow — alias для строк остатков (HTTP JSON и ReplaceStock).
type StockRow = domain.PartWarehouseQty

// CreatePartInput is the payload for Create.
type CreatePartInput struct {
	SKU, Name, Category string
	FolderID, BrandID, DealerPointID, LegalEntityID, WarehouseID *uuid.UUID
	Quantity                         int32
	Unit, Price, Location, Notes       string
	InitialStock                     []StockRow
}

// UpdatePartInput holds optional fields for Update (optional string IDs follow HTTP clear semantics).
type UpdatePartInput struct {
	SKU, Name, Category *string
	FolderID, BrandID, DealerPointID, LegalEntityID, WarehouseID *string
	Quantity *int32
	Unit, Price, Location, Notes *string
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
	ReplaceForPart(ctx context.Context, partID uuid.UUID, rows []domain.PartWarehouseQty) error
}

type PartService struct {
	repo       partRepository
	folderRepo folderRepository
	stockRepo  partStockRepository
}

func NewPartService(repo partRepository, folderRepo folderRepository, stockRepo partStockRepository) *PartService {
	return &PartService{repo: repo, folderRepo: folderRepo, stockRepo: stockRepo}
}

func (s *PartService) Create(ctx context.Context, in CreatePartInput) (*domain.Part, error) {
	unit := in.Unit
	if unit == "" {
		unit = "шт"
	}
	now := time.Now().UTC()
	p := &domain.Part{
		ID:            uuid.New(),
		SKU:           in.SKU,
		Name:          in.Name,
		Category:      in.Category,
		FolderID:      in.FolderID,
		BrandID:       in.BrandID,
		DealerPointID: in.DealerPointID,
		LegalEntityID: in.LegalEntityID,
		WarehouseID:   in.WarehouseID,
		Quantity:      0, // пересчитается из part_stock триггером
		Unit:          unit,
		Price:         in.Price,
		Location:      in.Location,
		Notes:         in.Notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	if len(in.InitialStock) > 0 {
		if err := s.stockRepo.ReplaceForPart(ctx, p.ID, in.InitialStock); err != nil {
			return nil, err
		}
		// перечитать part — quantity обновлён триггером
		p, _ = s.repo.GetByID(ctx, p.ID)
	} else if in.WarehouseID != nil && in.Quantity > 0 {
		if err := s.stockRepo.Upsert(ctx, p.ID, *in.WarehouseID, in.Quantity); err != nil {
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

func (s *PartService) List(ctx context.Context, filter domain.PartListFilter) ([]*domain.Part, int32, error) {
	f := filter
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.repo.List(ctx, f)
}

func (s *PartService) Update(ctx context.Context, id string, in UpdatePartInput) (*domain.Part, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrNotFound
	}
	if in.SKU != nil {
		existing.SKU = *in.SKU
	}
	if in.Name != nil {
		existing.Name = *in.Name
	}
	if in.Category != nil {
		existing.Category = *in.Category
	}
	if in.FolderID != nil {
		if *in.FolderID == "" {
			existing.FolderID = nil
		} else {
			fid, err := uuid.Parse(*in.FolderID)
			if err == nil {
				existing.FolderID = &fid
			}
		}
	}
	if in.BrandID != nil {
		if *in.BrandID == "" {
			existing.BrandID = nil
		} else {
			bid, err := uuid.Parse(*in.BrandID)
			if err == nil {
				existing.BrandID = &bid
			}
		}
	}
	if in.DealerPointID != nil {
		if *in.DealerPointID == "" {
			existing.DealerPointID = nil
		} else {
			did, err := uuid.Parse(*in.DealerPointID)
			if err == nil {
				existing.DealerPointID = &did
			}
		}
	}
	if in.LegalEntityID != nil {
		if *in.LegalEntityID == "" {
			existing.LegalEntityID = nil
		} else {
			lid, err := uuid.Parse(*in.LegalEntityID)
			if err == nil {
				existing.LegalEntityID = &lid
			}
		}
	}
	if in.WarehouseID != nil {
		if *in.WarehouseID == "" {
			existing.WarehouseID = nil
		} else {
			wid, err := uuid.Parse(*in.WarehouseID)
			if err == nil {
				existing.WarehouseID = &wid
			}
		}
	}
	// quantity теперь хранится в part_stock и пересчитывается триггером; прямое изменение parts.quantity не делаем
	if in.Unit != nil {
		existing.Unit = *in.Unit
	}
	if in.Price != nil {
		existing.Price = *in.Price
	}
	if in.Location != nil {
		existing.Location = *in.Location
	}
	if in.Notes != nil {
		existing.Notes = *in.Notes
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
	repoRows := make([]domain.PartWarehouseQty, 0, len(rows))
	for _, row := range rows {
		if row.Quantity < 0 {
			continue
		}
		repoRows = append(repoRows, row)
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
