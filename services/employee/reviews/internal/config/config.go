package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	GRPCPort           int
	HTTPPort           int
	PostgresDSN        string
	JWTSecret          string
	VehiclesGRPCAddr   string
	KafkaBrokers       []string
	KafkaReviewTopic   string
	KafkaConsumerGroup string
}

func Load() *Config {
	ports := configenv.LoadServicePorts("EMPLOYEE_REVIEWS_GRPC_PORT", 50063, "EMPLOYEE_REVIEWS_HTTP_PORT", 8096)
	pj := configenv.LoadPostgresJWT()
	kafkaReview := configenv.LoadKafkaReviewPublished("127.0.0.1:9092")
	return &Config{
		GRPCPort:           ports.GRPCPort,
		HTTPPort:           ports.HTTPPort,
		PostgresDSN:        pj.PostgresDSN,
		JWTSecret:          pj.JWTSecret,
		VehiclesGRPCAddr:   configenv.String("VEHICLES_GRPC_ADDR", "127.0.0.1:50053"),
		KafkaBrokers:       kafkaReview.Brokers,
		KafkaReviewTopic:   kafkaReview.Topic,
		KafkaConsumerGroup: configenv.String("KAFKA_CONSUMER_GROUP_EMPLOYEE_REVIEWS", "employee-reviews"),
	}
}
