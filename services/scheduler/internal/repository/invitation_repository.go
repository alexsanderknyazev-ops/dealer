package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type InvitationRepository struct {
	pool *pgxpool.Pool
}

func NewInvitationRepository(pool *pgxpool.Pool) *InvitationRepository {
	return &InvitationRepository{pool: pool}
}

// CreateFromClosedWorkOrders создаёт предложения для закрытых заказ-нарядов с зарегистрированным клиентом.
func (r *InvitationRepository) CreateFromClosedWorkOrders(ctx context.Context, limit int) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO reviews.review_invitations (
			client_id, user_id, vehicle_id, dealer_point_id,
			source_type, source_id, service_kind, status, created_at, updated_at
		)
		SELECT
			c.id,
			c.user_id,
			wo.vehicle_id,
			COALESCE(wo.dealer_point_id, v.dealer_point_id),
			'work_order',
			wo.id,
			'service',
			'pending',
			now(),
			now()
		FROM workorders.work_orders wo
		JOIN customers.customers cu ON cu.id = wo.customer_id
		JOIN vehicles.vehicles v ON v.id = wo.vehicle_id
		JOIN clients.clients c ON (
			(cu.email <> '' AND lower(trim(c.email)) = lower(trim(cu.email)))
			OR (
				cu.phone <> '' AND c.phone <> ''
				AND regexp_replace(c.phone, '[^0-9]', '', 'g') = regexp_replace(cu.phone, '[^0-9]', '', 'g')
			)
		)
		JOIN clients.client_vehicles cv ON cv.client_id = c.id AND cv.vehicle_id = wo.vehicle_id
		WHERE wo.status IN ('closed', 'paid', 'completed')
		  AND COALESCE(wo.dealer_point_id, v.dealer_point_id) IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM reviews.review_invitations ri
			WHERE ri.source_type = 'work_order' AND ri.source_id = wo.id
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM reviews.reviews rv
			WHERE rv.client_id = c.id AND rv.vehicle_id = wo.vehicle_id
		  )
		ORDER BY wo.updated_at DESC
		LIMIT %d
		ON CONFLICT (source_type, source_id) DO NOTHING
	`, limit)
	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// CreateFromCompletedDeals создаёт предложения для завершённых сделок с зарегистрированным клиентом.
func (r *InvitationRepository) CreateFromCompletedDeals(ctx context.Context, limit int) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO reviews.review_invitations (
			client_id, user_id, vehicle_id, dealer_point_id,
			source_type, source_id, service_kind, status, created_at, updated_at
		)
		SELECT
			c.id,
			c.user_id,
			d.vehicle_id,
			v.dealer_point_id,
			'deal',
			d.id,
			'sale',
			'pending',
			now(),
			now()
		FROM deals.deals d
		JOIN customers.customers cu ON cu.id = d.customer_id
		JOIN vehicles.vehicles v ON v.id = d.vehicle_id
		JOIN clients.clients c ON (
			(cu.email <> '' AND lower(trim(c.email)) = lower(trim(cu.email)))
			OR (
				cu.phone <> '' AND c.phone <> ''
				AND regexp_replace(c.phone, '[^0-9]', '', 'g') = regexp_replace(cu.phone, '[^0-9]', '', 'g')
			)
		)
		JOIN clients.client_vehicles cv ON cv.client_id = c.id AND cv.vehicle_id = d.vehicle_id
		WHERE d.stage = 'completed'
		  AND v.dealer_point_id IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM reviews.review_invitations ri
			WHERE ri.source_type = 'deal' AND ri.source_id = d.id
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM reviews.reviews rv
			WHERE rv.client_id = c.id AND rv.vehicle_id = d.vehicle_id
		  )
		ORDER BY d.updated_at DESC
		LIMIT %d
		ON CONFLICT (source_type, source_id) DO NOTHING
	`, limit)
	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
