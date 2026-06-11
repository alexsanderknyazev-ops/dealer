package config

import (
	"time"

	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort      int
	HTTPPort      int
	PostgresDSN   string
	RedisAddr     string
	RedisPass     string
	RedisDB       int
	JWTSecret     string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	RefreshPrefix      string
	KafkaBrokers       []string
	KafkaTopic         string
	KafkaConsumerGroup string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("CLIENT_AUTH_GRPC_PORT", 50059, "CLIENT_AUTH_HTTP_PORT", 8088)
	pj := configenv.LoadPostgresJWT()
	return &Config{
		GRPCPort:      ports.GRPCPort,
		HTTPPort:      ports.HTTPPort,
		PostgresDSN:   pj.PostgresDSN,
		RedisAddr:     configenv.String("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass:     configenv.String("REDIS_PASSWORD", ""),
		RedisDB:       configenv.Int("REDIS_DB", 0),
		JWTSecret:     pj.JWTSecret,
		AccessTTL:     configenv.Duration("JWT_ACCESS_TTL", 15*time.Minute),
		RefreshTTL:    configenv.Duration("JWT_REFRESH_TTL", 168*time.Hour),
		RefreshPrefix:      configenv.String("CLIENT_AUTH_REFRESH_PREFIX", "client-auth:refresh:"),
		KafkaBrokers:       configenv.Brokers("KAFKA_BROKERS", "127.0.0.1:9092"),
		KafkaTopic:         configenv.String("KAFKA_TOPIC_CLIENT_REGISTRATION", "client.registration.v1"),
		KafkaConsumerGroup: configenv.String("KAFKA_CONSUMER_GROUP_CLIENT_AUTH", "client-auth"),
	}
}
