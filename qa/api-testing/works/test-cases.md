# works — тест-кейсы

**Сервис:** `works-service`  
**REST:** `/api/works`  
**gRPC:** `:50065`  
**HTTP health:** `:8098/healthz`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-WRK-001 | P0 | GET /api/works | Bearer | limit=20 | 200, works[], total | WRK-001 |
| TC-WRK-002 | P0 | GET /api/works/{id} | Bearer | LAB-QA-004 id | 200, work.code | WRK-002 |
| TC-WRK-003 | P0 | POST /api/works | Bearer master/admin | code, name, category, labor_hours, unit_price | 200, work | manual |
| TC-WRK-004 | P1 | POST /api/works | Bearer sales | — | **403** | WRK-004 |
| TC-WRK-005 | P1 | PUT /api/works/{id} | Bearer master | name, unit_price | 200 | manual |
| TC-WRK-006 | P1 | DELETE /api/works/{id} | Bearer admin | QA-created work | 200 | manual |
| TC-WRK-007 | P1 | GET /api/works | Bearer | search=диагностика | 200, filtered | manual |
| TC-WRK-008 | P1 | GET /api/works | Bearer | category=ТО | 200 | manual |
| TC-WRK-009 | P1 | GET /api/works/{missing} | Bearer | random uuid | 404 | manual |
| TC-WRK-010 | P1 | GET /healthz :8098 | — | service up | 200 | WRK-010 |
| TC-WRK-011 | P2 | GET /api/works | Bearer | works-service down | **503** gateway | manual |

## Write roles

admin, manager, master, service_advisor

## Связь с work-orders

- `work_order_labor.work_id` → `works.works.id`
- При создании WO можно передать только `work_id` — description/qty/price подставятся из справочника (ResolveWork gRPC)
