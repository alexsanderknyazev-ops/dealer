package config

import (
	"os"
	"strconv"
	"time"

	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	HTTPPort       int
	PostgresDSN    string
	PollInterval   time.Duration
	BatchSize      int
}

func Load() *Config {
	httpPort := 8100
	if v := os.Getenv("SCHEDULER_HTTP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			httpPort = p
		}
	}
	interval := 5 * time.Minute
	if v := os.Getenv("SCHEDULER_POLL_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			interval = d
		}
	}
	batch := 100
	if v := os.Getenv("SCHEDULER_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			batch = n
		}
	}
	pj := configenv.LoadPostgresJWT()
	return &Config{
		HTTPPort:     httpPort,
		PostgresDSN:  pj.PostgresDSN,
		PollInterval: interval,
		BatchSize:    batch,
	}
}
