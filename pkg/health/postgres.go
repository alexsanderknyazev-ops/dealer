package health

import "context"

// Pinger — пул Postgres (pgxpool.Pool).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Postgres возвращает Check для /readyz.
func Postgres(pool Pinger) Check {
	if pool == nil {
		return nil
	}
	return func(ctx context.Context) error {
		return pool.Ping(ctx)
	}
}
