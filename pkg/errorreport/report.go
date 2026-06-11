package errorreport

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/dealer/dealer/pkg/configenv"
	"github.com/dealer/dealer/pkg/errorevent"
	"github.com/dealer/dealer/pkg/kafka"
)

// Publisher публикует событие (Kafka и т.п.).
type Publisher interface {
	Publish(ctx context.Context, ev errorevent.Event) error
}

// KafkaPublisher отправляет JSON в Kafka topic.
type KafkaPublisher struct {
	producer *kafka.Producer
}

// NewKafkaPublisher создаёт Kafka publisher.
func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	if len(brokers) == 0 || topic == "" {
		return nil
	}
	return &KafkaPublisher{producer: kafka.NewProducer(brokers, topic)}
}

// Publish сериализует событие и пишет в Kafka.
func (p *KafkaPublisher) Publish(ctx context.Context, ev errorevent.Event) error {
	if p == nil || p.producer == nil {
		return nil
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	key := ev.EventID
	if ev.TraceID != "" {
		key = ev.TraceID
	}
	return p.producer.Publish(ctx, []byte(key), body)
}

// Close закрывает producer.
func (p *KafkaPublisher) Close() error {
	if p == nil || p.producer == nil {
		return nil
	}
	return p.producer.Close()
}

// Reporter — best-effort отправка ошибок (не блокирует hot path).
type Reporter struct {
	pub         Publisher
	service     string
	environment string
	logger      *slog.Logger
}

// New создаёт reporter.
func New(pub Publisher, service, environment string, logger *slog.Logger) *Reporter {
	if pub == nil {
		return nil
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Reporter{pub: pub, service: service, environment: environment, logger: logger}
}

// NewFromEnv читает KAFKA_BROKERS и KAFKA_TOPIC_ERRORS; при отсутствии возвращает nil.
func NewFromEnv(service string, logger *slog.Logger) *Reporter {
	kafkaCfg := configenv.LoadKafkaErrors("")
	if len(kafkaCfg.Brokers) == 0 || kafkaCfg.Topic == "" {
		return nil
	}
	env := configenv.String("ENVIRONMENT", "development")
	return New(NewKafkaPublisher(kafkaCfg.Brokers, kafkaCfg.Topic), service, env, logger)
}

// Report публикует событие асинхронно (best-effort).
func (r *Reporter) Report(ev errorevent.Event) {
	if r == nil {
		return
	}
	if ev.Service == "" {
		ev.Service = r.service
	}
	if ev.Source == "" {
		ev.Source = r.service
	}
	if ev.Environment == "" {
		ev.Environment = r.environment
	}
	if ev.EventID == "" {
		copy := errorevent.New(r.service, ev.Kind, ev.Severity, ev.Message)
		ev.EventID = copy.EventID
		ev.OccurredAt = copy.OccurredAt
		ev.SchemaVersion = copy.SchemaVersion
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := r.pub.Publish(ctx, ev); err != nil {
			r.logger.Warn("error report publish failed", "err", err, "kind", ev.Kind)
		}
	}()
}
