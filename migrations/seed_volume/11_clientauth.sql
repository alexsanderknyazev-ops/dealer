-- clientauth: 100 B2C пользователей
SET search_path TO clientauth, public;

INSERT INTO users (id, email, password_hash, full_name, phone, created_at, updated_at)
SELECT
  ('9000000b-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'vol.client' || g.n || '@test.dealer.local',
  '$2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy',
  'Volume Client ' || g.n,
  '+7905' || lpad(g.n::text, 7, '0'),
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 100) AS g(n)
ON CONFLICT (email) DO UPDATE SET
  full_name = EXCLUDED.full_name,
  phone = EXCLUDED.phone,
  updated_at = now();
