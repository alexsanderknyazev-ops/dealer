-- client_statistics: 80 регистраций + 80 review events
SET search_path TO client_statistics, public;

INSERT INTO client_registration_events (id, user_id, email, vehicle_id, occurred_at, created_at)
SELECT
  ('90000010-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000b-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'vol.client' || g.n || '@test.dealer.local',
  ('90000006-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (user_id) DO UPDATE SET
  email = EXCLUDED.email,
  vehicle_id = EXCLUDED.vehicle_id,
  occurred_at = EXCLUDED.occurred_at;

INSERT INTO review_events (
  id, review_id, client_id, user_id, dealer_point_id, vehicle_id,
  rating, status, occurred_at, created_at
)
SELECT
  ('90000020-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000d-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000c-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000b-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000003-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000006-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  1 + (g.n % 5),
  'published',
  now() - (g.n || ' hours')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (review_id) DO UPDATE SET
  rating = EXCLUDED.rating,
  status = EXCLUDED.status,
  occurred_at = EXCLUDED.occurred_at;
