package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/dealer/dealer/scheduler-service/internal/service"
)

type Worker struct {
	svc      *service.SchedulerService
	interval time.Duration
	logger   *slog.Logger
}

func NewWorker(svc *service.SchedulerService, interval time.Duration, logger *slog.Logger) *Worker {
	return &Worker{svc: svc, interval: interval, logger: logger}
}

func (w *Worker) Run(ctx context.Context) {
	if w == nil || w.svc == nil {
		return
	}
	w.tick(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	result, err := w.svc.RunOnce(ctx)
	if err != nil {
		w.logger.Error("review invitation scheduler failed", "err", err)
		return
	}
	if result.WorkOrdersCreated > 0 || result.DealsCreated > 0 || result.GoodsSalesCreated > 0 ||
		result.OrderReceiptsNotified > 0 || result.AppointmentRemindersSent > 0 {
		w.logger.Info("scheduler tick",
			"review_work_orders", result.WorkOrdersCreated,
			"review_deals", result.DealsCreated,
			"review_goods_sales", result.GoodsSalesCreated,
			"order_receipt_notifications", result.OrderReceiptsNotified,
			"appointment_reminders", result.AppointmentRemindersSent,
		)
	}
}
