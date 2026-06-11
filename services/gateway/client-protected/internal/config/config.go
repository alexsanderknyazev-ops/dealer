package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	HTTPPort                   int
	ClientAuthGRPCAddr         string
	ClientRegistrationGRPCAddr string
	ClientReviewsGRPCAddr      string
}

func Load() *Config {
	return &Config{
		HTTPPort:                   configenv.Int("CLIENT_PROTECTED_GATEWAY_HTTP_PORT", 8093),
		ClientAuthGRPCAddr:         configenv.String("CLIENT_AUTH_GRPC_ADDR", "127.0.0.1:50059"),
		ClientRegistrationGRPCAddr: configenv.String("CLIENT_REGISTRATION_GRPC_ADDR", "127.0.0.1:50058"),
		ClientReviewsGRPCAddr:      configenv.String("CLIENT_REVIEWS_GRPC_ADDR", "127.0.0.1:50060"),
	}
}
