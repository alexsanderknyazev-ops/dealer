package reviewevent

// TopicDefault — топик опубликованных отзывов для client-statistics.
const TopicDefault = "review.published.v1"

// Published — тип события нового отзыва.
const Published = "review.published"

// PublishedEvent публикуется client-reviews-service после создания отзыва.
type PublishedEvent struct {
	Event         string `json:"event"`
	ReviewID      string `json:"review_id"`
	ClientID      string `json:"client_id"`
	UserID        string `json:"user_id"`
	DealerPointID string `json:"dealer_point_id"`
	VehicleID     string `json:"vehicle_id"`
	Rating        int32  `json:"rating"`
	Status        string `json:"status"`
	OccurredAt    int64  `json:"occurred_at"`
}
