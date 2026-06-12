-- Стоимость нормо-часа по бренду и дилерской точке (гарантийный / коммерческий)
SET search_path TO brands, public;

CREATE TABLE IF NOT EXISTS brand_labor_rates (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brand_id                UUID NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    dealer_point_id         UUID NOT NULL,
    warranty_hour_price     NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (warranty_hour_price >= 0),
    commercial_hour_price   NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (commercial_hour_price >= 0),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (brand_id, dealer_point_id)
);

CREATE INDEX IF NOT EXISTS idx_brand_labor_rates_brand_id ON brand_labor_rates (brand_id);
CREATE INDEX IF NOT EXISTS idx_brand_labor_rates_dealer_point_id ON brand_labor_rates (dealer_point_id);

COMMENT ON TABLE brand_labor_rates IS 'Стоимость нормо-часа: гарантийный и коммерческий по паре бренд + дилерская точка';
