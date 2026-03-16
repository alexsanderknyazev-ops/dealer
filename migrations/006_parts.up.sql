-- Запасные части (склад запчастей)
CREATE TABLE IF NOT EXISTS parts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku         TEXT NOT NULL,
    name        TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT '',
    quantity    INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    unit        TEXT NOT NULL DEFAULT 'шт',
    price       NUMERIC(14,2) NOT NULL DEFAULT 0,
    location    TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_parts_sku ON parts (sku);
CREATE INDEX idx_parts_category ON parts (category);
CREATE INDEX idx_parts_name ON parts (name);
CREATE INDEX idx_parts_created_at ON parts (created_at);

COMMENT ON TABLE parts IS 'Запасные части на складе дилера';
