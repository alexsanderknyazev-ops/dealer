package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort          int
	HTTPPort          int
	PostgresDSN       string
	JWTSecret         string
	ClientAuthGRPCAddr string
	VehiclesGRPCAddr   string
	KafkaBrokers       []string
	KafkaTopic         string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("CLIENT_REGISTRATION_GRPC_PORT", 50058, "CLIENT_REGISTRATION_HTTP_PORT", 8087)
	pj := configenv.LoadPostgresJWT()
	kafkaClientReg := configenv.LoadKafkaClientRegistration("127.0.0.1:9092")
	return &Config{
		GRPCPort:         ports.GRPCPort,
		HTTPPort:         ports.HTTPPort,
		PostgresDSN:      pj.PostgresDSN,
		JWTSecret:        pj.JWTSecret,
		ClientAuthGRPCAddr: configenv.String("CLIENT_AUTH_GRPC_ADDR", "127.0.0.1:50059"),
		VehiclesGRPCAddr:   configenv.String("VEHICLES_GRPC_ADDR", "127.0.0.1:50053"),
		KafkaBrokers:       kafkaClientReg.Brokers,
		KafkaTopic:         kafkaClientReg.Topic,
	}
}
