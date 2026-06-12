-- Поставщики и поля документа поступления товара
SET search_path TO parts, public;

CREATE TABLE IF NOT EXISTS suppliers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL DEFAULT '',
    inn         TEXT NOT NULL DEFAULT '',
    phone       TEXT NOT NULL DEFAULT '',
    email       TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_suppliers_name ON suppliers (name);

COMMENT ON TABLE suppliers IS 'Поставщики запчастей';

ALTER TABLE movement_documents
    ADD COLUMN IF NOT EXISTS supplier_id UUID NULL REFERENCES suppliers(id),
    ADD COLUMN IF NOT EXISTS receipt_warehouse_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_movement_documents_supplier_id ON movement_documents (supplier_id);
CREATE INDEX IF NOT EXISTS idx_movement_documents_receipt_warehouse_id ON movement_documents (receipt_warehouse_id);

ALTER TABLE movement_document_lines
    ADD COLUMN IF NOT EXISTS unit_cost NUMERIC(14,2) NOT NULL DEFAULT 0;

COMMENT ON COLUMN movement_documents.supplier_id IS 'Поставщик (обязателен для receipt)';
COMMENT ON COLUMN movement_documents.receipt_warehouse_id IS 'Склад поступления (обязателен для receipt)';
COMMENT ON COLUMN movement_document_lines.unit_cost IS 'Входная стоимость единицы (для receipt)';
COMMENT ON COLUMN movement_documents.movement_type IS 'receipt — поступление товара на склад';

INSERT INTO suppliers (id, name, inn, phone, email, notes)
VALUES
    ('a8800001-0000-4000-8000-000000000001', 'ООО «АвтоПоставка»', '7701234567', '+7 (495) 111-22-33', 'supply@autopostavka.local', 'Основной поставщик запчастей'),
    ('a8800001-0000-4000-8000-000000000002', 'ИП Сидоров', '', '+7 (916) 555-00-11', 'sidorov@mail.local', '')
ON CONFLICT (id) DO NOTHING;
