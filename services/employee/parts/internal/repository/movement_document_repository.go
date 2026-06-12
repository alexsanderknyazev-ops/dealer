package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

type MovementDocumentRepository struct {
	pool *pgxpool.Pool
}

func NewMovementDocumentRepository(pool *pgxpool.Pool) *MovementDocumentRepository {
	return &MovementDocumentRepository{pool: pool}
}

func (r *MovementDocumentRepository) NextDocumentNumber(ctx context.Context) (string, error) {
	var n int64
	if err := r.pool.QueryRow(ctx, "SELECT nextval('movement_documents_number_seq')").Scan(&n); err != nil {
		return "", err
	}
	return fmt.Sprintf("MOV-%d-%05d", time.Now().Year(), n), nil
}

func (r *MovementDocumentRepository) Create(ctx context.Context, doc *domain.MovementDocument) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		INSERT INTO movement_documents (
			id, document_number, status, movement_type, reference_type, reference_id,
			parent_document_id, customer_id, vehicle_id, supplier_id, receipt_warehouse_id,
			notes, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`, doc.ID, doc.DocumentNumber, doc.Status, doc.MovementType, doc.ReferenceType, doc.ReferenceID,
		doc.ParentDocumentID, doc.CustomerID, doc.VehicleID, doc.SupplierID, doc.ReceiptWarehouseID,
		doc.Notes, doc.CreatedBy, doc.CreatedAt, doc.UpdatedAt)
	if err != nil {
		return err
	}
	for _, line := range doc.Lines {
		_, err = tx.Exec(ctx, `
			INSERT INTO movement_document_lines (
				id, document_id, part_id, warehouse_id, destination_warehouse_id,
				quantity, unit_cost, reference_line_id, notes, sort_order, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7::numeric,$8,$9,$10,$11)
		`, line.ID, doc.ID, line.PartID, line.WarehouseID, line.DestinationWarehouseID,
			line.Quantity, line.UnitCost, line.ReferenceLineID, line.Notes, line.SortOrder, line.CreatedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

const documentSelect = `
	SELECT id, document_number, status, movement_type, reference_type, reference_id,
		parent_document_id, customer_id, vehicle_id, supplier_id, receipt_warehouse_id,
		notes, created_by, confirmed_by, created_at, confirmed_at, updated_at
	FROM movement_documents
`

func (r *MovementDocumentRepository) scanHeader(row pgx.Row) (*domain.MovementDocument, error) {
	var d domain.MovementDocument
	err := row.Scan(
		&d.ID, &d.DocumentNumber, &d.Status, &d.MovementType, &d.ReferenceType, &d.ReferenceID,
		&d.ParentDocumentID, &d.CustomerID, &d.VehicleID, &d.SupplierID, &d.ReceiptWarehouseID,
		&d.Notes, &d.CreatedBy, &d.ConfirmedBy, &d.CreatedAt, &d.ConfirmedAt, &d.UpdatedAt,
	)
	return &d, err
}

func (r *MovementDocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MovementDocument, error) {
	doc, err := r.scanHeader(r.pool.QueryRow(ctx, documentSelect+" WHERE id = $1", id))
	if err != nil {
		return nil, err
	}
	doc.Lines, err = r.listLines(ctx, id)
	return doc, err
}

func (r *MovementDocumentRepository) listLines(ctx context.Context, documentID uuid.UUID) ([]domain.MovementDocumentLine, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, document_id, part_id, warehouse_id, destination_warehouse_id,
		       quantity, unit_cost::text, reference_line_id, notes, sort_order, created_at
		FROM movement_document_lines WHERE document_id = $1 ORDER BY sort_order, created_at
	`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MovementDocumentLine
	for rows.Next() {
		var l domain.MovementDocumentLine
		if err := rows.Scan(
			&l.ID, &l.DocumentID, &l.PartID, &l.WarehouseID, &l.DestinationWarehouseID,
			&l.Quantity, &l.UnitCost, &l.ReferenceLineID, &l.Notes, &l.SortOrder, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

func (r *MovementDocumentRepository) List(ctx context.Context, limit, offset int32, status, referenceType, referenceID string) ([]*domain.MovementDocument, int32, error) {
	countQuery := "SELECT COUNT(*) FROM movement_documents WHERE 1=1"
	listQuery := documentSelect + " WHERE 1=1"
	args := []any{}
	argNum := 1
	if status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", argNum)
		listQuery += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, status)
		argNum++
	}
	if referenceType != "" {
		countQuery += fmt.Sprintf(" AND reference_type = $%d", argNum)
		listQuery += fmt.Sprintf(" AND reference_type = $%d", argNum)
		args = append(args, referenceType)
		argNum++
	}
	if referenceID != "" {
		if rid, err := uuid.Parse(referenceID); err == nil {
			countQuery += fmt.Sprintf(" AND reference_id = $%d", argNum)
			listQuery += fmt.Sprintf(" AND reference_id = $%d", argNum)
			args = append(args, rid)
			argNum++
		}
	}
	var total int32
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQuery += " ORDER BY created_at DESC LIMIT $" + fmt.Sprint(argNum) + " OFFSET $" + fmt.Sprint(argNum+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.MovementDocument
	for rows.Next() {
		doc, err := r.scanHeader(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, doc)
	}
	return list, total, nil
}

func (r *MovementDocumentRepository) UpdateStatus(ctx context.Context, doc *domain.MovementDocument) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE movement_documents
		SET status=$2, confirmed_by=$3, confirmed_at=$4, updated_at=$5
		WHERE id=$1
	`, doc.ID, doc.Status, doc.ConfirmedBy, doc.ConfirmedAt, doc.UpdatedAt)
	return err
}

func (r *MovementDocumentRepository) Update(ctx context.Context, doc *domain.MovementDocument, replaceLines bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		UPDATE movement_documents
		SET movement_type=$2, customer_id=$3, vehicle_id=$4, supplier_id=$5, receipt_warehouse_id=$6,
		    notes=$7, updated_at=$8
		WHERE id=$1
	`, doc.ID, doc.MovementType, doc.CustomerID, doc.VehicleID, doc.SupplierID, doc.ReceiptWarehouseID,
		doc.Notes, doc.UpdatedAt)
	if err != nil {
		return err
	}
	if replaceLines {
		if _, err = tx.Exec(ctx, `DELETE FROM movement_document_lines WHERE document_id = $1`, doc.ID); err != nil {
			return err
		}
		for _, line := range doc.Lines {
			_, err = tx.Exec(ctx, `
				INSERT INTO movement_document_lines (
					id, document_id, part_id, warehouse_id, destination_warehouse_id,
					quantity, unit_cost, reference_line_id, notes, sort_order, created_at
				) VALUES ($1,$2,$3,$4,$5,$6,$7::numeric,$8,$9,$10,$11)
			`, line.ID, doc.ID, line.PartID, line.WarehouseID, line.DestinationWarehouseID,
				line.Quantity, line.UnitCost, line.ReferenceLineID, line.Notes, line.SortOrder, line.CreatedAt)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

// ExtractedQuantityByParentLine — сколько уже извлечено по закрытым документам from_production.
func (r *MovementDocumentRepository) ExtractedQuantityByParentLine(ctx context.Context, parentDocumentID, parentLineID uuid.UUID) (int32, error) {
	var qty int32
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(l.quantity), 0)::int
		FROM movement_documents d
		JOIN movement_document_lines l ON l.document_id = d.id
		WHERE d.parent_document_id = $1
		  AND d.movement_type = 'from_production'
		  AND d.status = 'closed'
		  AND l.reference_line_id = $2
	`, parentDocumentID, parentLineID).Scan(&qty)
	return qty, err
}

func (r *MovementDocumentRepository) HasOpenExtractionForParent(ctx context.Context, parentDocumentID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM movement_documents
			WHERE parent_document_id = $1
			  AND movement_type = 'from_production'
			  AND status IN ('draft', 'in_progress')
		)
	`, parentDocumentID).Scan(&exists)
	return exists, err
}
