-- Клиенты (физ. и юр. лица)
CREATE TABLE IF NOT EXISTS customers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    email       TEXT NOT NULL DEFAULT '',
    phone       TEXT NOT NULL DEFAULT '',
    customer_type TEXT NOT NULL DEFAULT 'individual' CHECK (customer_type IN ('individual', 'legal')),
    inn         TEXT NOT NULL DEFAULT '',
    address     TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_customers_name ON customers (name);
CREATE INDEX idx_customers_email ON customers (email);
CREATE INDEX idx_customers_phone ON customers (phone);
CREATE INDEX idx_customers_created_at ON customers (created_at);
