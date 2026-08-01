//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresImage = "postgres:16-alpine"
	redisImage    = "redis:7-alpine"
	kafkaImage    = "confluentinc/confluent-local:7.5.0"
)

var (
	pgOnce    sync.Once
	pgInst    *pgInstance
	pgErr     error
	redisOnce sync.Once
	redisInst *tcredis.RedisContainer
	redisErr  error
	kafkaOnce sync.Once
	kafkaInst *kafka.KafkaContainer
	kafkaErr  error
	termOnce  sync.Once
)

type pgInstance struct {
	Container *tcpostgres.PostgresContainer
	Pool      *pgxpool.Pool
	DSN       string
}

func requireDocker(t *testing.T) {
	t.Helper()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skipf("docker client unavailable: %v", err)
	}
	defer cli.Close()
	if _, err := cli.Ping(context.Background()); err != nil {
		t.Skipf("docker daemon unavailable: %v", err)
	}
}

func requirePostgres(t *testing.T) *pgInstance {
	t.Helper()
	requireDocker(t)
	t.Cleanup(terminateAll)
	pgOnce.Do(func() { pgInst, pgErr = startPostgres() })
	if pgErr != nil {
		t.Fatal(pgErr)
	}
	return pgInst
}

func requireRedis(t *testing.T) *tcredis.RedisContainer {
	t.Helper()
	requireDocker(t)
	t.Cleanup(terminateAll)
	redisOnce.Do(func() { redisInst, redisErr = startRedis() })
	if redisErr != nil {
		t.Fatal(redisErr)
	}
	return redisInst
}

func requireKafka(t *testing.T) *kafka.KafkaContainer {
	t.Helper()
	requireDocker(t)
	t.Cleanup(terminateAll)
	kafkaOnce.Do(func() { kafkaInst, kafkaErr = startKafka() })
	if kafkaErr != nil {
		t.Fatal(kafkaErr)
	}
	return kafkaInst
}

func startPostgres() (*pgInstance, error) {
	ctx := context.Background()
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
	if err := applyMigrations(ctx, pool); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}
	return &pgInstance{Container: container, Pool: pool, DSN: dsn}, nil
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("no migrations found in ../../migrations")
	}
	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		conn, err := pool.Acquire(ctx)
		if err != nil {
			return err
		}
		_, execErr := conn.Exec(ctx, string(sql), pgx.QueryExecModeSimpleProtocol)
		conn.Release()
		if execErr != nil {
			return fmt.Errorf("%s: %w", filepath.Base(f), execErr)
		}
	}
	return nil
}

func startRedis() (*tcredis.RedisContainer, error) {
	ctx := context.Background()
	container, err := tcredis.Run(ctx, redisImage,
		testcontainers.WithWaitStrategyAndDeadline(2*time.Minute,
			wait.ForLog("Ready to accept connections").WithStartupTimeout(2*time.Minute)),
	)
	if err != nil {
		return nil, fmt.Errorf("start redis: %w", err)
	}
	return container, nil
}

func startKafka() (*kafka.KafkaContainer, error) {
	ctx := context.Background()
	container, err := kafka.Run(ctx, kafkaImage, kafka.WithClusterID("integration"))
	if err != nil {
		return nil, fmt.Errorf("start kafka: %w", err)
	}
	return container, nil
}

func terminateAll() {
	termOnce.Do(func() {
		ctx := context.Background()
		if pgInst != nil {
			pgInst.Pool.Close()
			_ = pgInst.Container.Terminate(ctx)
		}
		if redisInst != nil {
			_ = redisInst.Terminate(ctx)
		}
		if kafkaInst != nil {
			_ = kafkaInst.Terminate(ctx)
		}
	})
}
