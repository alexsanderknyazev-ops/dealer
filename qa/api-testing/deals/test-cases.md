# deals — тест-кейсы

**Сервис:** `deals-service`  
**REST:** `/api/deals`  
**gRPC:** `:50054`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-DEL-001 | P0 | GET /api/deals | Bearer | — | 200, deals[] | DEL-001 |
| TC-DEL-002 | P0 | POST /api/deals | Bearer sales+ | customer_id, vehicle_id, amount, stage=draft | 200, deal | DEL-002 |
| TC-DEL-003 | P0 | GET /api/deals/{id} | Bearer | — | 200 | DEL-003 |
| TC-DEL-004 | P1 | PUT /api/deals/{id} | Bearer | stage=in_progress | 200 | DEL-004 |
| TC-DEL-005 | P0 | PUT /api/deals/{id} | Bearer | stage=completed | 200, Kafka deal.completed.v1 | DEL-005 |
| TC-DEL-006 | P1 | DELETE /api/deals/{id} | Bearer | — | 200/204 | manual |
| TC-DEL-007 | P1 | POST /api/deals | Bearer | fake customer_id | 4xx customer not found | DEL-007 |
| TC-DEL-008 | P1 | POST /api/deals | Bearer | fake vehicle_id | 4xx vehicle not found | DEL-008 |
| TC-DEL-009 | P1 | POST /api/deals | — | Без JWT | 401 | DEL-009 |
| TC-DEL-010 | P2 | PUT /api/deals/{id} | Bearer | invalid stage transition | 4xx | manual |

## Kafka side-effect

При stage=**completed** → topic `deal.completed.v1` → employee-statistics consumer
