package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/customers-service/internal/domain"
)

type CustomerRepository struct {
	pool *pgxpool.Pool
}

func NewCustomerRepository(pool *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{pool: pool}
}

func (r *CustomerRepository) Create(ctx context.Context, c *domain.Customer) error {
	query := `
		INSERT INTO customers (id, name, email, phone, customer_type, inn, address, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		c.ID, c.Name, c.Email, c.Phone, c.CustomerType, c.INN, c.Address, c.Notes, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	query := `
		SELECT id, name, email, phone, customer_type, inn, address, notes, created_at, updated_at
		FROM customers WHERE id = $1
	`
	var c domain.Customer
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Email, &c.Phone, &c.CustomerType, &c.INN, &c.Address, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) List(ctx context.Context, limit, offset int32, search string) ([]*domain.Customer, int32, error) {
	searchPattern := "%" + search + "%"
	var total int32
	if search != "" {
		err := r.pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM customers WHERE name ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1",
			searchPattern,
		).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM customers").Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	var rows pgx.Rows
	var err error
	if search != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, name, email, phone, customer_type, inn, address, notes, created_at, updated_at
			FROM customers WHERE name ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`, searchPattern, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, name, email, phone, customer_type, inn, address, notes, created_at, updated_at
			FROM customers ORDER BY created_at DESC LIMIT $1 OFFSET $2
		`, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*domain.Customer
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.CustomerType, &c.INN, &c.Address, &c.Notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &c)
	}
	return list, total, nil
}

func (r *CustomerRepository) Update(ctx context.Context, c *domain.Customer) error {
	query := `
		UPDATE customers SET name=$2, email=$3, phone=$4, customer_type=$5, inn=$6, address=$7, notes=$8, updated_at=$9
		WHERE id=$1
	`
	_, err := r.pool.Exec(ctx, query, c.ID, c.Name, c.Email, c.Phone, c.CustomerType, c.INN, c.Address, c.Notes, c.UpdatedAt)
	return err
}

func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM customers WHERE id = $1", id)
	return err
}
