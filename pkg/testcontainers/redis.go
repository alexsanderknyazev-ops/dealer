package testcontainers

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Redis — запущенный контейнер Redis.
type Redis struct {
	Container *tcredis.RedisContainer
	Addr      string

	closeOnce sync.Once
}

// StartRedis запускает контейнер Redis и возвращает его адрес host:port.
func StartRedis(ctx context.Context) (*Redis, error) {
	container, err := tcredis.Run(ctx, redisImage,
		testcontainers.WithWaitStrategyAndDeadline(2*time.Minute,
			wait.ForLog("Ready to accept connections").WithStartupTimeout(2*time.Minute)),
	)
	if err != nil {
		return nil, fmt.Errorf("start redis: %w", err)
	}
	addr, err := redisAddr(ctx, container)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}
	return &Redis{Container: container, Addr: addr}, nil
}

func redisAddr(ctx context.Context, container *tcredis.RedisContainer) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("redis host: %w", err)
	}
	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return "", fmt.Errorf("redis port: %w", err)
	}
	return net.JoinHostPort(host, port.Port()), nil
}

// Close освобождает ресурсы. Безопасен для многократного вызова.
func (r *Redis) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	r.closeOnce.Do(func() {
		_ = r.Container.Terminate(ctx)
	})
	return nil
}
