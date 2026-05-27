package config

import (
	"os"
	"strconv"
)

type Config struct {
	GRPCPort    int
	HTTPPort    int
	PostgresDSN string
	JWTSecret   string
}

func Load() *Config {
	return &Config{
		GRPCPort:    getEnvInt("VEHICLES_GRPC_PORT", 50053),
		HTTPPort:    getEnvInt("VEHICLES_HTTP_PORT", 8082),
		PostgresDSN: getEnv("POSTGRES_DSN", ""),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
