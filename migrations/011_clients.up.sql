-- Клиенты (владельцы авто, самостоятельная регистрация)
CREATE SCHEMA IF NOT EXISTS clients;

SET search_path TO clients, public;

CREATE TABLE IF NOT EXISTS clients (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL UNIQUE,
    email       TEXT NOT NULL,
    full_name   TEXT NOT NULL,
    phone       TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_clients_email ON clients (email);
CREATE INDEX IF NOT EXISTS idx_clients_phone ON clients (phone);

CREATE TABLE IF NOT EXISTS client_vehicles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id   UUID NOT NULL REFERENCES clients (id) ON DELETE CASCADE,
    vehicle_id  UUID NOT NULL,
    vin         TEXT NOT NULL UNIQUE,
    make        TEXT NOT NULL DEFAULT '',
    model       TEXT NOT NULL DEFAULT '',
    year        INT NOT NULL DEFAULT 0,
    added_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_client_vehicles_client_id ON client_vehicles (client_id);
CREATE INDEX IF NOT EXISTS idx_client_vehicles_vehicle_id ON client_vehicles (vehicle_id);
