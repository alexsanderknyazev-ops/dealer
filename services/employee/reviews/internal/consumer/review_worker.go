package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/employee-reviews-service/internal/service"
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/reviewevent"
)

type ReviewWorker struct {
	consumer *kafka.Consumer
	svc      *service.ReviewService
	logger   *slog.Logger
}

func NewReviewWorker(consumer *kafka.Consumer, svc *service.ReviewService, logger *slog.Logger) *ReviewWorker {
	return &ReviewWorker{consumer: consumer, svc: svc, logger: logger}
}

func (w *ReviewWorker) Run(ctx context.Context) {
	if w == nil || w.consumer == nil || w.svc == nil {
		return
	}
	for {
		msg, err := w.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Warn("kafka fetch failed", "topic", "review.published", "err", err)
			continue
		}
		if err := w.handle(ctx, msg.Value); err != nil {
			w.logger.Error("review.published event failed", "err", err)
			continue
		}
		if err := w.consumer.CommitMessage(ctx, msg); err != nil {
			w.logger.Warn("kafka commit failed", "err", err)
		}
	}
}

func (w *ReviewWorker) handle(ctx context.Context, raw []byte) error {
	var ev reviewevent.PublishedEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		w.logger.Warn("invalid review.published json", "err", err)
		return nil
	}
	if ev.Event != reviewevent.Published {
		w.logger.Warn("skip unknown review event", "event", ev.Event)
		return nil
	}
	reviewID, err := uuid.Parse(ev.ReviewID)
	if err != nil {
		w.logger.Warn("skip review.published: invalid review_id")
		return nil
	}
	clientID, err := uuid.Parse(ev.ClientID)
	if err != nil {
		w.logger.Warn("skip review.published: invalid client_id")
		return nil
	}
	userID, err := uuid.Parse(ev.UserID)
	if err != nil {
		w.logger.Warn("skip review.published: invalid user_id")
		return nil
	}
	dealerPointID, err := uuid.Parse(ev.DealerPointID)
	if err != nil {
		w.logger.Warn("skip review.published: invalid dealer_point_id")
		return nil
	}
	vehicleID, err := uuid.Parse(ev.VehicleID)
	if err != nil {
		w.logger.Warn("skip review.published: invalid vehicle_id")
		return nil
	}
	occurredAt := time.Unix(ev.OccurredAt, 0).UTC()
	if ev.OccurredAt == 0 {
		occurredAt = time.Now().UTC()
	}
	err = w.svc.RecordReviewPublished(ctx, reviewID, clientID, userID, dealerPointID, vehicleID,
		ev.ClientEmail, ev.ClientFullName, ev.Text, ev.Status, ev.Rating, occurredAt)
	if err != nil {
		if errors.Is(err, service.ErrVehicleNotFound) {
			w.logger.Warn("skip review.published: vehicle not found", "vehicle_id", ev.VehicleID)
			return nil
		}
		return err
	}
	return nil
}
