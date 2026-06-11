package config

import (
	"time"

	"github.com/dealer/dealer/pkg/configenv"
)

// Config — конфигурация auth-микросервиса.
type Config struct {
	GRPCPort               int
	HTTPPort               int
	StaticDir              string
	GatewayServiceURL      string
	ErrorsIngestServiceURL string
	CustomersServiceURL    string
	VehiclesServiceURL     string
	DealsServiceURL        string
	PartsServiceURL        string
	BrandsServiceURL       string
	DealerPointsServiceURL string
	PostgresDSN            string
	RedisAddr              string
	RedisPass              string
	RedisDB                int
	KafkaBrokers           []string
	KafkaTopic             string
	JWTSecret              string
	AccessTTL              time.Duration
	RefreshTTL             time.Duration
}

// Load читает конфиг из переменных окружения.
func Load() *Config {
	ports := configenv.LoadServicePorts("AUTH_GRPC_PORT", 50051, "AUTH_HTTP_PORT", 8080)
	pj := configenv.LoadPostgresJWT()
	return &Config{
		GRPCPort:               ports.GRPCPort,
		HTTPPort:               ports.HTTPPort,
		StaticDir:              configenv.String("STATIC_DIR", ""),
		GatewayServiceURL:      configenv.String("GATEWAY_SERVICE_URL", ""),
		ErrorsIngestServiceURL: configenv.String("ERRORS_INGEST_SERVICE_URL", ""),
		CustomersServiceURL:    configenv.String("CUSTOMERS_SERVICE_URL", ""),
		VehiclesServiceURL:     configenv.String("VEHICLES_SERVICE_URL", ""),
		DealsServiceURL:        configenv.String("DEALS_SERVICE_URL", ""),
		PartsServiceURL:        configenv.String("PARTS_SERVICE_URL", ""),
		BrandsServiceURL:       configenv.String("BRANDS_SERVICE_URL", ""),
		DealerPointsServiceURL: configenv.String("DEALER_POINTS_SERVICE_URL", ""),
		PostgresDSN:            pj.PostgresDSN,
		RedisAddr:              configenv.String("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass:              configenv.String("REDIS_PASSWORD", ""),
		RedisDB:                configenv.Int("REDIS_DB", 0),
		KafkaBrokers:           configenv.Brokers("KAFKA_BROKERS", "127.0.0.1:9092"),
		KafkaTopic:             configenv.String("KAFKA_TOPIC_AUTH_EVENTS", "auth.events"),
		JWTSecret:              pj.JWTSecret,
		AccessTTL:              configenv.Duration("JWT_ACCESS_TTL", 15*time.Minute),
		RefreshTTL:             configenv.Duration("JWT_REFRESH_TTL", 168*time.Hour),
	}
}
