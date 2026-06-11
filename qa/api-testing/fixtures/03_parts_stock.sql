-- QA fixtures: parts + part_stock
-- Requires: seed_dealer_brands (warehouse parts Moscow)

SET search_path TO parts, public;

INSERT INTO parts (
  id, sku, name, category, quantity, unit, price, location, notes,
  folder_id, brand_id, dealer_point_id, legal_entity_id, warehouse_id,
  created_at, updated_at
)
VALUES
  (
    'a4400001-0000-4000-8000-000000000001',
    'QA-PART-001',
    'QA Oil Filter',
    'Фильтры',
    50,
    'шт',
    900,
    'QA shelf A1',
    'Main fixture part — sufficient stock',
    '50000000-0000-4000-8000-000000000001',
    '40000000-0000-4000-8000-000000000003',
    '10000000-0000-4000-8000-000000000001',
    '20000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000002',
    now(), now()
  ),
  (
    'a4400001-0000-4000-8000-000000000002',
    'QA-PART-LOW',
    'QA Low Stock Gasket',
    'Расходники',
    5,
    'шт',
    350,
    'QA shelf B2',
    'For insufficient stock test (confirm qty > 5)',
    '50000000-0000-4000-8000-000000000004',
    '40000000-0000-4000-8000-000000000001',
    '10000000-0000-4000-8000-000000000001',
    '20000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000002',
    now(), now()
  )
ON CONFLICT (sku) DO UPDATE SET
  name        = EXCLUDED.name,
  price       = EXCLUDED.price,
  warehouse_id = EXCLUDED.warehouse_id,
  updated_at  = now();

-- Sync part_stock (trigger recalculates parts.quantity)
INSERT INTO part_stock (part_id, warehouse_id, quantity, created_at, updated_at)
VALUES
  ('a4400001-0000-4000-8000-000000000001', '30000000-0000-4000-8000-000000000002', 50, now(), now()),
  ('a4400001-0000-4000-8000-000000000002', '30000000-0000-4000-8000-000000000002', 5, now(), now())
ON CONFLICT (part_id, warehouse_id) DO UPDATE SET
  quantity   = EXCLUDED.quantity,
  updated_at = now();

-- Ensure parts.quantity matches stock sum
UPDATE parts p SET quantity = COALESCE(
  (SELECT SUM(ps.quantity) FROM part_stock ps WHERE ps.part_id = p.id), 0
), updated_at = now()
WHERE p.id IN (
  'a4400001-0000-4000-8000-000000000001',
  'a4400001-0000-4000-8000-000000000002'
);
