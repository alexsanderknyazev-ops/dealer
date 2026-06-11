CREATE SCHEMA IF NOT EXISTS employee_statistics;

SET search_path TO employee_statistics, public;

CREATE TABLE IF NOT EXISTS deal_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id      UUID NOT NULL UNIQUE,
    customer_id  UUID NOT NULL,
    vehicle_id   UUID NOT NULL,
    amount       NUMERIC(14,2) NOT NULL DEFAULT 0,
    stage        TEXT NOT NULL CHECK (stage IN ('paid', 'completed')),
    occurred_at  TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_deal_events_occurred_at ON deal_events (occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_deal_events_stage ON deal_events (stage);
