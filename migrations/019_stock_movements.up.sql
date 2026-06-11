SET search_path TO parts, public;

CREATE TABLE IF NOT EXISTS stock_movements (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    part_id            UUID NOT NULL REFERENCES parts(id),
    warehouse_id       UUID NOT NULL,
    quantity           INT NOT NULL CHECK (quantity <> 0),
    movement_type      TEXT NOT NULL CHECK (movement_type IN (
        'work_order_issue', 'transfer', 'adjustment', 'receipt'
    )),
    reference_type     TEXT NOT NULL DEFAULT '',
    reference_id       UUID NULL,
    reference_line_id  UUID NULL,
    notes              TEXT NOT NULL DEFAULT '',
    created_by         UUID NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_stock_movements_part_id ON stock_movements (part_id);
CREATE INDEX idx_stock_movements_warehouse_id ON stock_movements (warehouse_id);
CREATE INDEX idx_stock_movements_reference ON stock_movements (reference_type, reference_id);
CREATE INDEX idx_stock_movements_created_at ON stock_movements (created_at);

COMMENT ON TABLE stock_movements IS 'Перемещения и списания товара со склада';
