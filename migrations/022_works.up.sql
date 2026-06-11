CREATE SCHEMA IF NOT EXISTS works;

SET search_path TO works, public;

CREATE TABLE IF NOT EXISTS works (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         TEXT NOT NULL DEFAULT '',
    name         TEXT NOT NULL DEFAULT '',
    category     TEXT NOT NULL DEFAULT '',
    labor_hours  NUMERIC(10,3) NOT NULL DEFAULT 1 CHECK (labor_hours > 0),
    unit_price   NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    notes        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_works_code ON works (LOWER(TRIM(code)));
CREATE INDEX IF NOT EXISTS idx_works_category ON works (category);
CREATE INDEX IF NOT EXISTS idx_works_name ON works (LOWER(name));

COMMENT ON TABLE works IS 'Справочник работ СТО (нормо-часы, цена)';
COMMENT ON COLUMN works.code IS 'Код работы, например LAB-001';
COMMENT ON COLUMN works.labor_hours IS 'Нормо-часы по умолчанию';

INSERT INTO works (code, name, category, labor_hours, unit_price, notes)
SELECT v.code, v.name, v.category, v.labor_hours, v.unit_price, v.notes
FROM (VALUES
    ('LAB-001', 'Замена масла двигателя', 'ТО', 0.5, 2500, ''),
    ('LAB-002', 'Замена воздушного фильтра', 'ТО', 0.3, 800, ''),
    ('LAB-003', 'Замена тормозных колодок передних', 'Тормоза', 1.2, 3500, ''),
    ('LAB-004', 'Диагностика подвески', 'Диагностика', 0.8, 2000, ''),
    ('LAB-005', 'Сход-развал', 'Ходовая', 1.0, 4500, ''),
    ('LAB-006', 'Замена свечей зажигания', 'Двигатель', 0.7, 1800, '')
) AS v(code, name, category, labor_hours, unit_price, notes)
WHERE NOT EXISTS (SELECT 1 FROM works w WHERE LOWER(TRIM(w.code)) = LOWER(TRIM(v.code)));
