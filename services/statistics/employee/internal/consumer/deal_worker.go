package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/employee-statistics-service/internal/service"
	"github.com/dealer/dealer/pkg/dealevent"
	"github.com/dealer/dealer/pkg/kafka"
)

type DealWorker struct {
	consumer *kafka.Consumer
	stats    *service.StatsService
	logger   *slog.Logger
}

func NewDealWorker(consumer *kafka.Consumer, stats *service.StatsService, logger *slog.Logger) *DealWorker {
	return &DealWorker{consumer: consumer, stats: stats, logger: logger}
}

func (w *DealWorker) Run(ctx context.Context) {
	if w == nil || w.consumer == nil || w.stats == nil {
		return
	}
	for {
		msg, err := w.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Warn("kafka fetch failed", "topic", "deal.completed", "err", err)
			continue
		}
		if err := w.handle(ctx, msg.Value); err != nil {
			w.logger.Error("deal.completed event failed", "err", err)
			continue
		}
		if err := w.consumer.CommitMessage(ctx, msg); err != nil {
			w.logger.Warn("kafka commit failed", "err", err)
		}
	}
}

func (w *DealWorker) handle(ctx context.Context, raw []byte) error {
	var ev dealevent.CompletedEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		w.logger.Warn("invalid deal.completed json", "err", err)
		return nil
	}
	if ev.Event != dealevent.Completed {
		w.logger.Warn("skip unknown deal event", "event", ev.Event)
		return nil
	}
	dealID, err := uuid.Parse(ev.DealID)
	if err != nil {
		w.logger.Warn("skip deal.completed: invalid deal_id")
		return nil
	}
	customerID, err := uuid.Parse(ev.CustomerID)
	if err != nil {
		w.logger.Warn("skip deal.completed: invalid customer_id")
		return nil
	}
	vehicleID, err := uuid.Parse(ev.VehicleID)
	if err != nil {
		w.logger.Warn("skip deal.completed: invalid vehicle_id")
		return nil
	}
	if ev.Stage != "completed" && ev.Stage != "paid" {
		w.logger.Warn("skip deal.completed: unexpected stage", "stage", ev.Stage)
		return nil
	}
	occurredAt := time.Unix(ev.OccurredAt, 0).UTC()
	if ev.OccurredAt == 0 {
		occurredAt = time.Now().UTC()
	}
	return w.stats.RecordDealCompleted(ctx, dealID, customerID, vehicleID, ev.Amount, ev.Stage, occurredAt)
}
