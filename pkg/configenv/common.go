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

// KafkaClientRegistration — события регистрации клиентов.
type KafkaClientRegistration struct {
	Brokers []string
	Topic   string
}

// LoadKafkaClientRegistration загружает KAFKA_BROKERS и KAFKA_TOPIC_CLIENT_REGISTRATION.
func LoadKafkaClientRegistration(brokersDefault string) KafkaClientRegistration {
	return KafkaClientRegistration{
		Brokers: Brokers("KAFKA_BROKERS", brokersDefault),
		Topic:   String("KAFKA_TOPIC_CLIENT_REGISTRATION", "client.registration.v1"),
	}
}

// KafkaDealCompleted — события успешных сделок.
type KafkaDealCompleted struct {
	Brokers []string
	Topic   string
}

func LoadKafkaDealCompleted(brokersDefault string) KafkaDealCompleted {
	return KafkaDealCompleted{
		Brokers: Brokers("KAFKA_BROKERS", brokersDefault),
		Topic:   String("KAFKA_TOPIC_DEAL_COMPLETED", "deal.completed.v1"),
	}
}

// KafkaReviewPublished — события опубликованных отзывов.
type KafkaReviewPublished struct {
	Brokers []string
	Topic   string
}

func LoadKafkaReviewPublished(brokersDefault string) KafkaReviewPublished {
	return KafkaReviewPublished{
		Brokers: Brokers("KAFKA_BROKERS", brokersDefault),
		Topic:   String("KAFKA_TOPIC_REVIEW_PUBLISHED", "review.published.v1"),
	}
}

// KafkaAppointmentCreated — события создания записей на ремонт.
type KafkaAppointmentCreated struct {
	Brokers []string
	Topic   string
}

// LoadKafkaAppointmentCreated загружает KAFKA_BROKERS и KAFKA_TOPIC_APPOINTMENT_CREATED.
func LoadKafkaAppointmentCreated(brokersDefault string) KafkaAppointmentCreated {
	return KafkaAppointmentCreated{
		Brokers: Brokers("KAFKA_BROKERS", brokersDefault),
		Topic:   String("KAFKA_TOPIC_APPOINTMENT_CREATED", "repair.appointment.created.v1"),
	}
}
