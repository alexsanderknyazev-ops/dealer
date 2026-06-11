# employee-statistics + client-statistics — детальные тест-кейсы

---

## Employee statistics

**Endpoint:** `GET /api/stats/employee/overview`  
**Схема:** `employee_statistics.deal_events`

### TC-EST-D001 — Overview baseline (P0)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/stats/employee/overview"
```

```sql
SELECT COUNT(*) AS deals, COALESCE(SUM(amount),0) AS revenue
FROM employee_statistics.deal_events WHERE stage = 'completed';
```

Сверить поля overview с агрегатами БД.

---

### TC-EST-D002 — Deal completed increments stats (P0)

1. Snapshot:
```sql
SELECT COUNT(*) AS n FROM employee_statistics.deal_events;
```

2. Complete deal (см. deals TC-DEL-D004)

3. Wait ≤15s

4. Verify:
```sql
SELECT * FROM employee_statistics.deal_events WHERE deal_id = '<DEAL_ID>';
-- amount, customer_id, vehicle_id match deal
```

5. GET overview → метрики изменились vs snapshot

---

### TC-EST-D003 — Idempotent deal event (P2)

Повторный complete → `deal_events` still COUNT=1 per deal_id (UNIQUE deal_id)

---

## Client statistics

**Endpoint:** `GET /api/stats/client/overview`  
**Схемы:** `client_statistics.client_registration_events`, `review_events`

### TC-CST-D001 — Overview (P0)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/stats/client/overview"
```

---

### TC-CST-D002 — Registration event (P0)

После client register:

```sql
SELECT user_id, email, vehicle_id, occurred_at
FROM client_statistics.client_registration_events WHERE user_id = '<USER_ID>';
```

Overview registrations +1

---

### TC-CST-D003 — Review event (P0)

После client create review:

```sql
SELECT review_id, rating, status FROM client_statistics.review_events
WHERE review_id = '<REVIEW_ID>';
```

Overview reviews +1

---

### TC-CST-D004 — No auth (P1)

GET overview без JWT → **401**
