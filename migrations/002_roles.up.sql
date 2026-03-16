-- Роли пользователей
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'sales';

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
    'cashier'
  )
);

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

COMMENT ON COLUMN users.role IS 'Роль: admin, manager, sales, accountant, viewer, warranty_engineer, parts_manager, storekeeper, master, consultant, cashier';
