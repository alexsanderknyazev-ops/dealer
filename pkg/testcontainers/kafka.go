package testcontainers

import (
	"context"
	"fmt"
	"sync"

	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

// Kafka — запущенный контейнер Kafka.
type Kafka struct {
	Container *kafka.KafkaContainer
	Brokers   []string

	closeOnce sync.Once
}

// StartKafka запускает контейнер Kafka и возвращает список внешних брокеров.
func StartKafka(ctx context.Context) (*Kafka, error) {
	container, err := kafka.Run(ctx, kafkaImage, kafka.WithClusterID("integration"))
	if err != nil {
		return nil, fmt.Errorf("start kafka: %w", err)
	}
	brokers, err := container.Brokers(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("kafka brokers: %w", err)
	}
	return &Kafka{Container: container, Brokers: brokers}, nil
}

// Close освобождает ресурсы. Безопасен для многократного вызова.
func (k *Kafka) Close(ctx context.Context) error {
	if k == nil {
		return nil
	}
	k.closeOnce.Do(func() {
		_ = k.Container.Terminate(ctx)
	})
	return nil
}
