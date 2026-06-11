CREATE SCHEMA IF NOT EXISTS employees;

SET search_path TO employees, public;

CREATE TABLE IF NOT EXISTS employees (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NULL UNIQUE,
    full_name   TEXT NOT NULL DEFAULT '',
    position    TEXT NOT NULL DEFAULT '',
    department  TEXT NOT NULL DEFAULT '',
    phone       TEXT NOT NULL DEFAULT '',
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_employees_user_id ON employees (user_id);
CREATE INDEX IF NOT EXISTS idx_employees_active ON employees (active);
CREATE INDEX IF NOT EXISTS idx_employees_full_name ON employees (LOWER(full_name));

COMMENT ON TABLE employees IS 'Справочник сотрудников СТО (ФИО, связь с auth.users)';
COMMENT ON COLUMN employees.user_id IS 'UUID пользователя auth.users';
COMMENT ON COLUMN employees.full_name IS 'ФИО сотрудника';

-- Сотрудники из QA-фикстур (auth.users)
INSERT INTO employees (user_id, full_name, position, department, phone, active)
SELECT v.user_id, v.full_name, v.position, v.department, v.phone, true
FROM (VALUES
    ('a1100001-0000-4000-8000-000000000001'::uuid, 'QA Admin', 'admin', 'Администрация', '+79001000001'),
    ('a1100001-0000-4000-8000-000000000002'::uuid, 'QA Sales', 'sales', 'Продажи', '+79001000002'),
    ('a1100001-0000-4000-8000-000000000003'::uuid, 'QA Master', 'master', 'СТО', '+79001000003'),
    ('a1100001-0000-4000-8000-000000000004'::uuid, 'QA Parts Manager', 'parts_manager', 'Склад', '+79001000004'),
    ('a1100001-0000-4000-8000-000000000005'::uuid, 'QA Storekeeper', 'storekeeper', 'Склад', '+79001000005')
) AS v(user_id, full_name, position, department, phone)
WHERE NOT EXISTS (
    SELECT 1 FROM employees e WHERE e.user_id = v.user_id
);
