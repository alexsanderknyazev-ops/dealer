package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/workorders/internal/domain"
)

type WorkOrderRepository struct {
	pool *pgxpool.Pool
}

func NewWorkOrderRepository(pool *pgxpool.Pool) *WorkOrderRepository {
	return &WorkOrderRepository{pool: pool}
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *WorkOrderRepository) NextOrderNumber(ctx context.Context) (string, error) {
	var n int64
	if err := r.pool.QueryRow(ctx, "SELECT nextval('work_orders_number_seq')").Scan(&n); err != nil {
		return "", err
	}
	return fmt.Sprintf("WO-%d-%05d", time.Now().Year(), n), nil
}

func (r *WorkOrderRepository) Create(ctx context.Context, wo *domain.WorkOrder) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	query := `
		INSERT INTO work_orders (
			id, order_number, customer_id, vehicle_id, dealer_point_id, warehouse_id,
			repair_type, status, service_advisor_id, complaint, diagnosis, mileage_km,
			labor_cost, parts_cost, total_cost, opened_at, closed_at, parts_issued, parts_issued_at,
			movement_document_id, movement_document_status, source_order_type, source_order_id,
			notes, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13::numeric,$14::numeric,$15::numeric,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)
	`
	_, err = tx.Exec(ctx, query,
		wo.ID, wo.OrderNumber, wo.CustomerID, wo.VehicleID, wo.DealerPointID, wo.WarehouseID,
		wo.RepairType, wo.Status, wo.ServiceAdvisorID, wo.Complaint, wo.Diagnosis, wo.MileageKm,
		wo.LaborCost, wo.PartsCost, wo.TotalCost, wo.OpenedAt, wo.ClosedAt, wo.PartsIssued, wo.PartsIssuedAt,
		wo.MovementDocumentID, wo.MovementDocumentStatus, nullIfEmpty(wo.SourceOrderType), wo.SourceOrderID,
		wo.Notes, wo.CreatedAt, wo.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if err := insertLabor(ctx, tx, wo.ID, wo.Labor); err != nil {
		return err
	}
	if err := insertParts(ctx, tx, wo.ID, wo.Parts); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func insertLabor(ctx context.Context, tx pgx.Tx, workOrderID uuid.UUID, rows []domain.WorkOrderLabor) error {
	for _, row := range rows {
		_, err := tx.Exec(ctx, `
			INSERT INTO work_order_labor (id, work_order_id, work_id, description, quantity, unit_price, amount, executor_id, sort_order, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5::numeric,$6::numeric,$7::numeric,$8,$9,$10,$11)
		`, row.ID, workOrderID, row.WorkID, row.Description, row.Quantity, row.UnitPrice, row.Amount, row.ExecutorID, row.SortOrder, row.CreatedAt, row.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertParts(ctx context.Context, tx pgx.Tx, workOrderID uuid.UUID, rows []domain.WorkOrderPart) error {
	for _, row := range rows {
		_, err := tx.Exec(ctx, `
			INSERT INTO work_order_parts (id, work_order_id, part_id, warehouse_id, description, quantity, unit_price, amount, issued, sort_order, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6::numeric,$7::numeric,$8::numeric,$9,$10,$11,$12)
		`, row.ID, workOrderID, row.PartID, row.WarehouseID, row.Description, row.Quantity, row.UnitPrice, row.Amount, row.Issued, row.SortOrder, row.CreatedAt, row.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *WorkOrderRepository) scanHeader(row pgx.Row) (*domain.WorkOrder, error) {
	var wo domain.WorkOrder
	err := row.Scan(
		&wo.ID, &wo.OrderNumber, &wo.CustomerID, &wo.VehicleID, &wo.DealerPointID, &wo.WarehouseID,
		&wo.RepairType, &wo.Status, &wo.ServiceAdvisorID, &wo.Complaint, &wo.Diagnosis, &wo.MileageKm,
		&wo.LaborCost, &wo.PartsCost, &wo.TotalCost, &wo.OpenedAt, &wo.ClosedAt, &wo.PartsIssued, &wo.PartsIssuedAt,
		&wo.MovementDocumentID, &wo.MovementDocumentStatus, &wo.SourceOrderType, &wo.SourceOrderID,
		&wo.Notes, &wo.CreatedAt, &wo.UpdatedAt,
	)
	return &wo, err
}

const headerSelect = `
	SELECT id, order_number, customer_id, vehicle_id, dealer_point_id, warehouse_id,
		repair_type, status, service_advisor_id, complaint, diagnosis, mileage_km,
		labor_cost::text, parts_cost::text, total_cost::text, opened_at, closed_at,
		parts_issued, parts_issued_at, movement_document_id, movement_document_status,
		source_order_type, source_order_id, notes, created_at, updated_at
	FROM work_orders
`

func (r *WorkOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkOrder, error) {
	wo, err := r.scanHeader(r.pool.QueryRow(ctx, headerSelect+" WHERE id = $1", id))
	if err != nil {
		return nil, err
	}
	wo.Labor, err = r.listLabor(ctx, id)
	if err != nil {
		return nil, err
	}
	wo.Parts, err = r.listParts(ctx, id)
	return wo, err
}

func (r *WorkOrderRepository) listLabor(ctx context.Context, workOrderID uuid.UUID) ([]domain.WorkOrderLabor, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, work_order_id, work_id, description, quantity::text, unit_price::text, amount::text, executor_id, sort_order, created_at, updated_at
		FROM work_order_labor WHERE work_order_id = $1 ORDER BY sort_order, created_at
	`, workOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WorkOrderLabor
	for rows.Next() {
		var l domain.WorkOrderLabor
		if err := rows.Scan(&l.ID, &l.WorkOrderID, &l.WorkID, &l.Description, &l.Quantity, &l.UnitPrice, &l.Amount, &l.ExecutorID, &l.SortOrder, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

func (r *WorkOrderRepository) listParts(ctx context.Context, workOrderID uuid.UUID) ([]domain.WorkOrderPart, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, work_order_id, part_id, warehouse_id, description, quantity::text, unit_price::text, amount::text, issued, sort_order, created_at, updated_at
		FROM work_order_parts WHERE work_order_id = $1 ORDER BY sort_order, created_at
	`, workOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WorkOrderPart
	for rows.Next() {
		var p domain.WorkOrderPart
		if err := rows.Scan(&p.ID, &p.WorkOrderID, &p.PartID, &p.WarehouseID, &p.Description, &p.Quantity, &p.UnitPrice, &p.Amount, &p.Issued, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *WorkOrderRepository) List(ctx context.Context, limit, offset int32, status, repairType, customerID, vehicleID string) ([]*domain.WorkOrder, int32, error) {
	countQuery := "SELECT COUNT(*) FROM work_orders WHERE 1=1"
	listQuery := headerSelect + " WHERE 1=1"
	args := []any{}
	argNum := 1
	addFilter := func(clause string, val any) {
		countQuery += fmt.Sprintf(" AND %s = $%d", clause, argNum)
		listQuery += fmt.Sprintf(" AND %s = $%d", clause, argNum)
		args = append(args, val)
		argNum++
	}
	if status != "" {
		addFilter("status", status)
	}
	if repairType != "" {
		addFilter("repair_type", repairType)
	}
	if customerID != "" {
		if cid, err := uuid.Parse(customerID); err == nil {
			addFilter("customer_id", cid)
		}
	}
	if vehicleID != "" {
		if vid, err := uuid.Parse(vehicleID); err == nil {
			addFilter("vehicle_id", vid)
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
	var list []*domain.WorkOrder
	for rows.Next() {
		wo, err := r.scanHeader(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, wo)
	}
	return list, total, nil
}

func (r *WorkOrderRepository) Update(ctx context.Context, wo *domain.WorkOrder, replaceLines bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		UPDATE work_orders SET
			customer_id=$2, vehicle_id=$3, dealer_point_id=$4, warehouse_id=$5,
			repair_type=$6, status=$7, service_advisor_id=$8, complaint=$9, diagnosis=$10, mileage_km=$11,
			labor_cost=$12::numeric, parts_cost=$13::numeric, total_cost=$14::numeric,
			opened_at=$15, closed_at=$16, parts_issued=$17, parts_issued_at=$18,
			movement_document_id=$19, movement_document_status=$20, notes=$21, updated_at=$22
		WHERE id=$1
	`, wo.ID, wo.CustomerID, wo.VehicleID, wo.DealerPointID, wo.WarehouseID,
		wo.RepairType, wo.Status, wo.ServiceAdvisorID, wo.Complaint, wo.Diagnosis, wo.MileageKm,
		wo.LaborCost, wo.PartsCost, wo.TotalCost, wo.OpenedAt, wo.ClosedAt, wo.PartsIssued, wo.PartsIssuedAt,
		wo.MovementDocumentID, wo.MovementDocumentStatus, wo.Notes, wo.UpdatedAt)
	if err != nil {
		return err
	}
	if replaceLines {
		if _, err = tx.Exec(ctx, "DELETE FROM work_order_labor WHERE work_order_id = $1", wo.ID); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "DELETE FROM work_order_parts WHERE work_order_id = $1 AND issued = false", wo.ID); err != nil {
			return err
		}
		if err = insertLabor(ctx, tx, wo.ID, wo.Labor); err != nil {
			return err
		}
		unissued := make([]domain.WorkOrderPart, 0, len(wo.Parts))
		for _, p := range wo.Parts {
			if !p.Issued {
				unissued = append(unissued, p)
			}
		}
		if err = insertParts(ctx, tx, wo.ID, unissued); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *WorkOrderRepository) SetMovementDocument(ctx context.Context, workOrderID, documentID uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE work_orders
		SET movement_document_id = $2, movement_document_status = $3, updated_at = now()
		WHERE id = $1
	`, workOrderID, documentID, status)
	return err
}

func (r *WorkOrderRepository) MarkPartsIssued(ctx context.Context, workOrderID uuid.UUID, lineIDs []uuid.UUID, issuedAt time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, lineID := range lineIDs {
		_, err = tx.Exec(ctx, `UPDATE work_order_parts SET issued = true, updated_at = $3 WHERE id = $1 AND work_order_id = $2`, lineID, workOrderID, issuedAt)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `
		UPDATE work_orders
		SET parts_issued = true, parts_issued_at = $2, movement_document_status = 'closed', updated_at = $2
		WHERE id = $1
	`, workOrderID, issuedAt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *WorkOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM work_orders WHERE id = $1", id)
	return err
}
