-- customers: 120 B2B клиентов
SET search_path TO customers, public;

INSERT INTO customers (id, name, email, phone, customer_type, inn, address, notes, created_at, updated_at)
SELECT
  ('90000005-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'Volume Customer ' || g.n,
  'vol.customer' || g.n || '@test.dealer.local',
  '+7903' || lpad(g.n::text, 7, '0'),
  CASE WHEN g.n % 5 = 0 THEN 'legal' ELSE 'individual' END,
  CASE WHEN g.n % 5 = 0 THEN lpad((7800000000 + g.n)::text, 10, '0') ELSE '' END,
  'г. Тестград, ул. Клиентская, ' || g.n,
  'Volume seed customer',
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 120) AS g(n)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  email = EXCLUDED.email,
  phone = EXCLUDED.phone,
  updated_at = now();
