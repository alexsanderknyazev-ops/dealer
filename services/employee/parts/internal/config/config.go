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
	WorkOrdersGRPCAddr   string
	EmployeesGRPCAddr    string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("PARTS_GRPC_PORT", 50055, "PARTS_HTTP_PORT", 8084)
	pj := configenv.LoadPostgresJWT()
	return &Config{
		GRPCPort:             ports.GRPCPort,
		HTTPPort:             ports.HTTPPort,
		PostgresDSN:          pj.PostgresDSN,
		JWTSecret:            pj.JWTSecret,
		BrandsGRPCAddr:       configenv.String("BRANDS_GRPC_ADDR", ""),
		DealerPointsGRPCAddr: configenv.String("DEALER_POINTS_GRPC_ADDR", ""),
		WorkOrdersGRPCAddr:   configenv.String("WORKORDERS_GRPC_ADDR", ""),
		EmployeesGRPCAddr:    configenv.String("EMPLOYEES_GRPC_ADDR", ""),
	}
}
