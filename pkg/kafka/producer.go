package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Producer обёртка над kafka.Writer для отправки сообщений.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer создаёт продюсер для топика.
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers[0]),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// Publish отправляет сообщение с ключом и телом.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{Key: key, Value: value})
}

// Close закрывает writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}
