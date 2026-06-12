-- deals: 100 сделок
SET search_path TO deals, public;

INSERT INTO deals (id, customer_id, vehicle_id, amount, stage, assigned_to, notes, created_at, updated_at)
SELECT
  ('90000009-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000005-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  ('90000006-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  400000 + (g.n % 80) * 15000,
  (ARRAY['draft','in_progress','paid','completed','cancelled'])[1 + (g.n % 5)],
  ('90000001-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 100)), 12, '0'))::uuid,
  'Volume deal #' || g.n,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 100) AS g(n)
ON CONFLICT (id) DO UPDATE SET
  amount = EXCLUDED.amount,
  stage = EXCLUDED.stage,
  updated_at = now();
