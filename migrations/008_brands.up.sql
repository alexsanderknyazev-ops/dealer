-- Бренды (производители авто / запчастей)
CREATE TABLE IF NOT EXISTS brands (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_brands_name ON brands (LOWER(TRIM(name)));
COMMENT ON TABLE brands IS 'Бренды для автомобилей и запасных частей';

-- Привязка автомобилей к бренду
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS brand_id UUID REFERENCES brands(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_brand_id ON vehicles (brand_id);

-- Привязка запчастей к бренду
ALTER TABLE parts ADD COLUMN IF NOT EXISTS brand_id UUID REFERENCES brands(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_parts_brand_id ON parts (brand_id);
