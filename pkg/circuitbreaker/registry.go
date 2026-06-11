package circuitbreaker

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	defaultRegistry     *Registry
	defaultRegistryOnce sync.Once
)

// DefaultConfig читает пороги из env (с разумными дефолтами).
func DefaultConfig() Config {
	cfg := Config{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		OpenTimeout:      30 * time.Second,
	}
	if v := strings.TrimSpace(os.Getenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.FailureThreshold = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.SuccessThreshold = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CIRCUIT_BREAKER_OPEN_TIMEOUT")); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.OpenTimeout = d
		}
	}
	return cfg
}

// Enabled — CIRCUIT_BREAKER_ENABLED (default true).
func Enabled() bool {
	v := strings.TrimSpace(os.Getenv("CIRCUIT_BREAKER_ENABLED"))
	if v == "" {
		return true
	}
	ok, err := strconv.ParseBool(v)
	return err == nil && ok
}

// Registry хранит breaker на каждый upstream (gRPC target).
type Registry struct {
	mu   sync.Mutex
	cfg  Config
	pool map[string]*Breaker
}

func NewRegistry(cfg Config) *Registry {
	return &Registry{cfg: cfg.normalized(), pool: make(map[string]*Breaker)}
}

func DefaultRegistry() *Registry {
	defaultRegistryOnce.Do(func() {
		defaultRegistry = NewRegistry(DefaultConfig())
	})
	return defaultRegistry
}

func (r *Registry) Get(name string) *Breaker {
	if name == "" {
		name = "unknown"
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if br, ok := r.pool[name]; ok {
		return br
	}
	br := New(r.cfg)
	r.pool[name] = br
	return br
}
