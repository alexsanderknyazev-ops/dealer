-- Перемещение в производство и извлечение обратно на склад
SET search_path TO parts, public;

ALTER TABLE movement_documents
    ADD COLUMN IF NOT EXISTS parent_document_id UUID NULL REFERENCES movement_documents(id);

CREATE INDEX IF NOT EXISTS idx_movement_documents_parent_id
    ON movement_documents (parent_document_id);

ALTER TABLE movement_documents DROP CONSTRAINT IF EXISTS movement_documents_movement_type_check;
ALTER TABLE movement_documents ADD CONSTRAINT movement_documents_movement_type_check
    CHECK (movement_type IN (
        'work_order_issue', 'transfer', 'adjustment', 'receipt',
        'to_production', 'from_production'
    ));

ALTER TABLE stock_movements DROP CONSTRAINT IF EXISTS stock_movements_movement_type_check;
ALTER TABLE stock_movements ADD CONSTRAINT stock_movements_movement_type_check
    CHECK (movement_type IN (
        'work_order_issue', 'transfer', 'adjustment', 'receipt',
        'to_production', 'from_production'
    ));

COMMENT ON COLUMN movement_documents.parent_document_id IS 'Для from_production — закрытый документ to_production';
COMMENT ON COLUMN movement_documents.movement_type IS 'to_production — в производство; from_production — извлечение на склад';
