package appointmentevent

import (
	"encoding/json"
	"testing"
)

func TestCreatedEventRoundtrip(t *testing.T) {
	in := CreatedEvent{
		Event:             Created,
		AppointmentID:     "11111111-1111-1111-1111-111111111111",
		AppointmentNumber: "RA-2026-00001",
		CustomerID:        "22222222-2222-2222-2222-222222222222",
		VehicleID:         "33333333-3333-3333-3333-333333333333",
		DealerPointID:     "44444444-4444-4444-4444-444444444444",
		ScheduledStart:    1785000000,
		ScheduledEnd:      1785003600,
		OccurredAt:        1785000001,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out CreatedEvent
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch:\n got %+v\nwant %+v", out, in)
	}
}

func TestCreatedEventEmptyDealerPoint(t *testing.T) {
	in := CreatedEvent{
		Event:             Created,
		AppointmentID:     "11111111-1111-1111-1111-111111111111",
		AppointmentNumber: "RA-2026-00002",
		CustomerID:        "22222222-2222-2222-2222-222222222222",
		VehicleID:         "33333333-3333-3333-3333-333333333333",
		ScheduledStart:    1785000000,
		ScheduledEnd:      1785003600,
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("empty payload")
	}
	var out CreatedEvent
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.DealerPointID != "" {
		t.Fatalf("dealer_point_id: got %q want empty (omitempty)", out.DealerPointID)
	}
}

func TestConstants(t *testing.T) {
	if TopicDefault != "repair.appointment.created.v1" {
		t.Fatalf("TopicDefault = %q", TopicDefault)
	}
	if Created != "appointment.created" {
		t.Fatalf("Created = %q", Created)
	}
}
