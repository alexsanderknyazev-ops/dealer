package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort    int
	HTTPPort    int
	PostgresDSN string
	JWTSecret   string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("BRANDS_GRPC_PORT", 50056, "BRANDS_HTTP_PORT", 8085)
	pj := configenv.LoadPostgresJWT()
	return &Config{
		GRPCPort:    ports.GRPCPort,
		HTTPPort:    ports.HTTPPort,
		PostgresDSN: pj.PostgresDSN,
		JWTSecret:   pj.JWTSecret,
	}
}
