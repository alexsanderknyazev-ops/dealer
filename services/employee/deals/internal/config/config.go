package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort          int
	HTTPPort          int
	PostgresDSN       string
	JWTSecret         string
	CustomersGRPCAddr string
	VehiclesGRPCAddr  string
	KafkaBrokers      []string
	KafkaTopic        string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("DEALS_GRPC_PORT", 50054, "DEALS_HTTP_PORT", 8083)
	pj := configenv.LoadPostgresJWT()
	return &Config{
		GRPCPort:          ports.GRPCPort,
		HTTPPort:          ports.HTTPPort,
		PostgresDSN:       pj.PostgresDSN,
		JWTSecret:         pj.JWTSecret,
		CustomersGRPCAddr: configenv.String("CUSTOMERS_GRPC_ADDR", ""),
		VehiclesGRPCAddr:  configenv.String("VEHICLES_GRPC_ADDR", ""),
		KafkaBrokers:      configenv.LoadKafkaDealCompleted("127.0.0.1:9092").Brokers,
		KafkaTopic:        configenv.LoadKafkaDealCompleted("127.0.0.1:9092").Topic,
	}
}
