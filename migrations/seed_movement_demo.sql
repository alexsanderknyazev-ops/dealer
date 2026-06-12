-- Демо-данные: остатки на складах, заказ-наряд и документ перемещения в черновике
SET search_path TO parts, workorders, customers, vehicles, public;

-- Синхронизировать остатки по всем запчастям со складами
INSERT INTO part_stock (part_id, warehouse_id, quantity, updated_at)
SELECT p.id, p.warehouse_id, GREATEST(p.quantity, 10), now()
FROM parts p
WHERE p.warehouse_id IS NOT NULL
ON CONFLICT (part_id, warehouse_id) DO UPDATE
SET quantity = GREATEST(part_stock.quantity, EXCLUDED.quantity), updated_at = now();

-- Исправить статус документа у QA заказ-наряда после миграции 026
UPDATE work_orders
SET movement_document_status = 'closed'
WHERE order_number = 'WO-QA-0001' AND movement_document_status = 'confirmed';

-- Заказ-наряд с запчастями (черновик) для демо перемещения
INSERT INTO work_orders (
    id, order_number, customer_id, vehicle_id, dealer_point_id, warehouse_id,
    repair_type, status, complaint, mileage_km, opened_at, created_at, updated_at
)
SELECT
    'a6600002-0000-4000-8000-000000000001',
    'WO-DEMO-0001',
    c.id,
    v.id,
    '10000000-0000-4000-8000-000000000001'::uuid,
    '30000000-0000-4000-8000-000000000002'::uuid,
    'commercial',
    'draft',
    'Замена масла и фильтров',
    45000,
    now(),
    now(),
    now()
FROM customers c
JOIN vehicles v ON v.make = 'Volkswagen'
WHERE c.email = 'ivan.petrov@example.com'
  AND NOT EXISTS (SELECT 1 FROM work_orders WHERE order_number = 'WO-DEMO-0001')
LIMIT 1;

INSERT INTO work_order_parts (
    id, work_order_id, part_id, warehouse_id, description, quantity, unit_price, amount, sort_order, created_at, updated_at
)
SELECT
    gen_random_uuid(),
    'a6600002-0000-4000-8000-000000000001'::uuid,
    p.id,
    p.warehouse_id,
    p.name,
    2,
    p.price,
    p.price * 2,
    row_number() OVER (ORDER BY p.sku),
    now(),
    now()
FROM parts p
WHERE p.sku IN ('FLT-OIL-001', 'OIL-5W30-001')
  AND EXISTS (SELECT 1 FROM work_orders WHERE id = 'a6600002-0000-4000-8000-000000000001')
  AND NOT EXISTS (
    SELECT 1 FROM work_order_parts wop
    WHERE wop.work_order_id = 'a6600002-0000-4000-8000-000000000001'
  );

-- Документ перемещения в черновике (не из заказ-наряда)
INSERT INTO movement_documents (
    id, document_number, status, movement_type, reference_type, reference_id, notes, created_at, updated_at
)
SELECT
    'b7700001-0000-4000-8000-000000000001',
    'MOV-DEMO-0001',
    'draft',
    'transfer',
    '',
    NULL,
    'Демо: перемещение между складами',
    now(),
    now()
WHERE NOT EXISTS (SELECT 1 FROM movement_documents WHERE document_number = 'MOV-DEMO-0001');

INSERT INTO movement_document_lines (
    id, document_id, part_id, warehouse_id, quantity, notes, sort_order, created_at
)
SELECT
    gen_random_uuid(),
    'b7700001-0000-4000-8000-000000000001'::uuid,
    p.id,
    p.warehouse_id,
    3,
    p.name,
    row_number() OVER (ORDER BY p.sku),
    now()
FROM parts p
WHERE p.sku IN ('BRK-PAD-F-001', 'WIP-BLADE-001')
  AND EXISTS (SELECT 1 FROM movement_documents WHERE id = 'b7700001-0000-4000-8000-000000000001')
  AND NOT EXISTS (
    SELECT 1 FROM movement_document_lines WHERE document_id = 'b7700001-0000-4000-8000-000000000001'
  );

-- Документ в работе (для теста закрытия и списания)
INSERT INTO movement_documents (
    id, document_number, status, movement_type, reference_type, notes, created_at, updated_at
)
SELECT
    'b7700002-0000-4000-8000-000000000001',
    'MOV-DEMO-0002',
    'in_progress',
    'work_order_issue',
    'work_order',
    'Демо: ожидает закрытия и списания',
    now(),
    now()
WHERE NOT EXISTS (SELECT 1 FROM movement_documents WHERE document_number = 'MOV-DEMO-0002');

INSERT INTO movement_document_lines (
    id, document_id, part_id, warehouse_id, quantity, notes, sort_order, created_at
)
SELECT
    gen_random_uuid(),
    'b7700002-0000-4000-8000-000000000001'::uuid,
    p.id,
    '30000000-0000-4000-8000-000000000002'::uuid,
    1,
    p.name,
    1,
    now()
FROM parts p
WHERE p.sku = 'FLT-OIL-001'
  AND EXISTS (SELECT 1 FROM movement_documents WHERE id = 'b7700002-0000-4000-8000-000000000001')
  AND NOT EXISTS (
    SELECT 1 FROM movement_document_lines WHERE document_id = 'b7700002-0000-4000-8000-000000000001'
  );
