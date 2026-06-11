# client-reviews — тест-кейсы

**Сервис:** `client-reviews-service`  
**REST:** `/api/client/reviews` (protected gateway :8093)  
**gRPC:** `:50060`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-CREV-001 | P0 | GET /api/client/reviews | Bearer client | — | 200, reviews[] | CREV-001 |
| TC-CREV-002 | P0 | POST /api/client/reviews | Bearer client | vehicle_id, rating 1-5, text | 200, review status=pending | CREV-002 |
| TC-CREV-003 | P1 | GET /api/client/reviews/{id} | Bearer client | own review id | 200 | CREV-003 |
| TC-CREV-004 | P1 | POST /api/client/reviews | Bearer client | чужой vehicle_id | 4xx | manual |
| TC-CREV-005 | P1 | POST /api/client/reviews | Bearer client | rating=0 или 6 | 400 | manual |
| TC-CREV-006 | P1 | GET /api/client/reviews | — | Без JWT | 401 | CREV-006 |
| TC-CREV-007 | P0 | Kafka review.published | — | после create | event в topic | manual |
| TC-CREV-008 | P1 | gRPC GetVehicle | internal | vehicle_id validate | enrich dealer_point_id | manual |

## Side-effects

- Kafka → employee-reviews, client-statistics
