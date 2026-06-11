package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/dealer/dealer/client-auth-service/internal/service"
	"github.com/dealer/dealer/pkg/clientevent"
	"github.com/dealer/dealer/pkg/kafka"
)

// Worker обрабатывает события регистрации клиентов из Kafka.
type Worker struct {
	consumer *kafka.Consumer
	auth     *service.AuthService
	logger   *slog.Logger
}

func NewWorker(consumer *kafka.Consumer, auth *service.AuthService, logger *slog.Logger) *Worker {
	return &Worker{consumer: consumer, auth: auth, logger: logger}
}

func (w *Worker) Run(ctx context.Context) {
	if w == nil || w.consumer == nil || w.auth == nil {
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
			w.logger.Error("client registration event failed", "err", err)
			continue
		}
		if err := w.consumer.CommitMessage(ctx, msg); err != nil {
			w.logger.Warn("kafka commit failed", "err", err)
		}
	}
}

func (w *Worker) handle(ctx context.Context, raw []byte) error {
	var ev clientevent.RegisteredEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		w.logger.Warn("invalid client registration json", "err", err)
		return nil
	}
	if ev.Event != clientevent.Registered {
		w.logger.Warn("skip unknown client event", "event", ev.Event)
		return nil
	}
	if ev.UserID == "" || ev.Email == "" || ev.PasswordHash == "" {
		w.logger.Warn("skip client registration: missing required fields")
		return nil
	}
	userID, err := uuid.Parse(ev.UserID)
	if err != nil {
		w.logger.Warn("skip client registration: invalid user_id", "user_id", ev.UserID)
		return nil
	}
	return w.auth.RegisterFromEvent(ctx, userID, ev.Email, ev.PasswordHash, ev.FullName, ev.Phone)
}
