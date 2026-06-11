package ingest

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/dealer/dealer/pkg/errorevent"
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/services/errors-ingest/internal/clickhouse"
)

// Worker читает Kafka и пишет события в ClickHouse.
type Worker struct {
	consumer *kafka.Consumer
	store    *clickhouse.Store
	logger   *slog.Logger
}

// NewWorker создаёт consumer worker.
func NewWorker(consumer *kafka.Consumer, store *clickhouse.Store, logger *slog.Logger) *Worker {
	return &Worker{consumer: consumer, store: store, logger: logger}
}

// Run блокируется до отмены ctx.
func (w *Worker) Run(ctx context.Context) {
	if w == nil || w.consumer == nil || w.store == nil {
		return
	}
	for {
		msg, err := w.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Warn("kafka fetch failed", "err", err)
			continue
		}
		var ev errorevent.Event
		if err := json.Unmarshal(msg.Value, &ev); err != nil {
			w.logger.Warn("invalid error event json", "err", err)
			_ = w.consumer.CommitMessage(ctx, msg)
			continue
		}
		if ev.EventID == "" || ev.Service == "" {
			w.logger.Warn("skip error event: missing event_id or service")
			_ = w.consumer.CommitMessage(ctx, msg)
			continue
		}
		if err := w.store.InsertEvents(ctx, []errorevent.Event{ev}); err != nil {
			w.logger.Error("clickhouse insert failed", "err", err, "event_id", ev.EventID)
			continue
		}
		if err := w.consumer.CommitMessage(ctx, msg); err != nil {
			w.logger.Warn("kafka commit failed", "err", err)
		}
	}
}
