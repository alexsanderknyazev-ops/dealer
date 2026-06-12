-- parts: заказы поставщику и покупателя
SET search_path TO parts, public;

INSERT INTO supplier_orders (
  id, order_number, status, supplier_id, receipt_warehouse_id,
  notes, created_by, created_at, updated_at
)
SELECT
  ('90000027-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOL-SO-' || lpad(g.n::text, 5, '0'),
  (ARRAY['draft', 'linked', 'fulfilled', 'cancelled', 'draft', 'linked'])[1 + (g.n % 6)],
  ('90000017-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000023-0000-4000-8000-' || lpad(to_hex((1 + (g.n % 60)) * 2), 12, '0'))::uuid,
  'Volume supplier order #' || g.n,
  ('90000001-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 100)), 12, '0'))::uuid,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (order_number) DO UPDATE SET
  status = EXCLUDED.status,
  notes = EXCLUDED.notes,
  updated_at = now();

INSERT INTO supplier_order_lines (
  id, order_id, part_id, quantity, unit_price, notes, sort_order, created_at
)
SELECT
  ('90000028-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000027-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000007-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  1 + (g.n % 5),
  500 + (g.n % 20) * 50,
  'Позиция заказа поставщику ' || g.n,
  0,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (id) DO NOTHING;

INSERT INTO supplier_order_lines (
  id, order_id, part_id, quantity, unit_price, notes, sort_order, created_at
)
SELECT
  ('90000029-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000027-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000007-0000-4000-8000-' || lpad(to_hex(1 + ((g.n + 17) % 120)), 12, '0'))::uuid,
  1 + (g.n % 3),
  400 + (g.n % 15) * 40,
  'Доп. позиция ' || g.n,
  1,
  now()
FROM generate_series(1, 60) AS g(n)
WHERE g.n % 2 = 0
ON CONFLICT (id) DO NOTHING;

INSERT INTO customer_orders (
  id, order_number, status, customer_id, vehicle_id, issue_warehouse_id,
  notes, created_by, created_at, updated_at
)
SELECT
  ('90000037-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOL-CO-' || lpad(g.n::text, 5, '0'),
  (ARRAY['draft', 'linked', 'fulfilled', 'cancelled', 'draft', 'linked'])[1 + (g.n % 6)],
  ('90000005-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  ('90000006-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  ('90000023-0000-4000-8000-' || lpad(to_hex((1 + (g.n % 60)) * 2), 12, '0'))::uuid,
  'Volume customer order #' || g.n,
  ('90000001-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 100)), 12, '0'))::uuid,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (order_number) DO UPDATE SET
  status = EXCLUDED.status,
  notes = EXCLUDED.notes,
  updated_at = now();

INSERT INTO customer_order_lines (
  id, order_id, part_id, quantity, unit_price, notes, sort_order, created_at
)
SELECT
  ('90000038-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000037-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000007-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  1 + (g.n % 4),
  600 + (g.n % 25) * 60,
  'Позиция заказа покупателя ' || g.n,
  0,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (id) DO NOTHING;

INSERT INTO customer_order_lines (
  id, order_id, part_id, quantity, unit_price, notes, sort_order, created_at
)
SELECT
  ('90000039-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000037-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000007-0000-4000-8000-' || lpad(to_hex(1 + ((g.n + 23) % 120)), 12, '0'))::uuid,
  1 + (g.n % 2),
  550 + (g.n % 20) * 55,
  'Доп. позиция ' || g.n,
  1,
  now()
FROM generate_series(1, 60) AS g(n)
WHERE g.n % 2 = 0
ON CONFLICT (id) DO NOTHING;

-- Связь части заказов поставщику с заказами покупателя (первые 40 пар)
UPDATE supplier_orders so
SET customer_order_id = co.id, updated_at = now()
FROM customer_orders co
WHERE so.order_number = replace(co.order_number, 'VOL-CO-', 'VOL-SO-')
  AND so.id::text LIKE '90000027-%'
  AND co.id::text LIKE '90000037-%'
  AND so.customer_order_id IS NULL
  AND right(so.order_number, 5)::int <= 40;
