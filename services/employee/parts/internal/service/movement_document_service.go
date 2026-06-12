package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

var (
	ErrMovementDocumentNotFound      = errors.New("movement document not found")
	ErrMovementDocumentNotDraft      = errors.New("movement document is not draft")
	ErrMovementDocumentNotInProgress = errors.New("movement document is not in progress")
	ErrMovementDocumentNotEditable   = errors.New("movement document cannot be edited")
	ErrMovementDocumentNoLines       = errors.New("movement document has no lines")
	ErrInsufficientStock             = errors.New("insufficient stock")
)

type stockMovementRepository interface {
	Create(ctx context.Context, m *domain.StockMovement) error
}

type movementDocumentRepository interface {
	NextDocumentNumber(ctx context.Context) (string, error)
	Create(ctx context.Context, doc *domain.MovementDocument) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.MovementDocument, error)
	List(ctx context.Context, limit, offset int32, status, referenceType, referenceID string) ([]*domain.MovementDocument, int32, error)
	UpdateStatus(ctx context.Context, doc *domain.MovementDocument) error
	Update(ctx context.Context, doc *domain.MovementDocument, replaceLines bool) error
}

type WorkOrdersNotifier interface {
	ApplyMovementDocument(ctx context.Context, workOrderID, documentID, status string) error
}

type noopWorkOrdersNotifier struct{}

func (noopWorkOrdersNotifier) ApplyMovementDocument(context.Context, string, string, string) error {
	return nil
}

func (s *PartService) buildMovementLines(ctx context.Context, docID uuid.UUID, inputs []domain.MovementDocumentLineInput, now time.Time) ([]domain.MovementDocumentLine, error) {
	lines := make([]domain.MovementDocumentLine, 0, len(inputs))
	for _, lineIn := range inputs {
		if lineIn.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for part %s", lineIn.PartID)
		}
		if _, err := s.repo.GetByID(ctx, lineIn.PartID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		if err := s.checkRef(ctx, &lineIn.WarehouseID, s.dealerPoints.WarehouseExists, ErrWarehouseNotFound); err != nil {
			return nil, err
		}
		lines = append(lines, domain.MovementDocumentLine{
			ID:              uuid.New(),
			DocumentID:      docID,
			PartID:          lineIn.PartID,
			WarehouseID:     lineIn.WarehouseID,
			Quantity:        lineIn.Quantity,
			ReferenceLineID: lineIn.ReferenceLineID,
			Notes:           lineIn.Notes,
			SortOrder:       lineIn.SortOrder,
			CreatedAt:       now,
		})
	}
	return lines, nil
}

func (s *PartService) CreateMovementDocument(ctx context.Context, in domain.CreateMovementDocumentInput) (*domain.MovementDocument, error) {
	number, err := s.movementDocRepo.NextDocumentNumber(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	docID := uuid.New()
	var lines []domain.MovementDocumentLine
	if len(in.Lines) > 0 {
		lines, err = s.buildMovementLines(ctx, docID, in.Lines, now)
		if err != nil {
			return nil, err
		}
	}
	doc := &domain.MovementDocument{
		ID:             docID,
		DocumentNumber: number,
		Status:         domain.DocumentStatusDraft,
		MovementType:   in.MovementType,
		ReferenceType:  in.ReferenceType,
		ReferenceID:    in.ReferenceID,
		Notes:          in.Notes,
		CreatedBy:      in.CreatedBy,
		Lines:          lines,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.movementDocRepo.Create(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *PartService) GetMovementDocument(ctx context.Context, id string) (*domain.MovementDocument, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrMovementDocumentNotFound
	}
	doc, err := s.movementDocRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMovementDocumentNotFound
		}
		return nil, err
	}
	return doc, nil
}

func (s *PartService) ListMovementDocuments(ctx context.Context, limit, offset int32, status, referenceType, referenceID string) ([]*domain.MovementDocument, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.movementDocRepo.List(ctx, limit, offset, status, referenceType, referenceID)
}

func (s *PartService) UpdateMovementDocument(ctx context.Context, id string, in domain.UpdateMovementDocumentInput) (*domain.MovementDocument, error) {
	doc, err := s.GetMovementDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocumentStatusDraft && doc.Status != domain.DocumentStatusInProgress {
		return nil, ErrMovementDocumentNotEditable
	}
	now := time.Now().UTC()
	if in.MovementType != nil {
		doc.MovementType = *in.MovementType
	}
	if in.Notes != nil {
		doc.Notes = *in.Notes
	}
	if in.ReplaceLines {
		doc.Lines, err = s.buildMovementLines(ctx, doc.ID, in.Lines, now)
		if err != nil {
			return nil, err
		}
	}
	doc.UpdatedAt = now
	if err := s.movementDocRepo.Update(ctx, doc, in.ReplaceLines); err != nil {
		return nil, err
	}
	return s.GetMovementDocument(ctx, id)
}

func (s *PartService) StartMovementDocument(ctx context.Context, id string) (*domain.MovementDocument, error) {
	doc, err := s.GetMovementDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocumentStatusDraft {
		return nil, ErrMovementDocumentNotDraft
	}
	if len(doc.Lines) == 0 {
		return nil, ErrMovementDocumentNoLines
	}
	doc.Status = domain.DocumentStatusInProgress
	doc.UpdatedAt = time.Now().UTC()
	if err := s.movementDocRepo.UpdateStatus(ctx, doc); err != nil {
		return nil, err
	}
	if doc.ReferenceType == domain.RefWorkOrder && doc.ReferenceID != nil {
		_ = s.workOrders.ApplyMovementDocument(ctx, doc.ReferenceID.String(), doc.ID.String(), doc.Status)
	}
	return s.GetMovementDocument(ctx, id)
}

func (s *PartService) CloseMovementDocument(ctx context.Context, id string, closedBy *uuid.UUID) (*domain.MovementDocument, error) {
	doc, err := s.GetMovementDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocumentStatusInProgress {
		return nil, ErrMovementDocumentNotInProgress
	}
	if len(doc.Lines) == 0 {
		return nil, ErrMovementDocumentNoLines
	}
	now := time.Now().UTC()
	for _, line := range doc.Lines {
		if _, err := s.stockRepo.Deduct(ctx, line.PartID, line.WarehouseID, line.Quantity); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				qty, _ := s.stockRepo.GetQuantity(ctx, line.PartID, line.WarehouseID)
				return nil, fmt.Errorf("%w: part %s on warehouse %s (have %d, need %d)",
					ErrInsufficientStock, line.PartID, line.WarehouseID, qty, line.Quantity)
			}
			return nil, err
		}
		refLine := line.ReferenceLineID
		movement := domain.StockMovement{
			ID:                 uuid.New(),
			PartID:             line.PartID,
			WarehouseID:        line.WarehouseID,
			Quantity:           -line.Quantity,
			MovementType:       doc.MovementType,
			ReferenceType:      doc.ReferenceType,
			ReferenceID:        doc.ReferenceID,
			ReferenceLineID:    refLine,
			MovementDocumentID: &doc.ID,
			Notes:              line.Notes,
			CreatedBy:          closedBy,
			CreatedAt:          now,
		}
		if err := s.movementRepo.Create(ctx, &movement); err != nil {
			return nil, err
		}
	}
	doc.Status = domain.DocumentStatusClosed
	doc.ConfirmedBy = closedBy
	doc.ConfirmedAt = &now
	doc.UpdatedAt = now
	if err := s.movementDocRepo.UpdateStatus(ctx, doc); err != nil {
		return nil, err
	}
	if doc.ReferenceType == domain.RefWorkOrder && doc.ReferenceID != nil {
		if err := s.workOrders.ApplyMovementDocument(ctx, doc.ReferenceID.String(), doc.ID.String(), doc.Status); err != nil {
			return nil, fmt.Errorf("notify work order: %w", err)
		}
	}
	return s.GetMovementDocument(ctx, id)
}

// ConfirmMovementDocument — совместимость: закрывает документ (списание при closed).
func (s *PartService) ConfirmMovementDocument(ctx context.Context, id string, confirmedBy *uuid.UUID) (*domain.MovementDocument, error) {
	return s.CloseMovementDocument(ctx, id, confirmedBy)
}

func (s *PartService) CancelMovementDocument(ctx context.Context, id string) (*domain.MovementDocument, error) {
	doc, err := s.GetMovementDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc.Status != domain.DocumentStatusDraft && doc.Status != domain.DocumentStatusInProgress {
		return nil, ErrMovementDocumentNotDraft
	}
	now := time.Now().UTC()
	doc.Status = domain.DocumentStatusCancelled
	doc.UpdatedAt = now
	if err := s.movementDocRepo.UpdateStatus(ctx, doc); err != nil {
		return nil, err
	}
	if doc.ReferenceType == domain.RefWorkOrder && doc.ReferenceID != nil {
		_ = s.workOrders.ApplyMovementDocument(ctx, doc.ReferenceID.String(), doc.ID.String(), doc.Status)
	}
	return doc, nil
}
