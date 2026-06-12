-- Реализация товара (продажа запчастей со склада)
SET search_path TO parts, public;

ALTER TABLE movement_documents
    ADD COLUMN IF NOT EXISTS customer_id UUID NULL,
    ADD COLUMN IF NOT EXISTS vehicle_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_movement_documents_customer_id ON movement_documents (customer_id);
CREATE INDEX IF NOT EXISTS idx_movement_documents_vehicle_id ON movement_documents (vehicle_id);

ALTER TABLE movement_documents DROP CONSTRAINT IF EXISTS movement_documents_movement_type_check;
ALTER TABLE movement_documents ADD CONSTRAINT movement_documents_movement_type_check
    CHECK (movement_type IN (
        'work_order_issue', 'transfer', 'adjustment', 'receipt',
        'to_production', 'from_production', 'sale'
    ));

ALTER TABLE stock_movements DROP CONSTRAINT IF EXISTS stock_movements_movement_type_check;
ALTER TABLE stock_movements ADD CONSTRAINT stock_movements_movement_type_check
    CHECK (movement_type IN (
        'work_order_issue', 'transfer', 'adjustment', 'receipt',
        'to_production', 'from_production', 'sale'
    ));

COMMENT ON COLUMN movement_documents.customer_id IS 'CRM-клиент (обязателен для sale)';
COMMENT ON COLUMN movement_documents.vehicle_id IS 'Автомобиль по VIN (необязателен для sale)';
COMMENT ON COLUMN movement_documents.movement_type IS 'sale — реализация товара (списание со склада)';
