package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/client-statistics-service/internal/service"
	"github.com/dealer/dealer/pkg/clientevent"
	"github.com/dealer/dealer/pkg/kafka"
)

type RegistrationWorker struct {
	consumer *kafka.Consumer
	stats    *service.StatsService
	logger   *slog.Logger
}

func NewRegistrationWorker(consumer *kafka.Consumer, stats *service.StatsService, logger *slog.Logger) *RegistrationWorker {
	return &RegistrationWorker{consumer: consumer, stats: stats, logger: logger}
}

func (w *RegistrationWorker) Run(ctx context.Context) {
	if w == nil || w.consumer == nil || w.stats == nil {
		return
	}
	for {
		msg, err := w.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Warn("kafka fetch failed", "topic", "client.registration", "err", err)
			continue
		}
		if err := w.handle(ctx, msg.Value); err != nil {
			w.logger.Error("client.registration stats event failed", "err", err)
			continue
		}
		if err := w.consumer.CommitMessage(ctx, msg); err != nil {
			w.logger.Warn("kafka commit failed", "err", err)
		}
	}
}

func (w *RegistrationWorker) handle(ctx context.Context, raw []byte) error {
	var ev clientevent.RegisteredEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		w.logger.Warn("invalid client.registration json", "err", err)
		return nil
	}
	if ev.Event != clientevent.Registered {
		w.logger.Warn("skip unknown client event", "event", ev.Event)
		return nil
	}
	userID, err := uuid.Parse(ev.UserID)
	if err != nil {
		return nil
	}
	var vehicleID *uuid.UUID
	if ev.VehicleID != "" {
		if vid, err := uuid.Parse(ev.VehicleID); err == nil {
			vehicleID = &vid
		}
	}
	occurredAt := time.Now().UTC()
	return w.stats.RecordClientRegistration(ctx, userID, ev.Email, vehicleID, occurredAt)
}
