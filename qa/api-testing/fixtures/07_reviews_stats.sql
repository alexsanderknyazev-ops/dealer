-- QA fixtures: optional statistics / reviews baseline (empty state helpers)
-- Use when testing GET stats with known zero baseline for QA IDs only.

SET search_path TO employee_statistics, client_statistics, reviews, employee_reviews, public;

-- Ensure no stale QA deal events (re-apply fixtures cleanly)
DELETE FROM employee_statistics.deal_events
WHERE deal_id IN (
  'a5500001-0000-4000-8000-000000000001',
  'a5500001-0000-4000-8000-000000000002'
);

-- Remove QA reviews if re-seeding
DELETE FROM employee_reviews.reviews
WHERE review_id IN (
  SELECT id FROM reviews.reviews WHERE client_id = 'a7700001-0000-4000-8000-000000000002'
);

DELETE FROM client_statistics.review_events
WHERE client_id = 'a7700001-0000-4000-8000-000000000002';

DELETE FROM reviews.reviews
WHERE client_id = 'a7700001-0000-4000-8000-000000000002';

-- Sample published review (optional — for employee-reviews GET without Kafka)
INSERT INTO reviews.reviews (
  id, client_id, user_id, dealer_point_id, vehicle_id, rating, text, status, created_at, updated_at
)
VALUES (
  'a9900001-0000-4000-8000-000000000001',
  'a7700001-0000-4000-8000-000000000002',
  'a7700001-0000-4000-8000-000000000001',
  '10000000-0000-4000-8000-000000000001',
  'a3300001-0000-4000-8000-000000000001',
  4,
  'QA fixture review — good service',
  'published',
  now(), now()
)
ON CONFLICT (id) DO UPDATE SET
  rating = EXCLUDED.rating,
  text   = EXCLUDED.text,
  updated_at = now();

INSERT INTO employee_reviews.reviews (
  id, review_id, client_id, user_id, client_email, client_full_name,
  dealer_point_id, vehicle_id, vehicle_vin, vehicle_make, vehicle_model, vehicle_year,
  rating, text, status, occurred_at, created_at
)
VALUES (
  'a9900001-0000-4000-8000-000000000002',
  'a9900001-0000-4000-8000-000000000001',
  'a7700001-0000-4000-8000-000000000002',
  'a7700001-0000-4000-8000-000000000001',
  'qa.client@test.local',
  'QA Client User',
  '10000000-0000-4000-8000-000000000001',
  'a3300001-0000-4000-8000-000000000001',
  'QAVINCLIENT001',
  'Hyundai',
  'QA Solaris',
  2024,
  4,
  'QA fixture review — good service',
  'published',
  now(),
  now()
)
ON CONFLICT (review_id) DO UPDATE SET
  rating = EXCLUDED.rating,
  text   = EXCLUDED.text;

INSERT INTO client_statistics.review_events (
  id, review_id, client_id, user_id, dealer_point_id, vehicle_id, rating, status, occurred_at, created_at
)
VALUES (
  'a9900001-0000-4000-8000-000000000003',
  'a9900001-0000-4000-8000-000000000001',
  'a7700001-0000-4000-8000-000000000002',
  'a7700001-0000-4000-8000-000000000001',
  '10000000-0000-4000-8000-000000000001',
  'a3300001-0000-4000-8000-000000000001',
  4,
  'published',
  now(),
  now()
)
ON CONFLICT (review_id) DO UPDATE SET
  rating = EXCLUDED.rating;
