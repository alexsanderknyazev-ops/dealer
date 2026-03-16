-- Папки для запасных частей (иерархия)
CREATE TABLE IF NOT EXISTS part_folders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL DEFAULT '',
    parent_id   UUID REFERENCES part_folders(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_part_folders_parent ON part_folders (parent_id);
CREATE INDEX idx_part_folders_name ON part_folders (name);

COMMENT ON TABLE part_folders IS 'Папки/категории для группировки запасных частей';

-- Связь частей с папкой
ALTER TABLE parts ADD COLUMN IF NOT EXISTS folder_id UUID REFERENCES part_folders(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_parts_folder_id ON parts (folder_id);
