SET search_path TO parts, public;

CREATE SEQUENCE IF NOT EXISTS movement_documents_number_seq START 1;

CREATE TABLE IF NOT EXISTS movement_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_number TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'confirmed', 'cancelled')),
    movement_type   TEXT NOT NULL CHECK (movement_type IN (
        'work_order_issue', 'transfer', 'adjustment', 'receipt'
    )),
    reference_type  TEXT NOT NULL DEFAULT '',
    reference_id    UUID NULL,
    notes           TEXT NOT NULL DEFAULT '',
    created_by      UUID NULL,
    confirmed_by    UUID NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    confirmed_at    TIMESTAMPTZ NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS movement_document_lines (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id       UUID NOT NULL REFERENCES movement_documents(id) ON DELETE CASCADE,
    part_id           UUID NOT NULL REFERENCES parts(id),
    warehouse_id      UUID NOT NULL,
    quantity          INT NOT NULL CHECK (quantity > 0),
    reference_line_id UUID NULL,
    notes             TEXT NOT NULL DEFAULT '',
    sort_order        INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE stock_movements
    ADD COLUMN IF NOT EXISTS movement_document_id UUID NULL REFERENCES movement_documents(id);

CREATE INDEX idx_movement_documents_status ON movement_documents (status);
CREATE INDEX idx_movement_documents_reference ON movement_documents (reference_type, reference_id);
CREATE INDEX idx_movement_document_lines_document_id ON movement_document_lines (document_id);
CREATE INDEX idx_stock_movements_document_id ON stock_movements (movement_document_id);

COMMENT ON TABLE movement_documents IS 'Документы перемещения товара (черновик → подтверждение → списание)';
COMMENT ON TABLE movement_document_lines IS 'Строки документа перемещения';
