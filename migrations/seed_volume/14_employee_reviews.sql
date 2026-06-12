-- employee_reviews: 80 отзывов для аналитики сотрудников
SET search_path TO employee_reviews, public;

INSERT INTO reviews (
  id, review_id, client_id, user_id, client_email, client_full_name,
  dealer_point_id, vehicle_id, vehicle_vin, vehicle_make, vehicle_model, vehicle_year,
  rating, text, status, occurred_at, created_at
)
SELECT
  ('9000000e-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000d-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000c-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000b-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'vol.client' || g.n || '@test.dealer.local',
  'Volume Client ' || g.n,
  ('90000003-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000006-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOLVIN' || lpad(g.n::text, 11, '0'),
  'VolumeMake-' || (1 + (g.n % 10)),
  'Model-' || g.n,
  2015 + (g.n % 11),
  1 + (g.n % 5),
  'Employee contour review #' || g.n,
  'published',
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (review_id) DO UPDATE SET
  rating = EXCLUDED.rating,
  text = EXCLUDED.text,
  status = EXCLUDED.status;
