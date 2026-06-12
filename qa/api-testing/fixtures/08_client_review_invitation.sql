-- QA: закрытый заказ-наряд → предложение оставить отзыв для qa.client@test.local
-- Requires: 02, 05, 06 fixtures. Удаляет fixture-отзыв 07 (блокирует invitation).
-- После применения scheduler создаст review_invitations (или выполните INSERT ниже вручную).

SET search_path TO customers, workorders, reviews, employee_reviews, client_statistics, public;

-- CRM-клиент с тем же email/phone, что B2C qa.client
INSERT INTO customers (id, name, email, phone, customer_type, inn, address, notes, created_at, updated_at)
VALUES (
  'a2200001-0000-4000-8000-000000000099',
  'QA Client User',
  'qa.client@test.local',
  '+79003000001',
  'individual',
  '',
  '',
  'CRM customer linked to B2C qa.client for review invitation demo',
  now(), now()
)
ON CONFLICT (id) DO UPDATE SET
  email      = EXCLUDED.email,
  phone      = EXCLUDED.phone,
  updated_at = now();

-- Закрыть WO-QA-0001 на авто QAVINCLIENT001
UPDATE work_orders
SET
  customer_id = 'a2200001-0000-4000-8000-000000000099',
  status      = 'closed',
  closed_at   = COALESCE(closed_at, now()),
  updated_at  = now()
WHERE id = 'a6600001-0000-4000-8000-000000000001';

-- Fixture-отзыв из 07_reviews_stats блокирует создание invitation для того же авто
DELETE FROM employee_reviews.reviews
WHERE review_id = 'a9900001-0000-4000-8000-000000000001';

DELETE FROM client_statistics.review_events
WHERE review_id = 'a9900001-0000-4000-8000-000000000001';

DELETE FROM reviews
WHERE id = 'a9900001-0000-4000-8000-000000000001';

DELETE FROM review_invitations
WHERE source_type = 'work_order'
  AND source_id = 'a6600001-0000-4000-8000-000000000001';

-- То же, что scheduler CreateFromClosedWorkOrders (без ожидания poll)
INSERT INTO review_invitations (
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
FROM work_orders wo
JOIN customers cu ON cu.id = wo.customer_id
JOIN vehicles.vehicles v ON v.id = wo.vehicle_id
JOIN clients.clients c ON (
  (cu.email <> '' AND lower(trim(c.email)) = lower(trim(cu.email)))
  OR (
    cu.phone <> '' AND c.phone <> ''
    AND regexp_replace(c.phone, '[^0-9]', '', 'g') = regexp_replace(cu.phone, '[^0-9]', '', 'g')
  )
)
JOIN clients.client_vehicles cv ON cv.client_id = c.id AND cv.vehicle_id = wo.vehicle_id
WHERE wo.id = 'a6600001-0000-4000-8000-000000000001'
  AND wo.status IN ('closed', 'paid', 'completed')
  AND COALESCE(wo.dealer_point_id, v.dealer_point_id) IS NOT NULL
ON CONFLICT (source_type, source_id) DO NOTHING;
