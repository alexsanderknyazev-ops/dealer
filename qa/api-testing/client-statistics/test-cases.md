# client-statistics — тест-кейсы

**Сервис:** `client-statistics-service`  
**REST:** `GET /api/stats/client/overview`  
**gRPC:** `:50062`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-CST-001 | P0 | GET /api/stats/client/overview | Bearer employee | — | 200, overview | CST-001 |
| TC-CST-002 | P1 | GET /api/stats/client/overview | — | Без JWT | 401 | CST-002 |
| TC-CST-003 | P0 | Kafka client.registration | — | Client register → overview | registrations_count↑ | manual |
| TC-CST-004 | P0 | Kafka review.published | — | Client create review | reviews_count↑ | manual |

## Consumers

- `client.registration.v1` (group: client-statistics)
- `review.published.v1` (group: client-statistics)
