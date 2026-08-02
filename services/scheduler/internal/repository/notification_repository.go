package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

// CreateRepairAppointmentBooked создаёт уведомление «вы записаны на ремонт»
// для зарегистрированного клиента сразу после создания записи.
func (r *NotificationRepository) CreateRepairAppointmentBooked(ctx context.Context, appointmentID uuid.UUID) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO clients.client_notifications (
			client_id, user_id, kind, source_type, source_id, title, body, status, created_at, updated_at
		)
		SELECT
			c.id,
			c.user_id,
			'repair_appointment_booked',
			'repair_appointment',
			ra.id,
			'Вы записаны на ремонт',
			'Запись на ремонт ' || ra.appointment_number || ' ' ||
			to_char(ra.scheduled_start AT TIME ZONE 'UTC', 'DD.MM.YYYY HH24:MI') || ' UTC.',
			'unread',
			now(),
			now()
		FROM appointments.repair_appointments ra
		JOIN customers.customers cu ON cu.id = ra.customer_id
		JOIN clients.clients c ON (
			(cu.email <> '' AND lower(trim(c.email)) = lower(trim(cu.email)))
			OR (
				cu.phone <> '' AND c.phone <> ''
				AND regexp_replace(c.phone, '[^0-9]', '', 'g') = regexp_replace(cu.phone, '[^0-9]', '', 'g')
			)
		)
		WHERE ra.id = '%s'
		ON CONFLICT (kind, source_type, source_id) DO NOTHING
	`, appointmentID)
	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// CreateFromClosedCustomerOrderReceipts уведомляет клиента, когда по связанному заказу поставщику закрыто поступление.
func (r *NotificationRepository) CreateFromClosedCustomerOrderReceipts(ctx context.Context, limit int) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO clients.client_notifications (
			client_id, user_id, kind, source_type, source_id, title, body, status, created_at, updated_at
		)
		SELECT
			c.id,
			c.user_id,
			'customer_order_receipt',
			'customer_order',
			co.id,
			'Запчасти поступили',
			'По заказу ' || co.order_number || ' запчасти поступили на склад и готовы к получению.',
			'unread',
			now(),
			now()
		FROM parts.movement_documents md
		JOIN parts.supplier_orders so ON so.id = md.reference_id AND md.reference_type = 'supplier_order'
		JOIN parts.customer_orders co ON co.id = so.customer_order_id
		JOIN customers.customers cu ON cu.id = co.customer_id
		JOIN clients.clients c ON (
			(cu.email <> '' AND lower(trim(c.email)) = lower(trim(cu.email)))
			OR (
				cu.phone <> '' AND c.phone <> ''
				AND regexp_replace(c.phone, '[^0-9]', '', 'g') = regexp_replace(cu.phone, '[^0-9]', '', 'g')
			)
		)
		WHERE md.movement_type = 'receipt'
		  AND md.status = 'closed'
		  AND so.customer_order_id IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM clients.client_notifications n
			WHERE n.kind = 'customer_order_receipt'
			  AND n.source_type = 'customer_order'
			  AND n.source_id = co.id
		  )
		ORDER BY md.updated_at DESC
		LIMIT %d
		ON CONFLICT (kind, source_type, source_id) DO NOTHING
	`, limit)
	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// CreateRepairAppointmentReminders создаёт напоминания за ~24ч и ~1ч до записи на ремонт.
func (r *NotificationRepository) CreateRepairAppointmentReminders(ctx context.Context, limit int) (int64, error) {
	var total int64
	day, err := r.insertRepairAppointmentReminders(ctx, limit, "repair_appointment_reminder_24h",
		`ra.scheduled_start > now() + interval '23 hours' AND ra.scheduled_start <= now() + interval '25 hours'`)
	if err != nil {
		return 0, err
	}
	total += day
	hour, err := r.insertRepairAppointmentReminders(ctx, limit, "repair_appointment_reminder_1h",
		`ra.scheduled_start > now() + interval '55 minutes' AND ra.scheduled_start <= now() + interval '65 minutes'`)
	if err != nil {
		return 0, err
	}
	total += hour
	return total, nil
}

func (r *NotificationRepository) insertRepairAppointmentReminders(ctx context.Context, limit int, kind, timeWindow string) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO clients.client_notifications (
			client_id, user_id, kind, source_type, source_id, title, body, status, created_at, updated_at
		)
		SELECT
			c.id,
			c.user_id,
			'%s',
			'repair_appointment',
			ra.id,
			CASE WHEN '%s' LIKE '%%24h%%' THEN 'Напоминание: запись завтра' ELSE 'Напоминание: запись через час' END,
			'Запись на ремонт ' || ra.appointment_number || ' ' ||
			to_char(ra.scheduled_start AT TIME ZONE 'UTC', 'DD.MM.YYYY HH24:MI') || ' UTC.',
			'unread',
			now(),
			now()
		FROM appointments.repair_appointments ra
		JOIN customers.customers cu ON cu.id = ra.customer_id
		JOIN clients.clients c ON (
			(cu.email <> '' AND lower(trim(c.email)) = lower(trim(cu.email)))
			OR (
				cu.phone <> '' AND c.phone <> ''
				AND regexp_replace(c.phone, '[^0-9]', '', 'g') = regexp_replace(cu.phone, '[^0-9]', '', 'g')
			)
		)
		WHERE ra.status IN ('scheduled', 'in_progress')
		  AND %s
		  AND NOT EXISTS (
			SELECT 1 FROM clients.client_notifications n
			WHERE n.kind = '%s' AND n.source_type = 'repair_appointment' AND n.source_id = ra.id
		  )
		ORDER BY ra.scheduled_start ASC
		LIMIT %d
		ON CONFLICT (kind, source_type, source_id) DO NOTHING
	`, kind, kind, timeWindow, kind, limit)
	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
