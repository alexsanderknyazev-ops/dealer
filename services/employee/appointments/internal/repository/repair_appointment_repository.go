package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/appointments/internal/domain"
)

type RepairAppointmentRepository struct {
	pool *pgxpool.Pool
}

func NewRepairAppointmentRepository(pool *pgxpool.Pool) *RepairAppointmentRepository {
	return &RepairAppointmentRepository{pool: pool}
}

func (r *RepairAppointmentRepository) NextNumber(ctx context.Context) (string, error) {
	var n int64
	if err := r.pool.QueryRow(ctx, "SELECT nextval('repair_appointments_number_seq')").Scan(&n); err != nil {
		return "", err
	}
	return fmt.Sprintf("RA-%d-%05d", time.Now().Year(), n), nil
}

const headerSelect = `
	SELECT id, appointment_number, customer_id, vehicle_id, dealer_point_id, warehouse_id,
	       scheduled_start, scheduled_end, status, work_order_id, complaint, notes, created_by,
	       created_at, updated_at
	FROM repair_appointments
`

func scanHeader(row pgx.Row) (*domain.RepairAppointment, error) {
	var a domain.RepairAppointment
	err := row.Scan(
		&a.ID, &a.AppointmentNumber, &a.CustomerID, &a.VehicleID, &a.DealerPointID, &a.WarehouseID,
		&a.ScheduledStart, &a.ScheduledEnd, &a.Status, &a.WorkOrderID, &a.Complaint, &a.Notes, &a.CreatedBy,
		&a.CreatedAt, &a.UpdatedAt,
	)
	return &a, err
}

func (r *RepairAppointmentRepository) listLabor(ctx context.Context, id uuid.UUID) ([]domain.RepairAppointmentLabor, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, appointment_id, work_id, description, quantity::text, unit_price::text, sort_order, created_at
		FROM repair_appointment_labor WHERE appointment_id = $1 ORDER BY sort_order, created_at
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RepairAppointmentLabor
	for rows.Next() {
		var l domain.RepairAppointmentLabor
		if err := rows.Scan(&l.ID, &l.AppointmentID, &l.WorkID, &l.Description, &l.Quantity, &l.UnitPrice, &l.SortOrder, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

func (r *RepairAppointmentRepository) listParts(ctx context.Context, id uuid.UUID) ([]domain.RepairAppointmentPart, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, appointment_id, part_id, warehouse_id, quantity, unit_price::text, notes, sort_order, created_at
		FROM repair_appointment_parts WHERE appointment_id = $1 ORDER BY sort_order, created_at
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RepairAppointmentPart
	for rows.Next() {
		var p domain.RepairAppointmentPart
		if err := rows.Scan(&p.ID, &p.AppointmentID, &p.PartID, &p.WarehouseID, &p.Quantity, &p.UnitPrice, &p.Notes, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *RepairAppointmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RepairAppointment, error) {
	a, err := scanHeader(r.pool.QueryRow(ctx, headerSelect+" WHERE id = $1", id))
	if err != nil {
		return nil, err
	}
	a.Labor, err = r.listLabor(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Parts, err = r.listParts(ctx, id)
	return a, err
}

func (r *RepairAppointmentRepository) HasOverlap(ctx context.Context, start, end time.Time, excludeID *uuid.UUID) (bool, error) {
	q := `
		SELECT EXISTS(
			SELECT 1 FROM repair_appointments
			WHERE status IN ('draft', 'scheduled', 'in_progress')
			  AND scheduled_start < $2 AND scheduled_end > $1
	`
	args := []any{start, end}
	if excludeID != nil {
		q += ` AND id <> $3`
		args = append(args, *excludeID)
	}
	q += `)`
	var exists bool
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *RepairAppointmentRepository) ListBusyInRange(ctx context.Context, from, to time.Time) ([]domain.RepairAppointment, error) {
	rows, err := r.pool.Query(ctx, headerSelect+`
		WHERE status IN ('draft', 'scheduled', 'in_progress')
		  AND scheduled_start < $2 AND scheduled_end > $1
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RepairAppointment
	for rows.Next() {
		a, err := scanHeader(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, nil
}

func (r *RepairAppointmentRepository) Create(ctx context.Context, a *domain.RepairAppointment) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		INSERT INTO repair_appointments (
			id, appointment_number, customer_id, vehicle_id, dealer_point_id, warehouse_id,
			scheduled_start, scheduled_end, status, complaint, notes, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, a.ID, a.AppointmentNumber, a.CustomerID, a.VehicleID, a.DealerPointID, a.WarehouseID,
		a.ScheduledStart, a.ScheduledEnd, a.Status, a.Complaint, a.Notes, a.CreatedBy, a.CreatedAt, a.UpdatedAt)
	if err != nil {
		return err
	}
	if err := insertLabor(ctx, tx, a.ID, a.Labor); err != nil {
		return err
	}
	if err := insertParts(ctx, tx, a.ID, a.Parts); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func insertLabor(ctx context.Context, tx pgx.Tx, appointmentID uuid.UUID, rows []domain.RepairAppointmentLabor) error {
	for _, row := range rows {
		_, err := tx.Exec(ctx, `
			INSERT INTO repair_appointment_labor (id, appointment_id, work_id, description, quantity, unit_price, sort_order, created_at)
			VALUES ($1,$2,$3,$4,$5::numeric,$6::numeric,$7,$8)
		`, row.ID, appointmentID, row.WorkID, row.Description, row.Quantity, row.UnitPrice, row.SortOrder, row.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertParts(ctx context.Context, tx pgx.Tx, appointmentID uuid.UUID, rows []domain.RepairAppointmentPart) error {
	for _, row := range rows {
		_, err := tx.Exec(ctx, `
			INSERT INTO repair_appointment_parts (id, appointment_id, part_id, warehouse_id, quantity, unit_price, notes, sort_order, created_at)
			VALUES ($1,$2,$3,$4,$5,$6::numeric,$7,$8,$9)
		`, row.ID, appointmentID, row.PartID, row.WarehouseID, row.Quantity, row.UnitPrice, row.Notes, row.SortOrder, row.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RepairAppointmentRepository) Update(ctx context.Context, a *domain.RepairAppointment, replaceLines bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		UPDATE repair_appointments SET
			customer_id=$2, vehicle_id=$3, dealer_point_id=$4, warehouse_id=$5,
			scheduled_start=$6, scheduled_end=$7, status=$8, complaint=$9, notes=$10, updated_at=$11
		WHERE id=$1
	`, a.ID, a.CustomerID, a.VehicleID, a.DealerPointID, a.WarehouseID,
		a.ScheduledStart, a.ScheduledEnd, a.Status, a.Complaint, a.Notes, a.UpdatedAt)
	if err != nil {
		return err
	}
	if replaceLines {
		if _, err = tx.Exec(ctx, `DELETE FROM repair_appointment_labor WHERE appointment_id = $1`, a.ID); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `DELETE FROM repair_appointment_parts WHERE appointment_id = $1`, a.ID); err != nil {
			return err
		}
		if err = insertLabor(ctx, tx, a.ID, a.Labor); err != nil {
			return err
		}
		if err = insertParts(ctx, tx, a.ID, a.Parts); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *RepairAppointmentRepository) SetWorkOrder(ctx context.Context, id, workOrderID uuid.UUID, updatedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE repair_appointments
		SET work_order_id = $2, status = 'in_progress', updated_at = $3
		WHERE id = $1 AND work_order_id IS NULL
	`, id, workOrderID, updatedAt)
	return err
}

func (r *RepairAppointmentRepository) List(ctx context.Context, limit, offset int32, status string, from, to *time.Time) ([]*domain.RepairAppointment, int32, error) {
	countQ := "SELECT COUNT(*) FROM repair_appointments WHERE 1=1"
	listQ := headerSelect + " WHERE 1=1"
	args := []any{}
	n := 1
	if status != "" {
		countQ += fmt.Sprintf(" AND status = $%d", n)
		listQ += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	if from != nil {
		countQ += fmt.Sprintf(" AND scheduled_start >= $%d", n)
		listQ += fmt.Sprintf(" AND scheduled_start >= $%d", n)
		args = append(args, *from)
		n++
	}
	if to != nil {
		countQ += fmt.Sprintf(" AND scheduled_start < $%d", n)
		listQ += fmt.Sprintf(" AND scheduled_start < $%d", n)
		args = append(args, *to)
		n++
	}
	var total int32
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQ += fmt.Sprintf(" ORDER BY scheduled_start ASC LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.RepairAppointment
	for rows.Next() {
		a, err := scanHeader(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, a)
	}
	return list, total, nil
}
