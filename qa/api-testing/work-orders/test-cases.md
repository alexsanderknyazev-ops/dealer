# work-orders — тест-кейсы

**Сервис:** `workorders-service`  
**REST:** `/api/work-orders`  
**gRPC:** `:50064`  
**Зависимости:** customers, vehicles, dealer-points, parts, **works**, **employees**

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-WO-001 | P0 | GET /api/work-orders | Bearer | — | 200, work_orders[] | WO-001 |
| TC-WO-002 | P0 | POST /api/work-orders | Bearer WO role | refs + labor[].**work_id** + parts[] | 200, order_number, costs | manual |
| TC-WO-003 | P1 | POST /api/work-orders | Bearer sales | — | **403** | WO-003 |
| TC-WO-004 | P0 | GET /api/work-orders/{id} | Bearer | fixture WO-QA-0001 | 200, labor.work_id, service_advisor_name | manual |
| TC-WO-005 | P1 | PUT /api/work-orders/{id} | Bearer WO role | status, diagnosis, replace labor | 200, total_cost пересчёт | manual |
| TC-WO-006 | P1 | DELETE /api/work-orders/{id} | Bearer WO role | draft status | 200 | manual |
| TC-WO-007 | P0 | POST .../move-parts-to-work | Bearer WO role | issued_by | 200, movement_document_id | manual |
| TC-WO-008 | P1 | POST /api/work-orders | Bearer WO | invalid customer_id | 4xx customer not found | manual |
| TC-WO-009 | P1 | POST /api/work-orders | Bearer WO | invalid **work_id** | 4xx work not found | manual |
| TC-WO-010 | P1 | POST /api/work-orders | Bearer WO | invalid **executor_id** / service_advisor_id | 4xx employee not found | manual |
| TC-WO-011 | P1 | GET /api/work-orders | Bearer | filter status, customer_id | 200 | manual |
| TC-WO-012 | P0 | gRPC ApplyMovementDocument | parts confirm | work_order_id, doc_id, status=confirmed | parts_issued=true | manual |
| TC-WO-013 | P1 | GET /healthz :8097 | — | service up | 200 | WO-013 |
| TC-WO-014 | P1 | GET /api/work-orders | Bearer | workorders-service down | **503** gateway | WO-014 |
| TC-WO-015 | P1 | POST /api/work-orders | Bearer WO | labor only work_id (auto-fill from works) | 200, description from catalog | manual |
| TC-WO-016 | P1 | POST .../move-parts-to-work | Bearer WO | повтор при draft doc | 4xx movement document exists | manual |

## Write roles

admin, manager, master, service_advisor, storekeeper, parts_manager

## E2E с parts + works + employees

1. Create WO: labor с `work_id`, `executor_id`, `service_advisor_id`  
2. move-parts-to-work → draft movement doc  
3. confirm movement doc (parts) → WO updated, stock −qty

## Изменения (v2)

- Обязательный **`work_id`** в каждой строке labor (справочник `works-service`)
- Валидация **`service_advisor_id`** и **`executor_id`** через `employees-service`
- Поля ответа: `movement_document_id`, `movement_document_status`, `service_advisor_name`, `labor[].work_id`, `labor[].executor_name`
