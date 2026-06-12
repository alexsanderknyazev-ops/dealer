-- Связь заказов с заказ-нарядами
SET search_path TO parts, public;

ALTER TABLE supplier_orders
    ADD COLUMN IF NOT EXISTS fulfillment_work_order_id UUID NULL;

ALTER TABLE customer_orders
    ADD COLUMN IF NOT EXISTS fulfillment_work_order_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_supplier_orders_fulfillment_wo ON supplier_orders (fulfillment_work_order_id)
    WHERE fulfillment_work_order_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customer_orders_fulfillment_wo ON customer_orders (fulfillment_work_order_id)
    WHERE fulfillment_work_order_id IS NOT NULL;

SET search_path TO workorders, public;

ALTER TABLE work_orders
    ADD COLUMN IF NOT EXISTS source_order_type TEXT NULL
        CHECK (source_order_type IS NULL OR source_order_type IN ('supplier_order', 'customer_order')),
    ADD COLUMN IF NOT EXISTS source_order_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_work_orders_source_order ON work_orders (source_order_type, source_order_id)
    WHERE source_order_id IS NOT NULL;

COMMENT ON COLUMN work_orders.source_order_type IS 'Тип исходного заказа запчастей (supplier_order / customer_order)';
COMMENT ON COLUMN work_orders.source_order_id IS 'ID исходного заказа запчастей';
