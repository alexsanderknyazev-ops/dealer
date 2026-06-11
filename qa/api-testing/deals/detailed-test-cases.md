# deals — детальные тест-кейсы (API + БД + Kafka)

**REST:** `/api/deals`  
**Схема:** `deals.deals`  
**Kafka:** `deal.completed.v1` → `employee_statistics.deal_events`

---

## TC-DEL-D001 — Create deal draft (P0)

### Preconditions
- CUSTOMER_ID, VEHICLE_ID существуют в БД

### До
```sql
SELECT COUNT(*) FROM deals.deals WHERE customer_id = '<CUSTOMER_ID>';
```

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/deals" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "customer_id": "<CUSTOMER_ID>",
    "vehicle_id": "<VEHICLE_ID>",
    "amount": "2500000",
    "stage": "draft",
    "responsible_id": "<USER_ID>",
    "notes": "QA deal"
  }'
```

Сохранить: `DEAL_ID`

### БД
```sql
SELECT id, customer_id, vehicle_id, amount, stage, assigned_to, notes
FROM deals.deals WHERE id = '<DEAL_ID>';
```
- stage = 'draft'
- assigned_to = USER_ID (responsible_id из API)

---

## TC-DEL-D002 — Invalid customer_id (P1)

### API
- customer_id = random uuid → **404/400**

### БД
```sql
SELECT COUNT(*) FROM deals.deals WHERE customer_id = '00000000-0000-4000-8000-000000000099';
-- 0
```

---

## TC-DEL-D003 — Invalid vehicle_id (P1)

Аналогично TC-DEL-D002 для vehicle_id.

---

## TC-DEL-D004 — Stage transition to completed (P0)

### API
```bash
curl -s -X PUT "$EMPLOYEE_API/api/deals/$DEAL_ID" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"stage": "completed", "amount": "2500000"}'
```

### БД (deals)
```sql
SELECT stage, updated_at FROM deals.deals WHERE id = '<DEAL_ID>';
-- stage = 'completed'
```

### БД (statistics) — подождать ≤15 с
```sql
SELECT deal_id, customer_id, vehicle_id, amount, stage, occurred_at
FROM employee_statistics.deal_events WHERE deal_id = '<DEAL_ID>';
```
- **1 row**, stage = 'completed'

### Kafka (optional)
- Topic `deal.completed.v1`, key = DEAL_ID

---

## TC-DEL-D005 — Completed не дублируется (P2)

Повторный PUT completed:

### БД
```sql
SELECT COUNT(*) FROM employee_statistics.deal_events WHERE deal_id = '<DEAL_ID>';
-- ожидание: 1 (idempotent consumer / unique deal_id)
```

---

## TC-DEL-D006 — List + Get (P1)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" "$EMPLOYEE_API/api/deals/$DEAL_ID"
```

Cross-check все поля с `deals.deals`.

---

## TC-DEL-D007 — Delete deal (P1)

### БД
- Row удалена из `deals.deals`
- `employee_statistics.deal_events` — row **остаётся** (историческая аналитика)

---

## TC-DEL-D008 — RBAC sales write (P1)

User role=sales:
- POST /api/deals → **200** (sales в WriteRoles)
- БД: row created

User role=viewer (если есть):
- POST → **403**, COUNT без изменений
