-- Дилерские точки
CREATE TABLE IF NOT EXISTS dealer_points (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL DEFAULT '',
    address     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dealer_points_name ON dealer_points (name);
COMMENT ON TABLE dealer_points IS 'Дилерские точки (филиалы)';

-- Юридические лица
CREATE TABLE IF NOT EXISTS legal_entities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL DEFAULT '',
    inn         TEXT NOT NULL DEFAULT '',
    address     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_legal_entities_inn ON legal_entities (inn);
CREATE INDEX IF NOT EXISTS idx_legal_entities_name ON legal_entities (name);
COMMENT ON TABLE legal_entities IS 'Юридические лица';

-- Связь дилерская точка — юр. лица (M2M)
CREATE TABLE IF NOT EXISTS dealer_point_legal_entities (
    dealer_point_id  UUID NOT NULL REFERENCES dealer_points(id) ON DELETE CASCADE,
    legal_entity_id  UUID NOT NULL REFERENCES legal_entities(id) ON DELETE CASCADE,
    PRIMARY KEY (dealer_point_id, legal_entity_id)
);
CREATE INDEX IF NOT EXISTS idx_dple_dealer_point ON dealer_point_legal_entities (dealer_point_id);
CREATE INDEX IF NOT EXISTS idx_dple_legal_entity ON dealer_point_legal_entities (legal_entity_id);
COMMENT ON TABLE dealer_point_legal_entities IS 'Привязка юр. лиц к дилерским точкам';

-- Склады (по точке, юр. лицу и типу: машины / запчасти)
CREATE TABLE IF NOT EXISTS warehouses (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_point_id  UUID NOT NULL REFERENCES dealer_points(id) ON DELETE CASCADE,
    legal_entity_id  UUID NOT NULL REFERENCES legal_entities(id) ON DELETE CASCADE,
    type             TEXT NOT NULL CHECK (type IN ('cars', 'parts')),
    name             TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_warehouses_dealer_point ON warehouses (dealer_point_id);
CREATE INDEX IF NOT EXISTS idx_warehouses_legal_entity ON warehouses (legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_warehouses_type ON warehouses (type);
COMMENT ON TABLE warehouses IS 'Склады: машин (cars) или запчастей (parts) по дилерской точке и юр. лицу';

-- Привязка автомобилей к точке, юр. лицу и складу
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS dealer_point_id UUID REFERENCES dealer_points(id) ON DELETE SET NULL;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS legal_entity_id  UUID REFERENCES legal_entities(id) ON DELETE SET NULL;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS warehouse_id     UUID REFERENCES warehouses(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_dealer_point ON vehicles (dealer_point_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_legal_entity ON vehicles (legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_warehouse ON vehicles (warehouse_id);

-- Привязка запчастей к точке, юр. лицу и складу
ALTER TABLE parts ADD COLUMN IF NOT EXISTS dealer_point_id UUID REFERENCES dealer_points(id) ON DELETE SET NULL;
ALTER TABLE parts ADD COLUMN IF NOT EXISTS legal_entity_id  UUID REFERENCES legal_entities(id) ON DELETE SET NULL;
ALTER TABLE parts ADD COLUMN IF NOT EXISTS warehouse_id     UUID REFERENCES warehouses(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_parts_dealer_point ON parts (dealer_point_id);
CREATE INDEX IF NOT EXISTS idx_parts_legal_entity ON parts (legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_parts_warehouse ON parts (warehouse_id);
