-- Предложения клиентам оставить отзыв после обслуживания или покупки
SET search_path TO reviews, public;

CREATE TABLE IF NOT EXISTS review_invitations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       UUID NOT NULL,
    user_id         UUID NOT NULL,
    vehicle_id      UUID NOT NULL,
    dealer_point_id UUID NOT NULL,
    source_type     TEXT NOT NULL CHECK (source_type IN ('work_order', 'deal')),
    source_id       UUID NOT NULL,
    service_kind    TEXT NOT NULL CHECK (service_kind IN ('service', 'sale')),
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'dismissed', 'completed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_type, source_id)
);

CREATE INDEX IF NOT EXISTS idx_review_invitations_user_status ON review_invitations (user_id, status);
CREATE INDEX IF NOT EXISTS idx_review_invitations_client_id ON review_invitations (client_id);
CREATE INDEX IF NOT EXISTS idx_review_invitations_vehicle_id ON review_invitations (vehicle_id);

COMMENT ON TABLE review_invitations IS 'Предложения зарегистрированным клиентам оценить обслуживание или покупку';
COMMENT ON COLUMN review_invitations.source_type IS 'work_order — заказ-наряд, deal — сделка';
COMMENT ON COLUMN review_invitations.service_kind IS 'service — СТО, sale — продажа авто';
