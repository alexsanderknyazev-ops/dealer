package testcontainers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Postgres — запущенный контейнер PostgreSQL с применёнными миграциями.
type Postgres struct {
	Container *tcpostgres.PostgresContainer
	Pool      *pgxpool.Pool
	DSN       string

	closeOnce sync.Once
}

// StartPostgres запускает контейнер Postgres и применяет все миграции
// репозитория. Соединение по умолчанию доступно через поле Pool.
func StartPostgres(ctx context.Context) (*Postgres, error) {
	container, err := tcpostgres.Run(ctx, postgresImage,
		tcpostgres.WithDatabase("dealer"),
		tcpostgres.WithUsername("dealer"),
		tcpostgres.WithPassword("dealer"),
		testcontainers.WithWaitStrategyAndDeadline(3*time.Minute,
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(3*time.Minute)),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres: %w", err)
	}
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("postgres connection string: %w", err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := ApplyMigrations(ctx, pool); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}
	return &Postgres{Container: container, Pool: pool, DSN: dsn}, nil
}

// Close освобождает ресурсы. Безопасен для многократного вызова.
func (p *Postgres) Close(ctx context.Context) error {
	if p == nil {
		return nil
	}
	p.closeOnce.Do(func() {
		p.Pool.Close()
		_ = p.Container.Terminate(ctx)
	})
	return nil
}
