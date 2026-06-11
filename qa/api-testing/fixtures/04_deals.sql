-- QA fixtures: draft deal
-- Requires: 02_customers_vehicles.sql, 01_employee_users.sql

SET search_path TO deals, public;

INSERT INTO deals (
  id, customer_id, vehicle_id, amount, stage, assigned_to, notes, created_at, updated_at
)
VALUES (
  'a5500001-0000-4000-8000-000000000001',
  'a2200001-0000-4000-8000-000000000001',
  'a3300001-0000-4000-8000-000000000003',
  1500000,
  'draft',
  'a1100001-0000-4000-8000-000000000002',
  'QA fixture deal — use for GET/UPDATE/complete tests',
  now(), now()
)
ON CONFLICT (id) DO UPDATE SET
  stage      = EXCLUDED.stage,
  amount     = EXCLUDED.amount,
  assigned_to = EXCLUDED.assigned_to,
  notes      = EXCLUDED.notes,
  updated_at = now();

-- Optional: in_progress deal (for stage transition tests)
INSERT INTO deals (
  id, customer_id, vehicle_id, amount, stage, assigned_to, notes, created_at, updated_at
)
VALUES (
  'a5500001-0000-4000-8000-000000000002',
  'a2200001-0000-4000-8000-000000000002',
  'a3300001-0000-4000-8000-000000000001',
  890000,
  'in_progress',
  'a1100001-0000-4000-8000-000000000002',
  'QA deal in progress',
  now(), now()
)
ON CONFLICT (id) DO UPDATE SET
  stage      = EXCLUDED.stage,
  updated_at = now();
