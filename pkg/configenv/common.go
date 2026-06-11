package configenv

const jwtSecretDefault = "change-me-in-production"

// PostgresJWT — общие поля domain-сервисов с PostgreSQL и JWT.
type PostgresJWT struct {
	PostgresDSN string
	JWTSecret   string
}

// LoadPostgresJWT загружает POSTGRES_DSN и JWT_SECRET.
func LoadPostgresJWT() PostgresJWT {
	return PostgresJWT{
		PostgresDSN: String("POSTGRES_DSN", ""),
		JWTSecret:   String("JWT_SECRET", jwtSecretDefault),
	}
}

// ServicePorts — gRPC + health HTTP порты.
type ServicePorts struct {
	GRPCPort int
	HTTPPort int
}

// LoadServicePorts загружает пары GRPC/HTTP портов по именам env-переменных.
func LoadServicePorts(grpcKey string, grpcDef int, httpKey string, httpDef int) ServicePorts {
	return ServicePorts{
		GRPCPort: Int(grpcKey, grpcDef),
		HTTPPort: Int(httpKey, httpDef),
	}
}

// KafkaErrors — публикация ошибок в Kafka (опционально).
type KafkaErrors struct {
	Brokers []string
	Topic   string
}

// LoadKafkaErrors загружает KAFKA_BROKERS и KAFKA_TOPIC_ERRORS.
func LoadKafkaErrors(brokersDefault string) KafkaErrors {
	return KafkaErrors{
		Brokers: Brokers("KAFKA_BROKERS", brokersDefault),
		Topic:   String("KAFKA_TOPIC_ERRORS", "platform.errors.v1"),
	}
}
