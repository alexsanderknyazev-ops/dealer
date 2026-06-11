# parts — детальные тест-кейсы (API + БД + gRPC callback)

**REST:** `/api/parts`, `/api/parts/folders`, `/api/movement-documents`  
**Схемы:** `parts.*`  
**Write roles:** admin, manager, parts_manager, storekeeper, master, service_advisor  
**Precondition:** employee token с role **admin** или **master** (см. `_shared/test-users.md`)

---

## TC-PRT-D001 — RBAC: sales cannot create (P0)

### Token
- role=sales (default register)

### API
```bash
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$EMPLOYEE_API/api/parts" \
  -H "Authorization: Bearer $SALES_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"sku":"SKU-QA-001","name":"Filter","quantity":1,"unit":"pcs","price":"500"}'
```

### HTTP
- **403**

### БД
```sql
SELECT COUNT(*) FROM parts.parts WHERE sku = 'SKU-QA-001';
-- 0
```

---

## TC-PRT-D002 — Create part + stock (P0)

### Preconditions
- BRAND_ID, WAREHOUSE_ID (type=parts), DEALER_POINT_ID

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/parts" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "sku": "SKU-QA-002",
    "name": "Oil filter",
    "category": "filters",
    "brand_id": "<BRAND_ID>",
    "warehouse_id": "<WAREHOUSE_ID>",
    "dealer_point_id": "<DEALER_POINT_ID>",
    "quantity": 10,
    "unit": "pcs",
    "price": "850",
    "initial_stock": [{"warehouse_id": "<WAREHOUSE_ID>", "quantity": 10}]
  }'
```

Сохранить: `PART_ID`

### БД — parts
```sql
SELECT id, sku, name, quantity, brand_id, warehouse_id FROM parts.parts WHERE id = '<PART_ID>';
```

### БД — part_stock
```sql
SELECT part_id, warehouse_id, quantity FROM parts.part_stock
WHERE part_id = '<PART_ID>' AND warehouse_id = '<WAREHOUSE_ID>';
-- quantity = 10
```

### БД — trigger sync
```sql
SELECT quantity FROM parts.parts WHERE id = '<PART_ID>';
-- quantity = SUM(part_stock) = 10
```

---

## TC-PRT-D003 — Create movement document draft (P0)

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/movement-documents" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "movement_type": "work_order_issue",
    "reference_type": "work_order",
    "reference_id": "<WO_ID>",
    "notes": "QA issue",
    "lines": [{
      "part_id": "<PART_ID>",
      "warehouse_id": "<WAREHOUSE_ID>",
      "quantity": 2,
      "sort_order": 1
    }]
  }'
```

Сохранить: `DOC_ID`

### БД — document
```sql
SELECT id, document_number, status, movement_type, reference_type, reference_id
FROM parts.movement_documents WHERE id = '<DOC_ID>';
-- status = 'draft'
```

### БД — lines
```sql
SELECT part_id, warehouse_id, quantity FROM parts.movement_document_lines
WHERE document_id = '<DOC_ID>';
-- quantity = 2
```

### БД — stock НЕ изменился
```sql
SELECT quantity FROM parts.part_stock WHERE part_id = '<PART_ID>';
-- still 10
```

### БД — stock_movements
```sql
SELECT COUNT(*) FROM parts.stock_movements WHERE movement_document_id = '<DOC_ID>';
-- 0 (до confirm)
```

---

## TC-PRT-D004 — Confirm movement document (P0)

### До
```sql
SELECT quantity FROM parts.part_stock
WHERE part_id = '<PART_ID>' AND warehouse_id = '<WAREHOUSE_ID>';
-- qty_before = 10
```

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/movement-documents/$DOC_ID/confirm" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"confirmed_by": "<USER_ID>"}'
```

### БД — document status
```sql
SELECT status, confirmed_by, confirmed_at IS NOT NULL AS confirmed
FROM parts.movement_documents WHERE id = '<DOC_ID>';
-- status = 'confirmed', confirmed = true
```

### БД — stock deducted
```sql
SELECT quantity FROM parts.part_stock
WHERE part_id = '<PART_ID>' AND warehouse_id = '<WAREHOUSE_ID>';
-- qty_before - 2 = 8
```

### БД — stock_movements
```sql
SELECT quantity, movement_type, reference_type, reference_id, movement_document_id
FROM parts.stock_movements WHERE movement_document_id = '<DOC_ID>';
-- quantity = -2 (negative = outbound)
-- movement_type = 'work_order_issue'
```

### БД — work order (gRPC ApplyMovementDocument)
```sql
SELECT movement_document_id, movement_document_status, parts_issued, parts_issued_at IS NOT NULL
FROM workorders.work_orders WHERE id = '<WO_ID>';
-- movement_document_id = DOC_ID
-- movement_document_status = 'confirmed'
-- parts_issued = true
```

---

## TC-PRT-D005 — Confirm insufficient stock (P1)

Создать doc с quantity > stock:

### API
- confirm → **4xx** insufficient stock

### БД
```sql
SELECT status FROM parts.movement_documents WHERE id = '<DOC_ID>';
-- still 'draft'
SELECT quantity FROM parts.part_stock ...;
-- unchanged
```

---

## TC-PRT-D006 — Cancel draft document (P1)

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/movement-documents/$DOC_ID/cancel" \
  -H "Authorization: Bearer $ADMIN_TOKEN" -d '{}'
```

### БД
```sql
SELECT status FROM parts.movement_documents WHERE id = '<DOC_ID>';
-- 'cancelled'
```

### workorders (if reference_type=work_order)
```sql
SELECT movement_document_status FROM workorders.work_orders WHERE id = '<WO_ID>';
-- 'cancelled'
```

---

## TC-PRT-D007 — Folders CRUD (P2)

### Create folder
```bash
curl -s -X POST "$EMPLOYEE_API/api/parts/folders" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name": "Engine parts"}'
```

### БД
```sql
SELECT id, name, parent_id FROM parts.part_folders WHERE name = 'Engine parts';
```

---

## TC-PRT-D008 — List movement documents filter (P2)

```bash
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$EMPLOYEE_API/api/movement-documents?status=confirmed&reference_type=work_order&reference_id=$WO_ID"
```

Cross-check count с БД:
```sql
SELECT COUNT(*) FROM parts.movement_documents
WHERE status = 'confirmed' AND reference_type = 'work_order' AND reference_id = '<WO_ID>';
```
