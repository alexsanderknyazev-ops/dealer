package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort            int
	HTTPPort            int
	PostgresDSN         string
	JWTSecret           string
	CustomersGRPCAddr   string
	VehiclesGRPCAddr    string
	DealerPointsGRPCAddr string
	PartsGRPCAddr       string
	WorksGRPCAddr       string
	WorkOrdersGRPCAddr  string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("APPOINTMENTS_GRPC_PORT", 50067, "APPOINTMENTS_HTTP_PORT", 8101)
	pj := configenv.LoadPostgresJWT()
	return &Config{
		GRPCPort:             ports.GRPCPort,
		HTTPPort:             ports.HTTPPort,
		PostgresDSN:          pj.PostgresDSN,
		JWTSecret:            pj.JWTSecret,
		CustomersGRPCAddr:    configenv.String("CUSTOMERS_GRPC_ADDR", ""),
		VehiclesGRPCAddr:     configenv.String("VEHICLES_GRPC_ADDR", ""),
		DealerPointsGRPCAddr: configenv.String("DEALER_POINTS_GRPC_ADDR", ""),
		PartsGRPCAddr:        configenv.String("PARTS_GRPC_ADDR", ""),
		WorksGRPCAddr:        configenv.String("WORKS_GRPC_ADDR", ""),
		WorkOrdersGRPCAddr:   configenv.String("WORKORDERS_GRPC_ADDR", ""),
	}
}
