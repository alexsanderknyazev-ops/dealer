CREATE SCHEMA IF NOT EXISTS client_statistics;

SET search_path TO client_statistics, public;

CREATE TABLE IF NOT EXISTS client_registration_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL UNIQUE,
    email        TEXT NOT NULL,
    vehicle_id   UUID,
    occurred_at  TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_client_registration_events_occurred_at ON client_registration_events (occurred_at DESC);

CREATE TABLE IF NOT EXISTS review_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id       UUID NOT NULL UNIQUE,
    client_id       UUID NOT NULL,
    user_id         UUID NOT NULL,
    dealer_point_id UUID NOT NULL,
    vehicle_id      UUID NOT NULL,
    rating          INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    status          TEXT NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_review_events_occurred_at ON review_events (occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_review_events_status ON review_events (status);
