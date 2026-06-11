package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	HTTPPort           int
	KafkaBrokers       []string
	KafkaTopicErrors   string
	KafkaConsumerGroup string
	ClickHouseAddr     string
	ClickHouseDatabase string
	ClickHouseUser     string
	ClickHousePassword string
	Environment        string
}

func Load() *Config {
	kafka := configenv.LoadKafkaErrors("127.0.0.1:9092")
	return &Config{
		HTTPPort:           configenv.Int("ERRORS_INGEST_HTTP_PORT", 8092),
		KafkaBrokers:       kafka.Brokers,
		KafkaTopicErrors:   kafka.Topic,
		KafkaConsumerGroup: configenv.String("KAFKA_CONSUMER_GROUP", "errors-ingest"),
		ClickHouseAddr:     configenv.String("CLICKHOUSE_ADDR", "127.0.0.1:9000"),
		ClickHouseDatabase: configenv.String("CLICKHOUSE_DATABASE", "analytics"),
		ClickHouseUser:     configenv.String("CLICKHOUSE_USER", "default"),
		ClickHousePassword: configenv.String("CLICKHOUSE_PASSWORD", ""),
		Environment:        configenv.String("ENVIRONMENT", "development"),
	}
}
