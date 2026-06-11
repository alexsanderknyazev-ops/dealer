CREATE SCHEMA IF NOT EXISTS employee_reviews;

SET search_path TO employee_reviews, public;

CREATE TABLE IF NOT EXISTS reviews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id       UUID NOT NULL UNIQUE,
    client_id       UUID NOT NULL,
    user_id         UUID NOT NULL,
    client_email    TEXT NOT NULL DEFAULT '',
    client_full_name TEXT NOT NULL DEFAULT '',
    dealer_point_id UUID NOT NULL,
    vehicle_id      UUID NOT NULL,
    vehicle_vin     TEXT NOT NULL DEFAULT '',
    vehicle_make    TEXT NOT NULL DEFAULT '',
    vehicle_model   TEXT NOT NULL DEFAULT '',
    vehicle_year    INT NOT NULL DEFAULT 0,
    rating          INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text            TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_employee_reviews_client_id ON reviews (client_id);
CREATE INDEX IF NOT EXISTS idx_employee_reviews_dealer_point_id ON reviews (dealer_point_id);
CREATE INDEX IF NOT EXISTS idx_employee_reviews_status ON reviews (status);
CREATE INDEX IF NOT EXISTS idx_employee_reviews_occurred_at ON reviews (occurred_at DESC);
