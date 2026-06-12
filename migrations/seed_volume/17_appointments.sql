-- appointments: 80 записей на ремонт
SET search_path TO appointments, public;

INSERT INTO repair_appointments (
  id, appointment_number, customer_id, vehicle_id, dealer_point_id, warehouse_id,
  scheduled_start, scheduled_end, status, complaint, notes, created_by,
  created_at, updated_at
)
SELECT
  ('90000011-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOL-APT-' || lpad(g.n::text, 5, '0'),
  ('90000005-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  ('90000006-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  ('90000003-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000023-0000-4000-8000-' || lpad(to_hex((1 + (g.n % 60)) * 2), 12, '0'))::uuid,
  now() + (g.n || ' days')::interval + time '10:00',
  now() + (g.n || ' days')::interval + time '12:00',
  (ARRAY['scheduled','scheduled','in_progress','completed','cancelled','draft'])[1 + (g.n % 6)],
  'Запись на ТО / ремонт #' || g.n,
  'Volume appointment seed',
  ('90000001-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 100)), 12, '0'))::uuid,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (appointment_number) DO UPDATE SET
  status = EXCLUDED.status,
  scheduled_start = EXCLUDED.scheduled_start,
  scheduled_end = EXCLUDED.scheduled_end,
  updated_at = now();

INSERT INTO repair_appointment_labor (id, appointment_id, work_id, description, quantity, unit_price, sort_order, created_at)
SELECT
  ('90000021-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000011-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000008-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 100)), 12, '0'))::uuid,
  'Работа по записи ' || g.n,
  1,
  2000 + (g.n % 10) * 100,
  0,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (id) DO NOTHING;

INSERT INTO repair_appointment_parts (
  id, appointment_id, part_id, warehouse_id, quantity, unit_price, notes, sort_order, created_at
)
SELECT
  ('90000031-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000011-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000007-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  ('90000023-0000-4000-8000-' || lpad(to_hex((1 + (g.n % 60)) * 2), 12, '0'))::uuid,
  1 + (g.n % 3),
  300 + (g.n % 15) * 40,
  'Запчасть по записи ' || g.n,
  0,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (id) DO NOTHING;
