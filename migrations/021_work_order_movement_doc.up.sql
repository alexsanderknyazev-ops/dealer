SET search_path TO workorders, public;

ALTER TABLE work_orders
    ADD COLUMN IF NOT EXISTS movement_document_id UUID NULL,
    ADD COLUMN IF NOT EXISTS movement_document_status TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_work_orders_movement_document_id ON work_orders (movement_document_id);

COMMENT ON COLUMN work_orders.movement_document_id IS 'Документ перемещения в parts-service';
COMMENT ON COLUMN work_orders.movement_document_status IS 'draft, confirmed, cancelled — зеркало статуса документа';
