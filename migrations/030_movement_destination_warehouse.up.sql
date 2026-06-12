-- Склад назначения для строк перемещения (transfer: другой склад; from_production: возврат)
SET search_path TO parts, public;

ALTER TABLE movement_document_lines
    ADD COLUMN IF NOT EXISTS destination_warehouse_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_movement_document_lines_dest_wh
    ON movement_document_lines (destination_warehouse_id);

COMMENT ON COLUMN movement_document_lines.warehouse_id IS 'Склад-источник (списание при закрытии to_production/transfer/work_order_issue)';
COMMENT ON COLUMN movement_document_lines.destination_warehouse_id IS 'Склад назначения (приход при transfer / from_production)';
