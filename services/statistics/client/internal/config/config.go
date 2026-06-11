package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort                 int
	HTTPPort                 int
	PostgresDSN              string
	JWTSecret                string
	KafkaBrokers             []string
	KafkaReviewTopic         string
	KafkaRegistrationTopic   string
	KafkaConsumerGroup       string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("CLIENT_STATISTICS_GRPC_PORT", 50062, "CLIENT_STATISTICS_HTTP_PORT", 8095)
	pj := configenv.LoadPostgresJWT()
	kafkaReview := configenv.LoadKafkaReviewPublished("127.0.0.1:9092")
	kafkaReg := configenv.LoadKafkaClientRegistration("127.0.0.1:9092")
	return &Config{
		GRPCPort:               ports.GRPCPort,
		HTTPPort:               ports.HTTPPort,
		PostgresDSN:            pj.PostgresDSN,
		JWTSecret:              pj.JWTSecret,
		KafkaBrokers:           kafkaReview.Brokers,
		KafkaReviewTopic:       kafkaReview.Topic,
		KafkaRegistrationTopic: kafkaReg.Topic,
		KafkaConsumerGroup:     configenv.String("KAFKA_CONSUMER_GROUP_CLIENT_STATISTICS", "client-statistics"),
	}
}
