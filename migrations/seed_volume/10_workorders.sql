-- workorders: 80 заказ-нарядов + строки работ
SET search_path TO workorders, public;

INSERT INTO work_orders (
  id, order_number, customer_id, vehicle_id, dealer_point_id, warehouse_id,
  repair_type, status, service_advisor_id, complaint, diagnosis,
  mileage_km, labor_cost, parts_cost, total_cost, notes, created_at, updated_at
)
SELECT
  ('9000000a-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOL-WO-' || lpad(g.n::text, 5, '0'),
  ('90000005-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  ('90000006-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  ('90000003-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000023-0000-4000-8000-' || lpad(to_hex((1 + (g.n % 60)) * 2), 12, '0'))::uuid,
  (ARRAY['commercial','warranty_manufacturer','pre_sale','maintenance'])[1 + (g.n % 4)],
  (ARRAY['draft','in_progress','completed','closed','paid'])[1 + (g.n % 5)],
  ('90000002-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 100)), 12, '0'))::uuid,
  'Жалоба клиента #' || g.n,
  'Диагностика volume #' || g.n,
  (g.n * 2100) % 300000,
  2000 + (g.n % 15) * 500,
  1500 + (g.n % 20) * 300,
  3500 + (g.n % 25) * 400,
  'Volume work order',
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (order_number) DO UPDATE SET
  status = EXCLUDED.status,
  total_cost = EXCLUDED.total_cost,
  updated_at = now();

INSERT INTO work_order_labor (id, work_order_id, description, quantity, unit_price, amount, executor_id, sort_order, created_at, updated_at)
SELECT
  ('9000001a-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000a-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'Работа по заказ-наряду ' || g.n,
  1,
  2500 + (g.n % 10) * 200,
  2500 + (g.n % 10) * 200,
  ('90000002-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 100)), 12, '0'))::uuid,
  0,
  now(), now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (id) DO NOTHING;
