-- Уведомления клиентам и связь заказа поставщику с заказом покупателя
SET search_path TO parts, public;

ALTER TABLE supplier_orders
    ADD COLUMN IF NOT EXISTS customer_order_id UUID NULL REFERENCES customer_orders(id);

CREATE INDEX IF NOT EXISTS idx_supplier_orders_customer_order_id ON supplier_orders (customer_order_id)
    WHERE customer_order_id IS NOT NULL;

COMMENT ON COLUMN supplier_orders.customer_order_id IS 'Заказ покупателя, для которого закупаются запчасти';

SET search_path TO clients, public;

CREATE TABLE IF NOT EXISTS client_notifications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id    UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL,
    kind         TEXT NOT NULL,
    source_type  TEXT NOT NULL,
    source_id    UUID NOT NULL,
    title        TEXT NOT NULL,
    body         TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'unread'
        CHECK (status IN ('unread', 'dismissed')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, source_type, source_id)
);

CREATE INDEX IF NOT EXISTS idx_client_notifications_user_status ON client_notifications (user_id, status);
CREATE INDEX IF NOT EXISTS idx_client_notifications_client_id ON client_notifications (client_id);

COMMENT ON TABLE client_notifications IS 'Уведомления зарегистрированным клиентам (личный кабинет)';
COMMENT ON COLUMN client_notifications.kind IS 'customer_order_receipt — поступление запчастей по заказу покупателя';
