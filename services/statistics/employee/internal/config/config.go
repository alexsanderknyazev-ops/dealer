package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort           int
	HTTPPort           int
	PostgresDSN        string
	JWTSecret          string
	KafkaBrokers       []string
	KafkaDealTopic     string
	KafkaConsumerGroup string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("EMPLOYEE_STATISTICS_GRPC_PORT", 50061, "EMPLOYEE_STATISTICS_HTTP_PORT", 8094)
	pj := configenv.LoadPostgresJWT()
	kafkaDeal := configenv.LoadKafkaDealCompleted("127.0.0.1:9092")
	return &Config{
		GRPCPort:           ports.GRPCPort,
		HTTPPort:           ports.HTTPPort,
		PostgresDSN:        pj.PostgresDSN,
		JWTSecret:          pj.JWTSecret,
		KafkaBrokers:       kafkaDeal.Brokers,
		KafkaDealTopic:     kafkaDeal.Topic,
		KafkaConsumerGroup: configenv.String("KAFKA_CONSUMER_GROUP_EMPLOYEE_STATISTICS", "employee-statistics"),
	}
}
