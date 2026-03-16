package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool создаёт пул подключений к PostgreSQL.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}
