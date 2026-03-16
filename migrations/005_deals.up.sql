-- Сделки (продажи)
CREATE TABLE IF NOT EXISTS deals (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id  UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    vehicle_id   UUID NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
    amount       NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    stage        TEXT NOT NULL DEFAULT 'draft' CHECK (stage IN ('draft', 'in_progress', 'paid', 'completed', 'cancelled')),
    assigned_to  UUID NULL,  -- user_id из auth (нет FK, другая БД)
    notes        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_deals_customer_id ON deals (customer_id);
CREATE INDEX idx_deals_vehicle_id ON deals (vehicle_id);
CREATE INDEX idx_deals_stage ON deals (stage);
CREATE INDEX idx_deals_created_at ON deals (created_at);

COMMENT ON TABLE deals IS 'Сделки купли-продажи: клиент, автомобиль, сумма, этап';
