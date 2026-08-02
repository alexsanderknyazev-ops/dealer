// Package appointmentevent описывает события записей на ремонт,
// публикуемые appointments-service в Kafka.
package appointmentevent

// TopicDefault — топик событий создания записи на ремонт.
const TopicDefault = "repair.appointment.created.v1"

// Created — тип события создания записи на ремонт.
const Created = "appointment.created"

// CreatedEvent публикуется appointments-service после успешного
// создания записи на ремонт (repair appointment).
type CreatedEvent struct {
	Event             string `json:"event"`
	AppointmentID     string `json:"appointment_id"`
	AppointmentNumber string `json:"appointment_number"`
	CustomerID        string `json:"customer_id"`
	VehicleID         string `json:"vehicle_id"`
	DealerPointID     string `json:"dealer_point_id,omitempty"`
	ScheduledStart    int64  `json:"scheduled_start"`
	ScheduledEnd      int64  `json:"scheduled_end"`
	OccurredAt        int64  `json:"occurred_at"`
}
