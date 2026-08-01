package testcontainers

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/client"
)

var (
	dockerOnce sync.Once
	dockerOK   bool
)

// DockerAvailable сообщает, доступен ли Docker daemon.
func DockerAvailable() bool {
	dockerOnce.Do(func() {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			dockerOK = false
			return
		}
		defer cli.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = cli.Ping(ctx)
		dockerOK = err == nil
	})
	return dockerOK
}

// SkipIfNoDocker пропускает тест, если Docker недоступен.
func SkipIfNoDocker(t *testing.T) {
	t.Helper()
	if !DockerAvailable() {
		t.Skip("docker daemon unavailable; skipping testcontainers test")
	}
}
