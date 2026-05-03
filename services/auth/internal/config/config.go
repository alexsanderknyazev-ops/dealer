package config

import (
	"os"
	"strconv"
	"time"
)

// Config — конфигурация auth-микросервиса.
type Config struct {
	GRPCPort     int
	HTTPPort     int    // HTTP API + фронт для браузера
	StaticDir         string // каталог статики (SPA); пусто = не раздавать
	CustomersServiceURL string // URL customers-service для прокси /api/customers (пусто = не проксировать)
	VehiclesServiceURL  string // URL vehicles-service для прокси /api/vehicles (пусто = не проксировать)
	DealsServiceURL    string // URL deals-service для прокси /api/deals (пусто = не проксировать)
	PartsServiceURL    string // URL parts-service для прокси /api/parts (пусто = не проксировать)
	BrandsServiceURL   string // URL brands-service для прокси /api/brands (пусто = не проксировать)
	DealerPointsServiceURL string // URL dealer-points-service для прокси /api/dealer-points, /api/legal-entities, /api/warehouses
	PostgresDSN  string
	RedisAddr    string
	RedisPass    string
	RedisDB      int
	KafkaBrokers []string
	KafkaTopic   string
	JWTSecret    string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
}

// Load читает конфиг из переменных окружения.
func Load() *Config {
	accessTTL, _ := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	refreshTTL, _ := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	return &Config{
		GRPCPort:     getEnvInt("AUTH_GRPC_PORT", 50051),
		HTTPPort:     getEnvInt("AUTH_HTTP_PORT", 8080),
		StaticDir:         getEnv("STATIC_DIR", ""),
		CustomersServiceURL: getEnv("CUSTOMERS_SERVICE_URL", ""),
		VehiclesServiceURL:  getEnv("VEHICLES_SERVICE_URL", ""),
		DealsServiceURL:    getEnv("DEALS_SERVICE_URL", ""),
		PartsServiceURL:    getEnv("PARTS_SERVICE_URL", ""),
		BrandsServiceURL:   getEnv("BRANDS_SERVICE_URL", ""),
		DealerPointsServiceURL: getEnv("DEALER_POINTS_SERVICE_URL", ""),
		PostgresDSN:  getEnv("POSTGRES_DSN", ""),
		RedisAddr:    getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:      getEnvInt("REDIS_DB", 0),
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "127.0.0.1:9092")},
		KafkaTopic:   getEnv("KAFKA_TOPIC_AUTH_EVENTS", "auth.events"),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
		AccessTTL:    accessTTL,
		RefreshTTL:   refreshTTL,
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
