package publisher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dealer/dealer/client-reviews-service/internal/domain"
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/reviewevent"
)

type ReviewPublished struct {
	producer *kafka.Producer
}

func NewReviewPublished(producer *kafka.Producer) *ReviewPublished {
	return &ReviewPublished{producer: producer}
}

func (p *ReviewPublished) Publish(ctx context.Context, review *domain.Review) error {
	if p == nil || p.producer == nil || review == nil {
		return nil
	}
	ev := reviewevent.PublishedEvent{
		Event:         reviewevent.Published,
		ReviewID:      review.ID.String(),
		ClientID:      review.ClientID.String(),
		UserID:        review.UserID.String(),
		DealerPointID: review.DealerPointID.String(),
		VehicleID:     review.VehicleID.String(),
		Rating:        review.Rating,
		Status:        review.Status,
		OccurredAt:    review.CreatedAt.UTC().Unix(),
	}
	if ev.OccurredAt == 0 {
		ev.OccurredAt = time.Now().UTC().Unix()
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return p.producer.Publish(ctx, []byte(review.ID.String()), body)
}
