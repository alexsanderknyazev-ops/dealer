package config

import (
	"os"
	"strconv"
)

type Config struct {
	GRPCPort          int
	HTTPPort          int
	PostgresDSN       string
	JWTSecret         string
	CustomersGRPCAddr string
	VehiclesGRPCAddr  string
}

func Load() *Config {
	return &Config{
		GRPCPort:          getEnvInt("DEALS_GRPC_PORT", 50054),
		HTTPPort:          getEnvInt("DEALS_HTTP_PORT", 8083),
		PostgresDSN:       getEnv("POSTGRES_DSN", ""),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		CustomersGRPCAddr: getEnv("CUSTOMERS_GRPC_ADDR", ""),
		VehiclesGRPCAddr:  getEnv("VEHICLES_GRPC_ADDR", ""),
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
