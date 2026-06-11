CREATE SCHEMA IF NOT EXISTS workorders;

SET search_path TO workorders, public;

CREATE SEQUENCE IF NOT EXISTS work_orders_number_seq START 1;

CREATE TABLE IF NOT EXISTS work_orders (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number       TEXT NOT NULL UNIQUE,
    customer_id        UUID NOT NULL,
    vehicle_id         UUID NOT NULL,
    dealer_point_id    UUID NULL,
    warehouse_id       UUID NULL,
    repair_type        TEXT NOT NULL DEFAULT 'commercial' CHECK (repair_type IN (
        'warranty_manufacturer', 'pre_sale', 'commercial', 'maintenance'
    )),
    status             TEXT NOT NULL DEFAULT 'draft' CHECK (status IN (
        'draft', 'in_progress', 'completed', 'closed', 'paid'
    )),
    service_advisor_id UUID NULL,
    complaint          TEXT NOT NULL DEFAULT '',
    diagnosis          TEXT NOT NULL DEFAULT '',
    mileage_km         BIGINT NOT NULL DEFAULT 0 CHECK (mileage_km >= 0),
    labor_cost         NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (labor_cost >= 0),
    parts_cost         NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (parts_cost >= 0),
    total_cost         NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (total_cost >= 0),
    opened_at          TIMESTAMPTZ NULL,
    closed_at          TIMESTAMPTZ NULL,
    parts_issued       BOOLEAN NOT NULL DEFAULT false,
    parts_issued_at    TIMESTAMPTZ NULL,
    notes              TEXT NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS work_order_labor (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    work_order_id   UUID NOT NULL REFERENCES work_orders(id) ON DELETE CASCADE,
    description     TEXT NOT NULL DEFAULT '',
    quantity        NUMERIC(10,3) NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price      NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    amount          NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    executor_id     UUID NULL,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS work_order_parts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    work_order_id   UUID NOT NULL REFERENCES work_orders(id) ON DELETE CASCADE,
    part_id         UUID NOT NULL,
    warehouse_id    UUID NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    quantity        NUMERIC(10,3) NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price      NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    amount          NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    issued          BOOLEAN NOT NULL DEFAULT false,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_work_orders_customer_id ON work_orders (customer_id);
CREATE INDEX idx_work_orders_vehicle_id ON work_orders (vehicle_id);
CREATE INDEX idx_work_orders_dealer_point_id ON work_orders (dealer_point_id);
CREATE INDEX idx_work_orders_status ON work_orders (status);
CREATE INDEX idx_work_orders_repair_type ON work_orders (repair_type);
CREATE INDEX idx_work_orders_order_number ON work_orders (order_number);
CREATE INDEX idx_work_orders_opened_at ON work_orders (opened_at);
CREATE INDEX idx_work_order_labor_work_order_id ON work_order_labor (work_order_id);
CREATE INDEX idx_work_order_parts_work_order_id ON work_order_parts (work_order_id);
CREATE INDEX idx_work_order_parts_part_id ON work_order_parts (part_id);

COMMENT ON TABLE work_orders IS 'Заказ-наряды СТО';
COMMENT ON COLUMN work_orders.repair_type IS 'warranty_manufacturer, pre_sale, commercial, maintenance';
COMMENT ON COLUMN work_orders.service_advisor_id IS 'Мастер-консультант';
COMMENT ON TABLE work_order_labor IS 'Работы заказ-наряда с исполнителем';
COMMENT ON TABLE work_order_parts IS 'Запчасти заказ-наряда';
