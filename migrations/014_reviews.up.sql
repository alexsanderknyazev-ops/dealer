CREATE SCHEMA IF NOT EXISTS reviews;

SET search_path TO reviews, public;

CREATE TABLE IF NOT EXISTS reviews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       UUID NOT NULL,
    user_id         UUID NOT NULL,
    dealer_point_id UUID NOT NULL,
    vehicle_id      UUID NOT NULL,
    rating          INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text            TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'published' CHECK (status IN ('draft', 'published', 'rejected')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_reviews_client_id ON reviews (client_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews (user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_dealer_point_id ON reviews (dealer_point_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_reviews_client_vehicle ON reviews (client_id, vehicle_id);
