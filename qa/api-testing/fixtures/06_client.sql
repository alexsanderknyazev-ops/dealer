-- QA fixtures: B2C client (direct DB — без Kafka)
-- Для тестов login/profile/reviews БЕЗ прохождения register flow.
-- Password: Test1234! (тот же bcrypt hash)
-- Requires: 02_customers_vehicles.sql (vehicle QAVINCLIENT001)

SET search_path TO clientauth, clients, client_statistics, public;

INSERT INTO clientauth.users (id, email, password_hash, full_name, phone, created_at, updated_at)
VALUES (
  'a7700001-0000-4000-8000-000000000001',
  'qa.client@test.local',
  '$2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy',
  'QA Client User',
  '+79003000001',
  now(), now()
)
ON CONFLICT (email) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  full_name     = EXCLUDED.full_name,
  updated_at    = now();

INSERT INTO clients.clients (id, user_id, email, full_name, phone, created_at, updated_at)
VALUES (
  'a7700001-0000-4000-8000-000000000002',
  'a7700001-0000-4000-8000-000000000001',
  'qa.client@test.local',
  'QA Client User',
  '+79003000001',
  now(), now()
)
ON CONFLICT (id) DO UPDATE SET
  email      = EXCLUDED.email,
  full_name  = EXCLUDED.full_name,
  updated_at = now();

INSERT INTO clients.client_vehicles (id, client_id, vehicle_id, vin, make, model, year, added_at)
VALUES (
  'a7700001-0000-4000-8000-000000000003',
  'a7700001-0000-4000-8000-000000000002',
  'a3300001-0000-4000-8000-000000000001',
  'QAVINCLIENT001',
  'Hyundai',
  'QA Solaris',
  2024,
  now()
)
ON CONFLICT (vin) DO UPDATE SET
  client_id  = EXCLUDED.client_id,
  vehicle_id = EXCLUDED.vehicle_id;

-- Optional: pre-seed client_statistics registration event (skip Kafka wait)
INSERT INTO client_statistics.client_registration_events (id, user_id, email, vehicle_id, occurred_at, created_at)
VALUES (
  'a8800001-0000-4000-8000-000000000001',
  'a7700001-0000-4000-8000-000000000001',
  'qa.client@test.local',
  'a3300001-0000-4000-8000-000000000001',
  now(),
  now()
)
ON CONFLICT (user_id) DO UPDATE SET
  email       = EXCLUDED.email,
  vehicle_id  = EXCLUDED.vehicle_id,
  occurred_at = EXCLUDED.occurred_at;
