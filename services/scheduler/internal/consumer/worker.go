package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/dealer/dealer/pkg/appointmentevent"
	"github.com/dealer/dealer/pkg/kafka"
)

// AppointmentCreatedRepository создаёт уведомления о записях на ремонт.
type AppointmentCreatedRepository interface {
	CreateRepairAppointmentBooked(ctx context.Context, appointmentID uuid.UUID) (int64, error)
}

// Worker обрабатывает события создания записей на ремонт из Kafka:
// при получении события создаёт клиенту уведомление «вы записаны на ремонт».
type Worker struct {
	consumer *kafka.Consumer
	repo     AppointmentCreatedRepository
	logger   *slog.Logger
}

func NewWorker(consumer *kafka.Consumer, repo AppointmentCreatedRepository, logger *slog.Logger) *Worker {
	return &Worker{consumer: consumer, repo: repo, logger: logger}
}

func (w *Worker) Run(ctx context.Context) {
	if w == nil || w.consumer == nil || w.repo == nil {
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
		if err := w.handle(ctx, msg.Value); err != nil {
			w.logger.Error("appointment.created event failed", "err", err)
			continue
		}
		if err := w.consumer.CommitMessage(ctx, msg); err != nil {
			w.logger.Warn("kafka commit failed", "err", err)
		}
	}
}

func (w *Worker) handle(ctx context.Context, raw []byte) error {
	var ev appointmentevent.CreatedEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		w.logger.Warn("invalid appointment.created json", "err", err)
		return nil
	}
	if ev.Event != appointmentevent.Created {
		w.logger.Warn("skip unknown appointment event", "event", ev.Event)
		return nil
	}
	appointmentID, err := uuid.Parse(ev.AppointmentID)
	if err != nil {
		w.logger.Warn("skip appointment.created: invalid appointment_id", "appointment_id", ev.AppointmentID)
		return nil
	}
	if _, err := w.repo.CreateRepairAppointmentBooked(ctx, appointmentID); err != nil {
		return err
	}
	return nil
}
