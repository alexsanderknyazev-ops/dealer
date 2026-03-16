-- Транспортные средства (склад)
CREATE TABLE IF NOT EXISTS vehicles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vin         TEXT NOT NULL,
    make        TEXT NOT NULL DEFAULT '',
    model       TEXT NOT NULL DEFAULT '',
    year        INT NOT NULL CHECK (year >= 1900 AND year <= 2100),
    mileage_km  BIGINT NOT NULL DEFAULT 0 CHECK (mileage_km >= 0),
    price       NUMERIC(14,2) NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'sold', 'reserved')),
    color       TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_vehicles_vin ON vehicles (vin);
CREATE INDEX idx_vehicles_make_model ON vehicles (make, model);
CREATE INDEX idx_vehicles_status ON vehicles (status);
CREATE INDEX idx_vehicles_created_at ON vehicles (created_at);

COMMENT ON TABLE vehicles IS 'Автомобили на складе дилера';
