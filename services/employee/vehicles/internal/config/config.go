package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort             int
	HTTPPort             int
	PostgresDSN          string
	JWTSecret            string
	BrandsGRPCAddr       string
	DealerPointsGRPCAddr string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("VEHICLES_GRPC_PORT", 50053, "VEHICLES_HTTP_PORT", 8082)
	pj := configenv.LoadPostgresJWT()
	return &Config{
		GRPCPort:             ports.GRPCPort,
		HTTPPort:             ports.HTTPPort,
		PostgresDSN:          pj.PostgresDSN,
		JWTSecret:            pj.JWTSecret,
		BrandsGRPCAddr:       configenv.String("BRANDS_GRPC_ADDR", ""),
		DealerPointsGRPCAddr: configenv.String("DEALER_POINTS_GRPC_ADDR", ""),
	}
}
