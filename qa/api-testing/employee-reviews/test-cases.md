# employee-reviews — тест-кейсы

**Сервис:** `employee-reviews-service`  
**REST:** `/api/reviews`, `/api/clients/{client_id}/reviews`, `/api/reviews/stats`  
**gRPC:** `:50063`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-EREV-001 | P0 | GET /api/reviews | Bearer | — | 200, reviews[] | EREV-001 |
| TC-EREV-002 | P1 | GET /api/reviews | Bearer | limit, offset, status filter | 200 | manual |
| TC-EREV-003 | P1 | GET /api/clients/{client_id}/reviews | Bearer | client_id из client-registration | 200 | EREV-003 |
| TC-EREV-004 | P0 | GET /api/reviews/stats | Bearer | — | 200, aggregates | EREV-004 |
| TC-EREV-005 | P1 | GET /api/reviews | — | Без JWT | 401 | EREV-005 |
| TC-EREV-006 | P0 | Kafka consumer | — | После client create review | отзыв появляется в list (async ≤10s) | manual |
| TC-EREV-007 | P2 | GET /api/clients/{bad-id}/reviews | Bearer | — | 200 empty или 404 | manual |

## Источник данных

- Kafka `review.published.v1` от client-reviews-service
- Enrichment: gRPC GetVehicle (vehicles-service)
