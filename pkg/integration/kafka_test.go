//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"

	tc "github.com/dealer/dealer/pkg/testcontainers"
	pkgkafka "github.com/dealer/dealer/pkg/kafka"
)

func TestKafka_ProduceConsumeRoundtrip(t *testing.T) {
	ctx := context.Background()
	container, err := tc.StartKafka(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = container.Close(ctx) })

	brokers := container.Brokers
	if len(brokers) == 0 {
		t.Fatal("kafka brokers list is empty")
	}

	topic := fmt.Sprintf("integration-topic-%d", time.Now().UnixNano())
	group := fmt.Sprintf("integration-group-%d", time.Now().UnixNano())

	client := &kafka.Client{Addr: kafka.TCP(brokers[0])}
	if _, err := client.CreateTopics(ctx, &kafka.CreateTopicsRequest{
		Topics: []kafka.TopicConfig{
			{Topic: topic, NumPartitions: 1, ReplicationFactor: 1},
		},
	}); err != nil {
		t.Fatalf("create topic: %v", err)
	}

	producer := pkgkafka.NewProducer(brokers, topic)
	defer producer.Close()

	const messages = 3
	for i := 0; i < messages; i++ {
		key := fmt.Sprintf("k-%d", i)
		value := fmt.Sprintf("value-%d", i)
		if err := producer.Publish(ctx, []byte(key), []byte(value)); err != nil {
			t.Fatalf("publish %d: %v", i, err)
		}
	}

	consumer := pkgkafka.NewConsumer(brokers, topic, group)
	defer consumer.Close()

	deadline := time.Now().Add(90 * time.Second)
	got := make(map[string]string, messages)
	for len(got) < messages && time.Now().Before(deadline) {
		fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		msg, err := consumer.FetchMessage(fetchCtx)
		cancel()
		if err != nil {
			continue
		}
		got[string(msg.Key)] = string(msg.Value)
		_ = consumer.CommitMessage(ctx, msg)
	}

	if len(got) != messages {
		t.Fatalf("received %d/%d messages: %v", len(got), messages, got)
	}
	for i := 0; i < messages; i++ {
		key := fmt.Sprintf("k-%d", i)
		want := fmt.Sprintf("value-%d", i)
		if got[key] != want {
			t.Errorf("message %q: got %q want %q", key, got[key], want)
		}
	}
}
