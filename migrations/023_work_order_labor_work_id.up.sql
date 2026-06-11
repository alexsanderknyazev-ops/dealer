SET search_path TO workorders, public;

ALTER TABLE work_order_labor
    ADD COLUMN IF NOT EXISTS work_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_work_order_labor_work_id ON work_order_labor (work_id);

COMMENT ON COLUMN work_order_labor.work_id IS 'Ссылка на справочник works.works';
