-- reviews (B2C): 80 отзывов
SET search_path TO reviews, public;

INSERT INTO reviews (id, client_id, user_id, dealer_point_id, vehicle_id, rating, text, status, created_at, updated_at)
SELECT
  ('9000000d-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000c-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000b-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000003-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000006-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  1 + (g.n % 5),
  'Volume review text #' || g.n || ' — тестовый отзыв клиента.',
  (ARRAY['published','published','draft','rejected'])[1 + (g.n % 4)],
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (client_id, vehicle_id) DO UPDATE SET
  rating = EXCLUDED.rating,
  text = EXCLUDED.text,
  status = EXCLUDED.status,
  updated_at = now();
