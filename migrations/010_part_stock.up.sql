-- Остатки запчастей по складам (одна запчасть может быть на нескольких складах)
CREATE TABLE IF NOT EXISTS part_stock (
    part_id      UUID NOT NULL REFERENCES parts(id) ON DELETE CASCADE,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    quantity     INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (part_id, warehouse_id)
);
CREATE INDEX IF NOT EXISTS idx_part_stock_part ON part_stock (part_id);
CREATE INDEX IF NOT EXISTS idx_part_stock_warehouse ON part_stock (warehouse_id);
COMMENT ON TABLE part_stock IS 'Остатки запчасти по складам (одна запчасть — несколько складов)';

-- Перенос текущих остатков из parts в part_stock
INSERT INTO part_stock (part_id, warehouse_id, quantity, created_at, updated_at)
SELECT id, warehouse_id, quantity, created_at, updated_at
FROM parts
WHERE warehouse_id IS NOT NULL
ON CONFLICT (part_id, warehouse_id) DO UPDATE SET quantity = part_stock.quantity + EXCLUDED.quantity;

-- Триггер: пересчёт parts.quantity как суммы part_stock
CREATE OR REPLACE FUNCTION sync_parts_quantity_from_stock()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE parts SET quantity = COALESCE((SELECT SUM(quantity) FROM part_stock WHERE part_id = OLD.part_id), 0), updated_at = now() WHERE id = OLD.part_id;
        RETURN OLD;
    END IF;
    UPDATE parts SET quantity = (SELECT SUM(quantity) FROM part_stock WHERE part_id = COALESCE(NEW.part_id, OLD.part_id)), updated_at = now() WHERE id = COALESCE(NEW.part_id, OLD.part_id);
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tr_part_stock_sync_quantity ON part_stock;
CREATE TRIGGER tr_part_stock_sync_quantity
AFTER INSERT OR UPDATE OR DELETE ON part_stock
FOR EACH ROW EXECUTE PROCEDURE sync_parts_quantity_from_stock();

-- Первый раз пересчитать quantity у всех частей, у которых есть part_stock
UPDATE parts p SET quantity = COALESCE((SELECT SUM(quantity) FROM part_stock WHERE part_id = p.id), 0)
WHERE EXISTS (SELECT 1 FROM part_stock WHERE part_id = p.id);
