//go:build integration

package integration

import (
	"context"
	"testing"

	tc "github.com/dealer/dealer/pkg/testcontainers"

	pkgredis "github.com/dealer/dealer/pkg/redis"
)

func TestRedis_Ping(t *testing.T) {
	ctx := context.Background()
	container, err := tc.StartRedis(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = container.Close(ctx) })

	rdb := pkgredis.NewClient(container.Addr, "", 0)
	defer rdb.Close()

	if err := pkgredis.Ping(ctx, rdb); err != nil {
		t.Fatalf("redis ping: %v", err)
	}
}

func TestRedis_SetGet(t *testing.T) {
	ctx := context.Background()
	container, err := tc.StartRedis(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = container.Close(ctx) })

	rdb := pkgredis.NewClient(container.Addr, "", 0)
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
