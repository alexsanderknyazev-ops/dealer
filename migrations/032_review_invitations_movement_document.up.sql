-- Приглашения на отзыв после реализации товара
SET search_path TO reviews, public;

ALTER TABLE review_invitations DROP CONSTRAINT IF EXISTS review_invitations_source_type_check;
ALTER TABLE review_invitations ADD CONSTRAINT review_invitations_source_type_check
    CHECK (source_type IN ('work_order', 'deal', 'movement_document'));

ALTER TABLE review_invitations DROP CONSTRAINT IF EXISTS review_invitations_service_kind_check;
ALTER TABLE review_invitations ADD CONSTRAINT review_invitations_service_kind_check
    CHECK (service_kind IN ('service', 'sale', 'parts'));

COMMENT ON COLUMN review_invitations.source_type IS 'work_order, deal, movement_document — реализация товара';
COMMENT ON COLUMN review_invitations.service_kind IS 'service — СТО, sale — авто, parts — запчасти';
