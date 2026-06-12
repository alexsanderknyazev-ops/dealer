-- works: папки и 100 работ СТО
SET search_path TO works, public;

INSERT INTO work_folders (id, name, parent_id, created_at, updated_at)
VALUES
  ('60000000-0000-4000-8000-000000000001', 'Техобслуживание', NULL, now(), now()),
  ('60000000-0000-4000-8000-000000000002', 'Диагностика', NULL, now(), now()),
  ('60000000-0000-4000-8000-000000000003', 'Ремонт ходовой', NULL, now(), now()),
  ('60000000-0000-4000-8000-000000000004', 'Двигатель', NULL, now(), now()),
  ('60000000-0000-4000-8000-000000000005', 'Кузовные работы', NULL, now(), now()),
  ('60000000-0000-4000-8000-000000000011', 'Замена масел', '60000000-0000-4000-8000-000000000001', now(), now()),
  ('60000000-0000-4000-8000-000000000012', 'Замена фильтров', '60000000-0000-4000-8000-000000000001', now(), now()),
  ('60000000-0000-4000-8000-000000000031', 'Тормозная система', '60000000-0000-4000-8000-000000000003', now(), now())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  parent_id = EXCLUDED.parent_id,
  updated_at = now();

INSERT INTO works (id, code, name, category, folder_id, labor_hours, unit_price, notes, created_at, updated_at)
SELECT
  ('90000008-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOL-LAB-' || lpad(g.n::text, 4, '0'),
  'Volume Work ' || g.n,
  cat.category,
  cat.folder_id::uuid,
  round((0.3 + (g.n % 20) * 0.1)::numeric, 2),
  800 + (g.n % 30) * 150,
  'Тестовая работа #' || g.n,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 100) AS g(n)
CROSS JOIN LATERAL (
  SELECT
    (ARRAY['ТО','Тормоза','Двигатель','Ходовая','Диагностика','Кузов'])[1 + (g.n % 6)] AS category,
    (ARRAY[
      '60000000-0000-4000-8000-000000000011',
      '60000000-0000-4000-8000-000000000031',
      '60000000-0000-4000-8000-000000000004',
      '60000000-0000-4000-8000-000000000003',
      '60000000-0000-4000-8000-000000000002',
      '60000000-0000-4000-8000-000000000005'
    ])[1 + (g.n % 6)] AS folder_id
) AS cat
WHERE NOT EXISTS (
  SELECT 1 FROM works w WHERE LOWER(TRIM(w.code)) = LOWER('VOL-LAB-' || lpad(g.n::text, 4, '0'))
);

UPDATE works
SET folder_id = cat.folder_id::uuid, updated_at = now()
FROM (
  SELECT
    g.n,
    (ARRAY[
      '60000000-0000-4000-8000-000000000011',
      '60000000-0000-4000-8000-000000000031',
      '60000000-0000-4000-8000-000000000004',
      '60000000-0000-4000-8000-000000000003',
      '60000000-0000-4000-8000-000000000002',
      '60000000-0000-4000-8000-000000000005'
    ])[1 + (g.n % 6)] AS folder_id
  FROM generate_series(1, 100) AS g(n)
) AS cat
WHERE works.id = ('90000008-0000-4000-8000-' || lpad(to_hex(cat.n), 12, '0'))::uuid
  AND (works.folder_id IS NULL OR works.folder_id::text NOT LIKE '60000000-%');
