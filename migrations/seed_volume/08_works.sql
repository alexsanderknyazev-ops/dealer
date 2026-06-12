-- works: 100 работ СТО
SET search_path TO works, public;

INSERT INTO works (id, code, name, category, labor_hours, unit_price, notes, created_at, updated_at)
SELECT
  ('90000008-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOL-LAB-' || lpad(g.n::text, 4, '0'),
  'Volume Work ' || g.n,
  (ARRAY['ТО','Тормоза','Двигатель','Ходовая','Диагностика','Кузов'])[1 + (g.n % 6)],
  round((0.3 + (g.n % 20) * 0.1)::numeric, 2),
  800 + (g.n % 30) * 150,
  'Тестовая работа #' || g.n,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 100) AS g(n)
WHERE NOT EXISTS (
  SELECT 1 FROM works w WHERE LOWER(TRIM(w.code)) = LOWER('VOL-LAB-' || lpad(g.n::text, 4, '0'))
);
