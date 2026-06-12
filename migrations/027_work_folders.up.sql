-- Папки для справочника работ (иерархия)
SET search_path TO works, public;

CREATE TABLE IF NOT EXISTS work_folders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL DEFAULT '',
    parent_id   UUID REFERENCES work_folders(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_work_folders_parent ON work_folders (parent_id);
CREATE INDEX IF NOT EXISTS idx_work_folders_name ON work_folders (name);

COMMENT ON TABLE work_folders IS 'Папки для группировки работ СТО';

ALTER TABLE works ADD COLUMN IF NOT EXISTS folder_id UUID REFERENCES work_folders(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_works_folder_id ON works (folder_id);
