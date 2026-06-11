# client-statistics — детальные тест-кейсы

> Полные сценарии с SQL: см. также [`../employee-statistics/detailed-test-cases.md`](../employee-statistics/detailed-test-cases.md) секция **Client statistics**.

**Endpoint:** `GET /api/stats/client/overview` (employee JWT)  
**Схема:** `client_statistics.client_registration_events`, `client_statistics.review_events`

---

## TC-CST-D001 — Overview vs DB aggregates

```sql
SELECT
  (SELECT COUNT(*) FROM client_statistics.client_registration_events) AS registrations,
  (SELECT COUNT(*) FROM client_statistics.review_events) AS reviews;
```

Сверить с JSON `overview` из API.

---

## TC-CST-D002 — After client register

См. E2E-002 step 8 в [`../integration/detailed-test-cases.md`](../integration/detailed-test-cases.md)

---

## TC-CST-D003 — After review create

```sql
SELECT review_id, rating FROM client_statistics.review_events
WHERE review_id = '<REVIEW_ID>';
```

Unique on `review_id` — повторный Kafka event не дублирует (если consumer idempotent).
