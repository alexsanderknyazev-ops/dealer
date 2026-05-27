package postgres

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool создаёт пул подключений к PostgreSQL.
func NewPool(ctx context.Context, dsn string, searchPath ...string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	if len(searchPath) > 0 {
		cfg.ConnConfig.RuntimeParams["search_path"] = strings.Join(searchPath, ",")
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}
