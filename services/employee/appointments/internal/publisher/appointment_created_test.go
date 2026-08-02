package publisher

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/pkg/appointmentevent"
	"github.com/dealer/dealer/services/appointments/internal/domain"
)

type recordingProducer struct {
	key  []byte
	body []byte
}

func (r *recordingProducer) Publish(_ context.Context, key, body []byte) error {
	r.key = key
	r.body = body
	return nil
}

func TestPublish_NilDealerPointIDDoesNotPanic(t *testing.T) {
	producer := &recordingProducer{}
	p := NewAppointmentCreated(producer)
	a := &domain.RepairAppointment{
		ID:                uuid.New(),
		AppointmentNumber: "RA-TEST-1",
		CustomerID:        uuid.New(),
		VehicleID:         uuid.New(),
		ScheduledStart:    time.Now().UTC(),
		ScheduledEnd:      time.Now().UTC().Add(time.Hour),
		CreatedAt:         time.Now().UTC(),
	}
	if err := p.Publish(context.Background(), a); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	var ev appointmentevent.CreatedEvent
	if err := json.Unmarshal(producer.body, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.DealerPointID != "" {
		t.Errorf("DealerPointID = %q, want empty", ev.DealerPointID)
	}
	if ev.AppointmentNumber != "RA-TEST-1" {
		t.Errorf("AppointmentNumber = %q", ev.AppointmentNumber)
	}
}

func TestPublish_NilPublisherIsNoop(t *testing.T) {
	var p *AppointmentCreated
	if err := p.Publish(context.Background(), &domain.RepairAppointment{ID: uuid.New()}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
}

func TestPublish_WithDealerPointID(t *testing.T) {
	producer := &recordingProducer{}
	p := NewAppointmentCreated(producer)
	pointID := uuid.New()
	a := &domain.RepairAppointment{
		ID:             uuid.New(),
		CustomerID:     uuid.New(),
		VehicleID:      uuid.New(),
		DealerPointID:  &pointID,
		ScheduledStart: time.Now().UTC(),
		ScheduledEnd:   time.Now().UTC().Add(time.Hour),
		CreatedAt:      time.Now().UTC(),
	}
	if err := p.Publish(context.Background(), a); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	var ev appointmentevent.CreatedEvent
	if err := json.Unmarshal(producer.body, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.DealerPointID != pointID.String() {
		t.Errorf("DealerPointID = %q, want %q", ev.DealerPointID, pointID.String())
	}
}
