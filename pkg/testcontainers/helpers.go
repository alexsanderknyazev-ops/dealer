package testcontainers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	postgresImage = "postgres:16-alpine"
	redisImage    = "redis:7-alpine"
	kafkaImage    = "confluentinc/confluent-local:7.5.0"
)

// findRepoRoot поднимается от текущего рабочего каталога до каталога с go.work
// (корень репозитория). go test запускается из каталога тестируемого пакета,
// поэтому всегда находится внутри репозитория.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.work not found starting from %s", dir)
		}
		dir = parent
	}
}

func migrationsDir() (string, error) {
	root, err := findRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "migrations"), nil
}

// ApplyMigrations применяет все migrations/*.up.sql по порядку к переданному
// пулу. Использует простой протокол pgx, чтобы поддерживать файлы с
// несколькими SQL-командами и DO-блоками.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("no migrations found in %s", dir)
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
