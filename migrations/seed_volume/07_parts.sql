-- parts: 120 запчастей, 60 поставщиков, остатки на складе
SET search_path TO parts, public;

INSERT INTO suppliers (id, name, inn, phone, email, notes, created_at, updated_at)
SELECT
  ('90000017-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'Volume Supplier ' || g.n,
  lpad((7701000000 + g.n)::text, 10, '0'),
  '+7904' || lpad(g.n::text, 7, '0'),
  'vol.supplier' || g.n || '@test.dealer.local',
  'Тестовый поставщик',
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 60) AS g(n)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  updated_at = now();

INSERT INTO parts (
  id, sku, name, category, quantity, unit, price, location, notes,
  brand_id, dealer_point_id, legal_entity_id, warehouse_id, folder_id,
  created_at, updated_at
)
SELECT
  ('90000007-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOL-PART-' || lpad(g.n::text, 5, '0'),
  'Volume Part ' || g.n,
  (ARRAY['Фильтры','Тормоза','Масла','Расходники','Подвеска'])[1 + (g.n % 5)],
  10 + (g.n % 90),
  'шт',
  100 + (g.n % 40) * 25.5,
  'Стеллаж V-' || (1 + (g.n % 30)),
  'Volume parts seed',
  ('90000004-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 80)), 12, '0'))::uuid,
  ('90000003-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000013-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000023-0000-4000-8000-' || lpad(to_hex((1 + (g.n % 60)) * 2), 12, '0'))::uuid,
  '50000000-0000-4000-8000-000000000001'::uuid,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 120) AS g(n)
ON CONFLICT (sku) DO UPDATE SET
  name = EXCLUDED.name,
  quantity = EXCLUDED.quantity,
  price = EXCLUDED.price,
  updated_at = now();

INSERT INTO part_stock (part_id, warehouse_id, quantity)
SELECT
  ('90000007-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000023-0000-4000-8000-' || lpad(to_hex((1 + (g.n % 60)) * 2), 12, '0'))::uuid,
  5 + (g.n % 95)
FROM generate_series(1, 120) AS g(n)
ON CONFLICT (part_id, warehouse_id) DO UPDATE SET
  quantity = EXCLUDED.quantity;
