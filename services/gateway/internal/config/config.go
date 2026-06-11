package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort             int
	AuthGRPCAddr         string
	CustomersGRPCAddr    string
	VehiclesGRPCAddr     string
	DealsGRPCAddr        string
	PartsGRPCAddr        string
	BrandsGRPCAddr       string
	DealerPointsGRPCAddr string
}

func Load() *Config {
	return &Config{
		HTTPPort:             getEnvInt("GATEWAY_HTTP_PORT", 8090),
		AuthGRPCAddr:         getEnv("AUTH_GRPC_ADDR", "127.0.0.1:50051"),
		CustomersGRPCAddr:    getEnv("CUSTOMERS_GRPC_ADDR", "127.0.0.1:50052"),
		VehiclesGRPCAddr:     getEnv("VEHICLES_GRPC_ADDR", "127.0.0.1:50053"),
		DealsGRPCAddr:        getEnv("DEALS_GRPC_ADDR", "127.0.0.1:50054"),
		PartsGRPCAddr:        getEnv("PARTS_GRPC_ADDR", "127.0.0.1:50055"),
		BrandsGRPCAddr:       getEnv("BRANDS_GRPC_ADDR", "127.0.0.1:50056"),
		DealerPointsGRPCAddr: getEnv("DEALER_POINTS_GRPC_ADDR", "127.0.0.1:50057"),
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
