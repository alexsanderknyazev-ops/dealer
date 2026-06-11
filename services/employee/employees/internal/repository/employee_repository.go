package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/employees/internal/domain"
)

type EmployeeRepository struct {
	pool *pgxpool.Pool
}

func NewEmployeeRepository(pool *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{pool: pool}
}

const employeeSelect = `
	SELECT id, user_id, full_name, position, department, phone, active, created_at, updated_at
	FROM employees
`

func scanEmployee(row interface {
	Scan(dest ...any) error
}) (*domain.Employee, error) {
	var e domain.Employee
	err := row.Scan(&e.ID, &e.UserID, &e.FullName, &e.Position, &e.Department, &e.Phone, &e.Active, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepository) Create(ctx context.Context, e *domain.Employee) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO employees (id, user_id, full_name, position, department, phone, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, e.ID, e.UserID, e.FullName, e.Position, e.Department, e.Phone, e.Active, e.CreatedAt, e.UpdatedAt)
	return err
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	return scanEmployee(r.pool.QueryRow(ctx, employeeSelect+" WHERE id = $1", id))
}

func (r *EmployeeRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Employee, error) {
	return scanEmployee(r.pool.QueryRow(ctx, employeeSelect+" WHERE user_id = $1", userID))
}

func (r *EmployeeRepository) ResolveRef(ctx context.Context, ref uuid.UUID) (*domain.Employee, error) {
	e, err := r.GetByID(ctx, ref)
	if err == nil {
		return e, nil
	}
	return r.GetByUserID(ctx, ref)
}

func (r *EmployeeRepository) List(ctx context.Context, limit, offset int32, search, position string, activeOnly bool) ([]*domain.Employee, int32, error) {
	countQuery := "SELECT COUNT(*) FROM employees WHERE 1=1"
	listQuery := employeeSelect + " WHERE 1=1"
	args := []any{}
	argNum := 1
	if search != "" {
		pattern := "%" + search + "%"
		clause := fmt.Sprintf(" AND (full_name ILIKE $%d OR position ILIKE $%d OR department ILIKE $%d)", argNum, argNum, argNum)
		countQuery += clause
		listQuery += clause
		args = append(args, pattern)
		argNum++
	}
	if position != "" {
		countQuery += fmt.Sprintf(" AND position = $%d", argNum)
		listQuery += fmt.Sprintf(" AND position = $%d", argNum)
		args = append(args, position)
		argNum++
	}
	if activeOnly {
		countQuery += " AND active = true"
		listQuery += " AND active = true"
	}
	var total int32
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQuery += " ORDER BY full_name LIMIT $" + fmt.Sprint(argNum) + " OFFSET $" + fmt.Sprint(argNum+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.Employee
	for rows.Next() {
		e, err := scanEmployee(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, nil
}

func (r *EmployeeRepository) Update(ctx context.Context, e *domain.Employee) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE employees
		SET user_id=$2, full_name=$3, position=$4, department=$5, phone=$6, active=$7, updated_at=$8
		WHERE id=$1
	`, e.ID, e.UserID, e.FullName, e.Position, e.Department, e.Phone, e.Active, e.UpdatedAt)
	return err
}

func (r *EmployeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM employees WHERE id = $1`, id)
	return err
}
