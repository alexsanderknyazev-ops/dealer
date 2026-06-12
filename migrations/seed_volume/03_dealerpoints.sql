-- dealerpoints: 60 точек, 60 юрлиц, 120 складов
SET search_path TO dealerpoints, public;

INSERT INTO dealer_points (id, name, address, created_at, updated_at)
SELECT
  ('90000003-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'Volume Dealer Point ' || g.n,
  'г. Тестград, ул. Volume, д. ' || g.n,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 60) AS g(n)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  address = EXCLUDED.address,
  updated_at = now();

INSERT INTO legal_entities (id, name, inn, address, created_at, updated_at)
SELECT
  ('90000013-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'ООО Volume Auto ' || g.n,
  lpad((7700000000 + g.n)::text, 10, '0'),
  'г. Тестград, юр. адрес ' || g.n,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 60) AS g(n)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  inn = EXCLUDED.inn,
  address = EXCLUDED.address,
  updated_at = now();

INSERT INTO dealer_point_legal_entities (dealer_point_id, legal_entity_id)
SELECT
  ('90000003-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000013-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid
FROM generate_series(1, 60) AS g(n)
ON CONFLICT DO NOTHING;

INSERT INTO warehouses (id, dealer_point_id, legal_entity_id, type, name, created_at, updated_at)
SELECT
  ('90000023-0000-4000-8000-' || lpad(to_hex(g.n * 2 - 1), 12, '0'))::uuid,
  ('90000003-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000013-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'cars',
  'Склад авто VP-' || g.n,
  now(), now()
FROM generate_series(1, 60) AS g(n)
ON CONFLICT (id) DO NOTHING;

INSERT INTO warehouses (id, dealer_point_id, legal_entity_id, type, name, created_at, updated_at)
SELECT
  ('90000023-0000-4000-8000-' || lpad(to_hex(g.n * 2), 12, '0'))::uuid,
  ('90000003-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000013-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'parts',
  'Склад запчастей VP-' || g.n,
  now(), now()
FROM generate_series(1, 60) AS g(n)
ON CONFLICT (id) DO NOTHING;
