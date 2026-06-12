-- employees: 100 сотрудников, связь с auth volume users
SET search_path TO employees, public;

INSERT INTO employees (id, user_id, full_name, position, department, phone, active, created_at, updated_at)
SELECT
  ('90000002-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000001-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'Volume Employee ' || g.n,
  (ARRAY['master','sales','consultant','storekeeper','parts_manager'])[1 + (g.n % 5)],
  (ARRAY['СТО','Продажи','Склад','Администрация','Сервис'])[1 + (g.n % 5)],
  '+7902' || lpad(g.n::text, 7, '0'),
  g.n % 17 <> 0,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 100) AS g(n)
ON CONFLICT (user_id) DO UPDATE SET
  full_name = EXCLUDED.full_name,
  position = EXCLUDED.position,
  department = EXCLUDED.department,
  phone = EXCLUDED.phone,
  active = EXCLUDED.active,
  updated_at = now();
