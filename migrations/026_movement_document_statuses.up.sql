-- Жизненный цикл документа: draft → in_progress → closed (списание при closed)
SET search_path TO parts, public;

ALTER TABLE movement_documents DROP CONSTRAINT IF EXISTS movement_documents_status_check;

UPDATE movement_documents SET status = 'closed' WHERE status = 'confirmed';

ALTER TABLE movement_documents ADD CONSTRAINT movement_documents_status_check
    CHECK (status IN ('draft', 'in_progress', 'closed', 'cancelled'));

COMMENT ON COLUMN movement_documents.status IS 'draft — черновик, in_progress — в работе, closed — закрыт (остатки списаны)';
