-- parts: папки, 120 запчастей, 60 поставщиков, остатки на складе
SET search_path TO parts, public;

INSERT INTO part_folders (id, name, parent_id, created_at, updated_at)
VALUES
  ('50000000-0000-4000-8000-000000000001', 'Фильтры', NULL, now(), now()),
  ('50000000-0000-4000-8000-000000000002', 'Тормоза', NULL, now(), now()),
  ('50000000-0000-4000-8000-000000000003', 'Масла', NULL, now(), now()),
  ('50000000-0000-4000-8000-000000000004', 'Расходники', NULL, now(), now()),
  ('50000000-0000-4000-8000-000000000005', 'Подвеска', NULL, now(), now()),
  ('50000000-0000-4000-8000-000000000011', 'Масляные фильтры', '50000000-0000-4000-8000-000000000001', now(), now()),
  ('50000000-0000-4000-8000-000000000012', 'Воздушные фильтры', '50000000-0000-4000-8000-000000000001', now(), now()),
  ('50000000-0000-4000-8000-000000000021', 'Колодки', '50000000-0000-4000-8000-000000000002', now(), now()),
  ('50000000-0000-4000-8000-000000000022', 'Диски', '50000000-0000-4000-8000-000000000002', now(), now())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  parent_id = EXCLUDED.parent_id,
  updated_at = now();

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
  (ARRAY[
    '50000000-0000-4000-8000-000000000011',
    '50000000-0000-4000-8000-000000000021',
    '50000000-0000-4000-8000-000000000003',
    '50000000-0000-4000-8000-000000000004',
    '50000000-0000-4000-8000-000000000005'
  ])[1 + (g.n % 5)]::uuid,
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

UPDATE parts p
SET folder_id = mapping.folder_id::uuid, updated_at = now()
FROM (
  SELECT
    g.n,
    (ARRAY[
      '50000000-0000-4000-8000-000000000011',
      '50000000-0000-4000-8000-000000000021',
      '50000000-0000-4000-8000-000000000003',
      '50000000-0000-4000-8000-000000000004',
      '50000000-0000-4000-8000-000000000005'
    ])[1 + (g.n % 5)] AS folder_id
  FROM generate_series(1, 120) AS g(n)
) AS mapping
WHERE p.id = ('90000007-0000-4000-8000-' || lpad(to_hex(mapping.n), 12, '0'))::uuid;
