package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort         int
	HTTPPort         int
	PostgresDSN      string
	JWTSecret        string
	VehiclesGRPCAddr string
	KafkaBrokers     []string
	KafkaTopic       string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("CLIENT_REVIEWS_GRPC_PORT", 50060, "CLIENT_REVIEWS_HTTP_PORT", 8089)
	pj := configenv.LoadPostgresJWT()
	return &Config{
		GRPCPort:         ports.GRPCPort,
		HTTPPort:         ports.HTTPPort,
		PostgresDSN:      pj.PostgresDSN,
		JWTSecret:        pj.JWTSecret,
		VehiclesGRPCAddr: configenv.String("VEHICLES_GRPC_ADDR", "127.0.0.1:50053"),
		KafkaBrokers:     configenv.LoadKafkaReviewPublished("127.0.0.1:9092").Brokers,
		KafkaTopic:       configenv.LoadKafkaReviewPublished("127.0.0.1:9092").Topic,
	}
}
