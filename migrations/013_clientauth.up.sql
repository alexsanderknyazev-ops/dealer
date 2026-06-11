-- Учётные записи клиентов (отдельно от auth.users сотрудников)
CREATE SCHEMA IF NOT EXISTS clientauth;

SET search_path TO clientauth, public;

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY,
    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    full_name       TEXT NOT NULL,
    phone           TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_clientauth_users_email ON users (email);

-- email клиента уникален в профилях регистрации
SET search_path TO clients, public;

CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_email_unique ON clients (email);
