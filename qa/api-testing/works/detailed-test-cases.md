# works — детальные тест-кейсы (API + БД)

**REST:** `/api/works`  
**Схема:** `works.works`  
**Precondition:** Bearer token (любой employee для read; master/admin для write)

Фиксированные ID: `fixtures/manifest.json` → `works.diagnostic`, `works.oil_change`

---

## TC-WRK-D001 — List works (P0)

### API
```bash
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$EMPLOYEE_API/api/works?limit=10"
```

### HTTP
- `works[]` не пуст (после migrate + fixtures)
- `total >= 1`

### БД
```sql
SELECT COUNT(*) FROM works.works;
SELECT id, code, name, labor_hours::text, unit_price::text
FROM works.works WHERE code = 'LAB-QA-004';
```

---

## TC-WRK-D002 — Get work by id (P0)

```bash
WORK_ID="a8800001-0000-4000-8000-000000000001"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$EMPLOYEE_API/api/works/$WORK_ID"
```

### БД
```sql
SELECT code, name, category FROM works.works WHERE id = '$WORK_ID';
-- code = LAB-QA-004
```

---

## TC-WRK-D003 — Create work (P1, master token)

### До
```sql
SELECT COUNT(*) FROM works.works WHERE code = 'LAB-QA-NEW';
```

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/works" \
  -H "Authorization: Bearer $MASTER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "code": "LAB-QA-NEW",
    "name": "QA custom work",
    "category": "Тест",
    "labor_hours": "1.5",
    "unit_price": "3000",
    "notes": "created by QA"
  }'
```

Сохранить `NEW_WORK_ID` из response.

### БД
```sql
SELECT id, code, labor_hours, unit_price FROM works.works WHERE id = '<NEW_WORK_ID>';
-- labor_hours = 1.5, unit_price = 3000
```

---

## TC-WRK-D004 — Sales forbidden write (P1)

POST /api/works с `SALES_TOKEN` → **403**, COUNT без изменений.

---

## TC-WRK-D005 — Use work in work-order labor (P0 integration)

Создать WO с labor только по `work_id` (без description/qty/price):

```bash
curl -s -X POST "$EMPLOYEE_API/api/work-orders" \
  -H "Authorization: Bearer $MASTER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "customer_id": "<CUSTOMER_ID>",
    "vehicle_id": "<VEHICLE_ID>",
    "dealer_point_id": "<DEALER_POINT_ID>",
    "warehouse_id": "<WAREHOUSE_ID>",
    "labor": [{"work_id": "a8800001-0000-4000-8000-000000000002", "sort_order": 1}],
    "parts": []
  }'
```

### БД
```sql
SELECT wol.work_id, wol.description, wol.quantity::text, wol.unit_price::text, w.name
FROM workorders.work_order_labor wol
JOIN works.works w ON w.id = wol.work_id
WHERE wol.work_order_id = '<WO_ID>';
-- description/name from catalog LAB-QA-001
```

---

## TC-WRK-D006 — Service unavailable (P2)

works-service down → GET /api/works через gateway → **503**

```bash
curl -s -w "\nHTTP:%{http_code}\n" -H "Authorization: Bearer $TOKEN" \
  "$EMPLOYEE_API/api/works"
```
