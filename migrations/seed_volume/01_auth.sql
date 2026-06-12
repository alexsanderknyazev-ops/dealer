-- auth: 100 сотрудников (volume namespace 90000001-*)
-- Password: Test1234!
SET search_path TO auth, public;

INSERT INTO users (id, email, password_hash, name, phone, role, created_at, updated_at)
SELECT
  ('90000001-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'vol.employee' || g.n || '@test.dealer.local',
  '$2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy',
  'Volume Employee ' || g.n,
  '+7901' || lpad(g.n::text, 7, '0'),
  (ARRAY['sales','master','manager','parts_manager','storekeeper','consultant','cashier','accountant','viewer','warranty_engineer'])[1 + (g.n % 10)],
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 100) AS g(n)
ON CONFLICT (email) DO UPDATE SET
  name = EXCLUDED.name,
  phone = EXCLUDED.phone,
  role = EXCLUDED.role,
  updated_at = now();
