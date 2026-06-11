-- QA fixtures: works catalog + employee directory (fixed UUIDs)
-- Requires: 01_employee_users.sql (auth.users)

SET search_path TO works, employees, public;

-- Works catalog (LAB-QA-* for tests; do not collide with migration seed LAB-001..006)
INSERT INTO works (id, code, name, category, labor_hours, unit_price, notes, created_at, updated_at)
VALUES
  (
    'a8800001-0000-4000-8000-000000000001',
    'LAB-QA-004',
    'QA Диагностика тормозов',
    'Диагностика',
    0.8,
    2000,
    'QA fixture work for WO labor line',
    now(), now()
  ),
  (
    'a8800001-0000-4000-8000-000000000002',
    'LAB-QA-001',
    'QA Замена масла',
    'ТО',
    0.5,
    2500,
    'QA fixture for create WO tests',
    now(), now()
  )
ON CONFLICT (id) DO UPDATE SET
  code        = EXCLUDED.code,
  name        = EXCLUDED.name,
  category    = EXCLUDED.category,
  labor_hours = EXCLUDED.labor_hours,
  unit_price  = EXCLUDED.unit_price,
  notes       = EXCLUDED.notes,
  updated_at  = now();

-- Employees linked to QA auth users (ResolveRef accepts user_id or employee id)
INSERT INTO employees (id, user_id, full_name, position, department, phone, active, created_at, updated_at)
VALUES
  (
    'a9900001-0000-4000-8000-000000000001',
    'a1100001-0000-4000-8000-000000000001',
    'QA Admin',
    'admin',
    'Администрация',
    '+79001000001',
    true,
    now(), now()
  ),
  (
    'a9900001-0000-4000-8000-000000000003',
    'a1100001-0000-4000-8000-000000000003',
    'QA Master',
    'master',
    'СТО',
    '+79001000003',
    true,
    now(), now()
  ),
  (
    'a9900001-0000-4000-8000-000000000005',
    'a1100001-0000-4000-8000-000000000005',
    'QA Storekeeper',
    'storekeeper',
    'Склад',
    '+79001000005',
    true,
    now(), now()
  )
ON CONFLICT (id) DO UPDATE SET
  user_id    = EXCLUDED.user_id,
  full_name  = EXCLUDED.full_name,
  position   = EXCLUDED.position,
  department = EXCLUDED.department,
  phone      = EXCLUDED.phone,
  active     = EXCLUDED.active,
  updated_at = now();

-- Ensure user_id uniqueness for upsert path (migration may have inserted without fixed id)
INSERT INTO employees (user_id, full_name, position, department, phone, active)
SELECT v.user_id, v.full_name, v.position, v.department, v.phone, true
FROM (VALUES
  ('a1100001-0000-4000-8000-000000000002'::uuid, 'QA Sales', 'sales', 'Продажи', '+79001000002'),
  ('a1100001-0000-4000-8000-000000000004'::uuid, 'QA Parts Manager', 'parts_manager', 'Склад', '+79001000004')
) AS v(user_id, full_name, position, department, phone)
WHERE NOT EXISTS (SELECT 1 FROM employees e WHERE e.user_id = v.user_id);
