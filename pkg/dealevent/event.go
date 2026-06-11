package dealevent

// TopicDefault — топик успешных сделок для employee-statistics.
const TopicDefault = "deal.completed.v1"

// Completed — тип события успешной сделки (paid/completed).
const Completed = "deal.completed"

// CompletedEvent публикуется deals-service при переходе сделки в paid/completed.
type CompletedEvent struct {
	Event      string `json:"event"`
	DealID     string `json:"deal_id"`
	CustomerID string `json:"customer_id"`
	VehicleID  string `json:"vehicle_id"`
	Amount     string `json:"amount"`
	Stage      string `json:"stage"`
	OccurredAt int64  `json:"occurred_at"`
}
