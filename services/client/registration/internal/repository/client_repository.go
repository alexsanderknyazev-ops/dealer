package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/client-registration-service/internal/domain"
)

var ErrVINAlreadyLinked = errors.New("vin already linked to another client")

type ClientRepository struct {
	pool *pgxpool.Pool
}

func NewClientRepository(pool *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{pool: pool}
}

func (r *ClientRepository) CreateClientWithVehicle(ctx context.Context, c *domain.Client, v *domain.ClientVehicle) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.insertClient(ctx, tx, c); err != nil {
		return err
	}
	v.ClientID = c.ID
	if err := r.insertVehicle(ctx, tx, v); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *ClientRepository) insertClient(ctx context.Context, tx pgx.Tx, c *domain.Client) error {
	query := `
		INSERT INTO clients (id, user_id, email, full_name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.Exec(ctx, query, c.ID, c.UserID, c.Email, c.FullName, c.Phone, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *ClientRepository) insertVehicle(ctx context.Context, tx pgx.Tx, v *domain.ClientVehicle) error {
	query := `
		INSERT INTO client_vehicles (id, client_id, vehicle_id, vin, make, model, year, added_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := tx.Exec(ctx, query, v.ID, v.ClientID, v.VehicleID, v.VIN, v.Make, v.Model, v.Year, v.AddedAt)
	if err != nil && isUniqueViolation(err) {
		return ErrVINAlreadyLinked
	}
	return err
}

func (r *ClientRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Client, error) {
	query := `
		SELECT id, user_id, email, full_name, phone, created_at, updated_at
		FROM clients WHERE user_id = $1
	`
	var c domain.Client
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&c.ID, &c.UserID, &c.Email, &c.FullName, &c.Phone, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ClientRepository) ListVehiclesByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientVehicle, error) {
	query := `
		SELECT cv.id, cv.client_id, cv.vehicle_id, cv.vin, cv.make, cv.model, cv.year, cv.added_at
		FROM client_vehicles cv
		WHERE cv.client_id = $1
		ORDER BY cv.added_at DESC
	`
	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.ClientVehicle
	for rows.Next() {
		var v domain.ClientVehicle
		if err := rows.Scan(&v.ID, &v.ClientID, &v.VehicleID, &v.VIN, &v.Make, &v.Model, &v.Year, &v.AddedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}

func (r *ClientRepository) AddVehicle(ctx context.Context, clientID uuid.UUID, v *domain.ClientVehicle) error {
	v.ID = uuid.New()
	v.ClientID = clientID
	if v.AddedAt.IsZero() {
		v.AddedAt = time.Now().UTC()
	}
	query := `
		INSERT INTO client_vehicles (id, client_id, vehicle_id, vin, make, model, year, added_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query, v.ID, v.ClientID, v.VehicleID, v.VIN, v.Make, v.Model, v.Year, v.AddedAt)
	if err != nil && isUniqueViolation(err) {
		return ErrVINAlreadyLinked
	}
	return err
}

func (r *ClientRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT 1 FROM clients WHERE email = $1`
	var one int
	err := r.pool.QueryRow(ctx, query, email).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ClientRepository) VINLinked(ctx context.Context, vin string) (bool, error) {
	query := `SELECT 1 FROM client_vehicles WHERE vin = $1`
	var one int
	err := r.pool.QueryRow(ctx, query, vin).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}
