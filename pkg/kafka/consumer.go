package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Consumer читает сообщения из Kafka topic.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer создаёт consumer group reader.
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	if len(brokers) == 0 || topic == "" || groupID == "" {
		return nil
	}
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

// FetchMessage читает следующее сообщение.
func (c *Consumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

// CommitMessage фиксирует offset.
func (c *Consumer) CommitMessage(ctx context.Context, msg kafka.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

// Close закрывает reader.
func (c *Consumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
