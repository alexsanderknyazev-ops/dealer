-- Запись клиентов на ремонт
CREATE SCHEMA IF NOT EXISTS appointments;

SET search_path TO appointments, public;

CREATE SEQUENCE IF NOT EXISTS repair_appointments_number_seq START 1;

CREATE TABLE IF NOT EXISTS repair_appointments (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_number TEXT NOT NULL UNIQUE,
    customer_id        UUID NOT NULL,
    vehicle_id         UUID NOT NULL,
    dealer_point_id    UUID NULL,
    warehouse_id       UUID NULL,
    scheduled_start    TIMESTAMPTZ NOT NULL,
    scheduled_end      TIMESTAMPTZ NOT NULL,
    status             TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN (
        'draft', 'scheduled', 'in_progress', 'completed', 'cancelled'
    )),
    work_order_id      UUID NULL,
    complaint          TEXT NOT NULL DEFAULT '',
    notes              TEXT NOT NULL DEFAULT '',
    created_by         UUID NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (scheduled_end > scheduled_start)
);

CREATE TABLE IF NOT EXISTS repair_appointment_labor (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id  UUID NOT NULL REFERENCES repair_appointments(id) ON DELETE CASCADE,
    work_id         UUID NULL,
    description     TEXT NOT NULL DEFAULT '',
    quantity        NUMERIC(10,3) NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price      NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS repair_appointment_parts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id  UUID NOT NULL REFERENCES repair_appointments(id) ON DELETE CASCADE,
    part_id         UUID NOT NULL,
    warehouse_id    UUID NOT NULL,
    quantity        INT NOT NULL CHECK (quantity > 0),
    unit_price      NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    notes           TEXT NOT NULL DEFAULT '',
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_repair_appointments_scheduled_start ON repair_appointments (scheduled_start);
CREATE INDEX IF NOT EXISTS idx_repair_appointments_status ON repair_appointments (status);
CREATE INDEX IF NOT EXISTS idx_repair_appointments_customer_id ON repair_appointments (customer_id);
CREATE INDEX IF NOT EXISTS idx_repair_appointment_labor_appointment_id ON repair_appointment_labor (appointment_id);
CREATE INDEX IF NOT EXISTS idx_repair_appointment_parts_appointment_id ON repair_appointment_parts (appointment_id);

COMMENT ON TABLE repair_appointments IS 'Запись клиента на ремонт (слот в расписании СТО)';
COMMENT ON COLUMN repair_appointments.work_order_id IS 'Связанный заказ-наряд после открытия ремонта';
