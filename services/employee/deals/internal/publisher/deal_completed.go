package publisher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dealer/dealer/pkg/dealevent"
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/services/deals/internal/domain"
)

type DealCompleted struct {
	producer *kafka.Producer
}

func NewDealCompleted(producer *kafka.Producer) *DealCompleted {
	return &DealCompleted{producer: producer}
}

func (p *DealCompleted) Publish(ctx context.Context, deal *domain.Deal) error {
	if p == nil || p.producer == nil || deal == nil {
		return nil
	}
	ev := dealevent.CompletedEvent{
		Event:      dealevent.Completed,
		DealID:     deal.ID.String(),
		CustomerID: deal.CustomerID.String(),
		VehicleID:  deal.VehicleID.String(),
		Amount:     deal.Amount,
		Stage:      deal.Stage,
		OccurredAt: deal.UpdatedAt.UTC().Unix(),
	}
	if ev.OccurredAt == 0 {
		ev.OccurredAt = time.Now().UTC().Unix()
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return p.producer.Publish(ctx, []byte(deal.ID.String()), body)
}
