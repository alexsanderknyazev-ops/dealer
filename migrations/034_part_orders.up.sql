-- Заказы поставщику и покупателя
SET search_path TO parts, public;

CREATE SEQUENCE IF NOT EXISTS supplier_orders_number_seq START 1;
CREATE SEQUENCE IF NOT EXISTS customer_orders_number_seq START 1;

CREATE TABLE IF NOT EXISTS supplier_orders (
    id                            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number                  TEXT NOT NULL UNIQUE,
    status                        TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'linked', 'fulfilled', 'cancelled')),
    supplier_id                   UUID NOT NULL REFERENCES suppliers(id),
    receipt_warehouse_id          UUID NOT NULL,
    fulfillment_movement_document_id UUID NULL REFERENCES movement_documents(id),
    notes                         TEXT NOT NULL DEFAULT '',
    created_by                    UUID NULL,
    created_at                    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS supplier_order_lines (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES supplier_orders(id) ON DELETE CASCADE,
    part_id     UUID NOT NULL REFERENCES parts(id),
    quantity    INT NOT NULL CHECK (quantity > 0),
    unit_price  NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    notes       TEXT NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS customer_orders (
    id                            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number                  TEXT NOT NULL UNIQUE,
    status                        TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'linked', 'fulfilled', 'cancelled')),
    customer_id                   UUID NOT NULL,
    vehicle_id                    UUID NULL,
    issue_warehouse_id            UUID NOT NULL,
    fulfillment_movement_document_id UUID NULL REFERENCES movement_documents(id),
    notes                         TEXT NOT NULL DEFAULT '',
    created_by                    UUID NULL,
    created_at                    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS customer_order_lines (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES customer_orders(id) ON DELETE CASCADE,
    part_id     UUID NOT NULL REFERENCES parts(id),
    quantity    INT NOT NULL CHECK (quantity > 0),
    unit_price  NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    notes       TEXT NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_supplier_orders_status ON supplier_orders (status);
CREATE INDEX IF NOT EXISTS idx_supplier_orders_supplier_id ON supplier_orders (supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_order_lines_order_id ON supplier_order_lines (order_id);
CREATE INDEX IF NOT EXISTS idx_customer_orders_status ON customer_orders (status);
CREATE INDEX IF NOT EXISTS idx_customer_orders_customer_id ON customer_orders (customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_order_lines_order_id ON customer_order_lines (order_id);

COMMENT ON TABLE supplier_orders IS 'Заказ поставщику — основание для поступления товара';
COMMENT ON TABLE customer_orders IS 'Заказ покупателя — основание для реализации товара';
