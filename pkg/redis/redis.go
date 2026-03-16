package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// NewClient создаёт Redis-клиент.
func NewClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// Ping проверяет соединение с Redis.
func Ping(ctx context.Context, rdb *redis.Client) error {
	return rdb.Ping(ctx).Err()
}
