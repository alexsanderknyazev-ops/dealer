-- Роль client для владельцев авто (самостоятельная регистрация)
SET search_path TO auth, public;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (
  role IN (
    'admin',
    'manager',
    'sales',
    'accountant',
    'viewer',
    'warranty_engineer',
    'parts_manager',
    'storekeeper',
    'master',
    'consultant',
    'cashier',
    'client'
  )
);

COMMENT ON COLUMN users.role IS 'Роль: staff-роли или client (владелец авто)';
