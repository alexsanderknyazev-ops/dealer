-- employee_statistics: 80 событий завершённых сделок
SET search_path TO employee_statistics, public;

INSERT INTO deal_events (id, deal_id, customer_id, vehicle_id, amount, stage, occurred_at, created_at)
SELECT
  ('9000000f-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000009-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000005-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  ('90000006-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 120)), 12, '0'))::uuid,
  400000 + (g.n % 80) * 15000,
  (ARRAY['paid','completed'])[1 + (g.n % 2)],
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (deal_id) DO UPDATE SET
  amount = EXCLUDED.amount,
  stage = EXCLUDED.stage,
  occurred_at = EXCLUDED.occurred_at;
