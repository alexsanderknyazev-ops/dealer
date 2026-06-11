# work-orders — детальные тест-кейсы (API + БД + parts gRPC)

**REST:** `/api/work-orders`  
**Схема:** `workorders.*`  
**Precondition:** ADMIN/MASTER token, refs: customer, vehicle, dealer_point, warehouse, part+stock, **work** (works catalog), **employee** (advisor/executor)

---

## TC-WO-D001 — Create work order (P0)

### До
```sql
SELECT COUNT(*) FROM workorders.work_orders;
```

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/work-orders" \
  -H "Authorization: Bearer $MASTER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "customer_id": "<CUSTOMER_ID>",
    "vehicle_id": "<VEHICLE_ID>",
    "dealer_point_id": "<DEALER_POINT_ID>",
    "warehouse_id": "<WAREHOUSE_ID>",
    "repair_type": "commercial",
    "status": "draft",
    "service_advisor_id": "<EMPLOYEE_MASTER_ID>",
    "complaint": "Стук подвески",
    "mileage_km": 45000,
    "labor": [{
      "work_id": "<WORK_ID>",
      "description": "Диагностика",
      "quantity": "1",
      "unit_price": "2000",
      "executor_id": "<EMPLOYEE_MASTER_ID>",
      "sort_order": 1
    }],
    "parts": [{
      "part_id": "<PART_ID>",
      "warehouse_id": "<WAREHOUSE_ID>",
      "description": "Filter",
      "quantity": "2",
      "unit_price": "850",
      "sort_order": 1
    }]
  }'
```

Сохранить: `WO_ID`, `ORDER_NUMBER`

### HTTP
- `service_advisor_name` не пуст (если employees-service доступен)

### БД — header
```sql
SELECT id, order_number, customer_id, vehicle_id, status, service_advisor_id,
       labor_cost, parts_cost, total_cost, parts_issued, movement_document_id
FROM workorders.work_orders WHERE id = '<WO_ID>';
```
- status = 'draft'
- parts_issued = false
- movement_document_id IS NULL

### БД — labor lines
```sql
SELECT work_id, description, quantity, unit_price, amount, executor_id
FROM workorders.work_order_labor WHERE work_order_id = '<WO_ID>';
-- work_id NOT NULL, amount = quantity * unit_price
```

### БД — parts lines
```sql
SELECT part_id, warehouse_id, quantity, issued FROM workorders.work_order_parts
WHERE work_order_id = '<WO_ID>';
-- issued = false
```

---

## TC-WO-D001b — Labor auto-fill from works catalog (P0)

POST WO с labor `[{"work_id": "<WORK_ID>", "sort_order": 1}]` без description/qty/price.

### БД
```sql
SELECT wol.description, wol.quantity::text, wol.unit_price::text, w.code
FROM workorders.work_order_labor wol
JOIN works.works w ON w.id = wol.work_id
WHERE wol.work_order_id = '<WO_ID>';
-- values match works.labor_hours and works.unit_price
```

---

## TC-WO-D002 — Invalid refs (P1)

| Case | Field | Expected HTTP | БД delta |
|------|-------|---------------|----------|
| bad customer | customer_id random | 4xx | 0 |
| bad vehicle | vehicle_id random | 4xx | 0 |
| bad warehouse | warehouse_id random | 4xx | 0 |
| bad part | part_id random | 4xx | 0 |
| bad work | work_id random | 4xx work not found | 0 |
| bad employee | service_advisor_id random | 4xx employee not found | 0 |

Проверка gRPC ReferenceChecker через отсутствие новых rows в `workorders.work_orders`.

---

## TC-WO-D003 — MovePartsToWork (P0)

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/work-orders/$WO_ID/move-parts-to-work" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"issued_by": "<USER_ID>"}'
```

### HTTP
- `movement_document_id` в response work_order

Сохранить: `DOC_ID`

### БД — workorders
```sql
SELECT movement_document_id, movement_document_status, parts_issued
FROM workorders.work_orders WHERE id = '<WO_ID>';
-- movement_document_id = DOC_ID
-- movement_document_status = 'draft'
-- parts_issued may still false until confirm
```

### БД — parts (created via gRPC)
```sql
SELECT id, status, reference_type, reference_id, movement_type
FROM parts.movement_documents WHERE id = '<DOC_ID>';
-- status = 'draft'
-- reference_type = 'work_order'
-- reference_id = WO_ID
-- movement_type = 'work_order_issue'
```

### БД — lines match WO parts
```sql
SELECT mdl.part_id, mdl.quantity, wop.quantity
FROM parts.movement_document_lines mdl
JOIN workorders.work_order_parts wop ON wop.part_id = mdl.part_id::text
WHERE mdl.document_id = '<DOC_ID>' AND wop.work_order_id = '<WO_ID>';
-- quantities aligned
```

---

## TC-WO-D004 — Full E2E with confirm (P0)

1. TC-WO-D003 (move-parts-to-work)
2. TC-PRT-D004 (confirm document)
3. Verify cross-schema:

```sql
SELECT wo.parts_issued, wo.movement_document_status, md.status, ps.quantity
FROM workorders.work_orders wo
JOIN parts.movement_documents md ON md.id = wo.movement_document_id
JOIN parts.part_stock ps ON ps.part_id = '<PART_ID>' AND ps.warehouse_id = '<WAREHOUSE_ID>'
WHERE wo.id = '<WO_ID>';
```

---

## TC-WO-D005 — Update status (P1)

```bash
curl -s -X PUT "$EMPLOYEE_API/api/work-orders/$WO_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"status": "in_progress", "diagnosis": "Износ сайлентблоков"}'
```

### БД
```sql
SELECT status, diagnosis, opened_at IS NOT NULL FROM workorders.work_orders WHERE id = '<WO_ID>';
```

---

## TC-WO-D006 — Delete draft WO (P1)

### БД до
```sql
SELECT COUNT(*) FROM workorders.work_order_labor WHERE work_order_id = '<WO_ID>';
```

### API DELETE

### БД после
- CASCADE: labor + parts lines удалены
- movement document **не** удаляется автоматически — проверить orphan policy

---

## TC-WO-D007 — Sales role forbidden (P1)

POST /api/work-orders с SALES_TOKEN → **403**, COUNT без изменений

---

## TC-WO-D008 — Service unavailable (P2)

Если workorders-service down:

```bash
curl -s -w "\nHTTP:%{http_code}\n" -H "Authorization: Bearer $TOKEN" \
  "$EMPLOYEE_API/api/work-orders"
```
- **503** от gateway

```bash
curl -s -w "\nHTTP:%{http_code}\n" http://127.0.0.1:8097/healthz
```
