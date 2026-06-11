# employee-statistics — тест-kейсы

**Сервис:** `employee-statistics-service`  
**REST:** `GET /api/stats/employee/overview`  
**gRPC:** `:50061`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-EST-001 | P0 | GET /api/stats/employee/overview | Bearer | — | 200, overview (deals, revenue, …) | EST-001 |
| TC-EST-002 | P1 | GET /api/stats/employee/overview | — | Без JWT | 401 | EST-002 |
| TC-EST-003 | P0 | Kafka deal.completed | — | Complete deal → подождать 5s → GET overview | метрики выросли | manual |
| TC-EST-004 | P2 | GET /metrics :8094 | — | — | 200 Prometheus | manual |

## Consumer

- Topic: `deal.completed.v1`, group: `employee-statistics`
