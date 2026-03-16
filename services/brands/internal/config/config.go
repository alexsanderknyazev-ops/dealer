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
		GRPCPort:    getEnvInt("BRANDS_GRPC_PORT", 50056),
		HTTPPort:    getEnvInt("BRANDS_HTTP_PORT", 8085),
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://dealer:dealer_secret@127.0.0.1:5433/dealer?sslmode=disable"),
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
