# integration — детальные E2E сценарии (API + БД + Kafka + Redis)

Полные цепочки с проверкой всех слоёв. Выполнять **последовательно**, сохранять ID в переменные.

---

## E2E-001 — Продажа автомобиля (employee contour)

| # | Действие | API | Проверка БД |
|---|----------|-----|-------------|
| 1 | Register employee | POST /api/register | auth.users +1, Redis refresh |
| 2 | Create customer | POST /api/customers | customers.customers +1 |
| 3 | Get brand | GET /api/brands | — |
| 4 | Create vehicle | POST /api/vehicles | vehicles.vehicles +1, vin unique |
| 5 | Create deal draft | POST /api/deals | deals.deals stage=draft |
| 6 | Complete deal | PUT stage=completed | deals.stage=completed |
| 7 | Wait 15s | — | employee_statistics.deal_events +1 |
| 8 | Stats | GET /api/stats/employee/overview | matches SQL aggregates |

**Rollback (optional):** DELETE deal, vehicle, customer — зафиксировать orphan behavior.

---

## E2E-002 — Клиент: регистрация → профиль → отзыв → видимость у дилера

| # | Действие | API | Проверка БД / async |
|---|----------|-----|---------------------|
| 1 | Employee creates vehicle | POST /api/vehicles | vin=VIN_E2E |
| 2 | Client register | POST :8091 /api/client/register | clients +1, client_vehicles +1 |
| 3 | Wait / verify | — | clientauth.users +1 (Kafka) |
| 4 | Profile | GET :8093 /api/client/profile | fields match clients |
| 5 | Create review | POST /api/client/reviews | reviews.reviews +1 |
| 6 | Wait 15s | — | employee_reviews.reviews +1 |
| 7 | Employee list | GET /api/reviews | contains review_id |
| 8 | Stats | GET /api/stats/client/overview | registration + review events |

**Redis:** client refresh exists after step 2.

---

## E2E-003 — Заказ-наряд: works + employees + выдача запчастей (admin/master)

| # | Действие | API | Проверка БД |
|---|----------|-----|-------------|
| 1 | Setup refs | fixtures + dealer-point, warehouse, customer, vehicle, part+stock(10) | all tables |
| 2 | Verify works | GET /api/works | works.works LAB-QA-* |
| 3 | Verify employees | GET /api/employees | employees for qa.master |
| 4 | Create WO | POST /api/work-orders с work_id, executor_id, service_advisor_id | work_orders + labor.work_id + parts |
| 5 | Move to work | POST .../move-parts-to-work | movement_documents draft |
| 6 | Confirm doc | POST .../movement-documents/{id}/confirm | stock −qty, doc confirmed |
| 7 | Verify WO | GET /api/work-orders/{id} | parts_issued=true, service_advisor_name set |
| 8 | Cross join | SQL wo + md + ps + works | all consistent |

**Negative branch:** confirm with qty=999 → draft unchanged, stock unchanged.

---

## E2E-004 — Auth isolation (security)

| # | Test | Expected |
|---|------|----------|
| 1 | Employee token → :8093 /api/client/profile | 401/403 |
| 2 | Client token → :8090 /api/customers | 401/403 |
| 3 | No token → both | 401 |

БД: COUNT(*) unchanged after failed writes.

---

## E2E-005 — Token lifecycle (both contours)

Employee + Client parallel:

1. login → Redis key exists  
2. refresh → old key deleted, new exists  
3. logout → key deleted  
4. refresh with old → 401  

---

## E2E-006 — Telemetry pipeline (optional)

1. POST :8080 /api/telemetry/events (js_error)  
2. errors-ingest → ClickHouse row (if stack up)  
3. Kafka platform.errors.v1 message  

```sql
-- ClickHouse via HTTP
-- SELECT count() FROM analytics.events WHERE message LIKE '%qa%'
```

---

## Чеклист перед прогоном E2E-003

- [ ] `./qa/api-testing/fixtures/apply-fixtures.sh` (включая `08_works_employees.sql`)
- [ ] Login `qa.master@test.local` / `Test1234!` (или admin)
- [ ] workorders-service, works-service, employees-service running
- [ ] `curl :8097/healthz`, `:8098/healthz`, `:8099/healthz` → 200
- [ ] parts WORKORDERS_GRPC_ADDR configured in docker-compose
- [ ] Stock quantity ≥ quantity in WO parts line
