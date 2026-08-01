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

type SupplierOrderRepository struct {
	pool *pgxpool.Pool
}

func NewSupplierOrderRepository(pool *pgxpool.Pool) *SupplierOrderRepository {
	return &SupplierOrderRepository{pool: pool}
}

func (r *SupplierOrderRepository) NextOrderNumber(ctx context.Context) (string, error) {
	var n int64
	if err := r.pool.QueryRow(ctx, "SELECT nextval('supplier_orders_number_seq')").Scan(&n); err != nil {
		return "", err
	}
	return fmt.Sprintf("SO-%d-%05d", time.Now().Year(), n), nil
}

const supplierOrderSelect = `
	SELECT id, order_number, status, supplier_id, receipt_warehouse_id, customer_order_id,
	       fulfillment_movement_document_id, fulfillment_work_order_id, notes, created_by, created_at, updated_at
	FROM supplier_orders
`

func scanSupplierOrder(row pgx.Row) (*domain.SupplierOrder, error) {
	var o domain.SupplierOrder
	err := row.Scan(
		&o.ID, &o.OrderNumber, &o.Status, &o.SupplierID, &o.ReceiptWarehouseID, &o.CustomerOrderID,
		&o.FulfillmentMovementDocumentID, &o.FulfillmentWorkOrderID, &o.Notes, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
	)
	return &o, err
}

func (r *SupplierOrderRepository) listLines(ctx context.Context, orderID uuid.UUID) ([]domain.PartOrderLine, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, part_id, quantity, unit_price::text, notes, sort_order, created_at
		FROM supplier_order_lines WHERE order_id = $1 ORDER BY sort_order, created_at
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PartOrderLine
	for rows.Next() {
		var l domain.PartOrderLine
		if err := rows.Scan(&l.ID, &l.OrderID, &l.PartID, &l.Quantity, &l.UnitPrice, &l.Notes, &l.SortOrder, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

func (r *SupplierOrderRepository) Create(ctx context.Context, o *domain.SupplierOrder) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, `
		INSERT INTO supplier_orders (
			id, order_number, status, supplier_id, receipt_warehouse_id, customer_order_id, notes, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, o.ID, o.OrderNumber, o.Status, o.SupplierID, o.ReceiptWarehouseID, o.CustomerOrderID, o.Notes, o.CreatedBy, o.CreatedAt, o.UpdatedAt)
	if err != nil {
		return err
	}
	for _, line := range o.Lines {
		_, err = tx.Exec(ctx, `
			INSERT INTO supplier_order_lines (id, order_id, part_id, quantity, unit_price, notes, sort_order, created_at)
			VALUES ($1,$2,$3,$4,$5::numeric,$6,$7,$8)
		`, line.ID, o.ID, line.PartID, line.Quantity, line.UnitPrice, line.Notes, line.SortOrder, line.CreatedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *SupplierOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SupplierOrder, error) {
	o, err := scanSupplierOrder(r.pool.QueryRow(ctx, supplierOrderSelect+" WHERE id = $1", id))
	if err != nil {
		return nil, err
	}
	o.Lines, err = r.listLines(ctx, id)
	return o, err
}

func (r *SupplierOrderRepository) List(ctx context.Context, limit, offset int32, status string) ([]*domain.SupplierOrder, int32, error) {
	countQ := "SELECT COUNT(*) FROM supplier_orders WHERE 1=1"
	listQ := supplierOrderSelect + " WHERE 1=1"
	args := []any{}
	n := 1
	if status != "" {
		countQ += fmt.Sprintf(" AND status = $%d", n)
		listQ += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	var total int32
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQ += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.SupplierOrder
	for rows.Next() {
		o, err := scanSupplierOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, o)
	}
	return list, total, nil
}

func (r *SupplierOrderRepository) Update(ctx context.Context, o *domain.SupplierOrder, replaceLines bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, `
		UPDATE supplier_orders
		SET supplier_id=$2, receipt_warehouse_id=$3, customer_order_id=$4, notes=$5, updated_at=$6
		WHERE id=$1
	`, o.ID, o.SupplierID, o.ReceiptWarehouseID, o.CustomerOrderID, o.Notes, o.UpdatedAt)
	if err != nil {
		return err
	}
	if replaceLines {
		if _, err = tx.Exec(ctx, `DELETE FROM supplier_order_lines WHERE order_id = $1`, o.ID); err != nil {
			return err
		}
		for _, line := range o.Lines {
			_, err = tx.Exec(ctx, `
				INSERT INTO supplier_order_lines (id, order_id, part_id, quantity, unit_price, notes, sort_order, created_at)
				VALUES ($1,$2,$3,$4,$5::numeric,$6,$7,$8)
			`, line.ID, o.ID, line.PartID, line.Quantity, line.UnitPrice, line.Notes, line.SortOrder, line.CreatedAt)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (r *SupplierOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, movementDocID *uuid.UUID, updatedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE supplier_orders
		SET status=$2, fulfillment_movement_document_id=COALESCE($3, fulfillment_movement_document_id), updated_at=$4
		WHERE id=$1
	`, id, status, movementDocID, updatedAt)
	return err
}

func (r *SupplierOrderRepository) LinkWorkOrder(ctx context.Context, id uuid.UUID, workOrderID uuid.UUID, updatedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE supplier_orders
		SET status='linked', fulfillment_work_order_id=$2, updated_at=$3
		WHERE id=$1
	`, id, workOrderID, updatedAt)
	return err
}

func (r *SupplierOrderRepository) MarkFulfilled(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE supplier_orders SET status='fulfilled', updated_at=now() WHERE id=$1 AND status='linked'
	`, id)
	return err
}
