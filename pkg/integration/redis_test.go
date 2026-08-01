//go:build integration

package integration

import (
	"context"
	"net"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/redis"

	pkgredis "github.com/dealer/dealer/pkg/redis"
)

func redisAddr(ctx context.Context, container *redis.RedisContainer) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return "", err
	}
	return net.JoinHostPort(host, port.Port()), nil
}

func TestRedis_Ping(t *testing.T) {
	container := requireRedis(t)
	ctx := context.Background()

	addr, err := redisAddr(ctx, container)
	if err != nil {
		t.Fatal(err)
	}

	rdb := pkgredis.NewClient(addr, "", 0)
	defer rdb.Close()

	if err := pkgredis.Ping(ctx, rdb); err != nil {
		t.Fatalf("redis ping: %v", err)
	}
}

func TestRedis_SetGet(t *testing.T) {
	container := requireRedis(t)
	ctx := context.Background()

	addr, err := redisAddr(ctx, container)
	if err != nil {
		t.Fatal(err)
	}

	rdb := pkgredis.NewClient(addr, "", 0)
	defer rdb.Close()

	key := "integration:key:1"
	want := "integration-value-1"
	if err := rdb.Set(ctx, key, want, 0).Err(); err != nil {
		t.Fatalf("redis set: %v", err)
	}

	got, err := rdb.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("redis get: %v", err)
	}
	if got != want {
		t.Fatalf("redis roundtrip mismatch: got %q want %q", got, want)
	}
}
