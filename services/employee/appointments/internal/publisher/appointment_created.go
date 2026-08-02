package publisher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dealer/dealer/pkg/appointmentevent"
	"github.com/dealer/dealer/services/appointments/internal/domain"
)

// Producer отправляет сообщения в Kafka.
type Producer interface {
	Publish(ctx context.Context, key, value []byte) error
}

// AppointmentCreated публикует событие создания записи на ремонт.
type AppointmentCreated struct {
	producer Producer
}

func NewAppointmentCreated(producer Producer) *AppointmentCreated {
	return &AppointmentCreated{producer: producer}
}

func (p *AppointmentCreated) Publish(ctx context.Context, a *domain.RepairAppointment) error {
	if p == nil || p.producer == nil || a == nil {
		return nil
	}
	ev := appointmentevent.CreatedEvent{
		Event:             appointmentevent.Created,
		AppointmentID:     a.ID.String(),
		AppointmentNumber: a.AppointmentNumber,
		CustomerID:        a.CustomerID.String(),
		VehicleID:         a.VehicleID.String(),
		ScheduledStart:    a.ScheduledStart.Unix(),
		ScheduledEnd:      a.ScheduledEnd.Unix(),
		OccurredAt:        a.CreatedAt.Unix(),
	}
	if a.DealerPointID != nil {
		ev.DealerPointID = a.DealerPointID.String()
	}
	if ev.OccurredAt == 0 {
		ev.OccurredAt = time.Now().UTC().Unix()
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return p.producer.Publish(ctx, []byte(a.ID.String()), body)
}
