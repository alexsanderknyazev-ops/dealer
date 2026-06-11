-- QA fixtures: draft work order + lines
-- Requires: 01–03 fixtures, seed_dealer_brands

SET search_path TO workorders, public;

INSERT INTO work_orders (
  id, order_number, customer_id, vehicle_id, dealer_point_id, warehouse_id,
  repair_type, status, service_advisor_id,
  complaint, diagnosis, mileage_km,
  labor_cost, parts_cost, total_cost,
  parts_issued, notes,
  movement_document_id, movement_document_status,
  created_at, updated_at
)
VALUES (
  'a6600001-0000-4000-8000-000000000001',
  'WO-QA-0001',
  'a2200001-0000-4000-8000-000000000001',
  'a3300001-0000-4000-8000-000000000001',
  '10000000-0000-4000-8000-000000000001',
  '30000000-0000-4000-8000-000000000002',
  'commercial',
  'draft',
  'a9900001-0000-4000-8000-000000000003',
  'Стук при торможении',
  '',
  42000,
  2000,
  1800,
  3800,
  false,
  'QA fixture WO — move-parts-to-work / confirm tests',
  NULL,
  '',
  now(), now()
)
ON CONFLICT (id) DO UPDATE SET
  status     = EXCLUDED.status,
  complaint  = EXCLUDED.complaint,
  parts_issued = false,
  movement_document_id = NULL,
  movement_document_status = '',
  updated_at = now();

-- Labor line (work_id → works catalog, executor_id → auth user / employees)
INSERT INTO work_order_labor (
  id, work_order_id, work_id, description, quantity, unit_price, amount, executor_id, sort_order, created_at, updated_at
)
VALUES (
  'a6600001-0000-4000-8000-000000000012',
  'a6600001-0000-4000-8000-000000000001',
  'a8800001-0000-4000-8000-000000000001',
  'Диагностика тормозов',
  1,
  2000,
  2000,
  'a9900001-0000-4000-8000-000000000003',
  1,
  now(), now()
)
ON CONFLICT (id) DO UPDATE SET
  work_id     = EXCLUDED.work_id,
  description = EXCLUDED.description,
  amount      = EXCLUDED.amount,
  executor_id = EXCLUDED.executor_id,
  updated_at  = now();

-- Part line (not yet issued)
INSERT INTO work_order_parts (
  id, work_order_id, part_id, warehouse_id, description, quantity, unit_price, amount, issued, sort_order, created_at, updated_at
)
VALUES (
  'a6600001-0000-4000-8000-000000000011',
  'a6600001-0000-4000-8000-000000000001',
  'a4400001-0000-4000-8000-000000000001',
  '30000000-0000-4000-8000-000000000002',
  'QA Oil Filter',
  2,
  900,
  1800,
  false,
  1,
  now(), now()
)
ON CONFLICT (id) DO UPDATE SET
  quantity   = EXCLUDED.quantity,
  issued     = false,
  updated_at = now();
