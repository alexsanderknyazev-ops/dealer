# gateway-employee — тест-кейсы

**Сервис:** `gateway-service`  
**HTTP:** `:8090`  
**Роль:** grpc-gateway → 13 backend gRPC

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-GW-001 | P0 | GET /healthz | — | — | 200 | GW-001 |
| TC-GW-002 | P0 | OPTIONS /api/customers | — | CORS preflight | 204, Access-Control-Allow-Origin | GW-002 |
| TC-GW-003 | P1 | GET /api/customers | Bearer | Валидный JWT | 200, JSON customers[] | GW-003 |
| TC-GW-004 | P1 | GET /api/customers | — | Без JWT | 401 | GW-004 |
| TC-GW-005 | P1 | GET /api/vehicles | Bearer | — | 200 | GW-005 |
| TC-GW-006 | P1 | GET /api/deals | Bearer | — | 200 | GW-006 |
| TC-GW-007 | P1 | GET /api/parts | Bearer | — | 200 | GW-007 |
| TC-GW-008 | P1 | GET /api/brands | Bearer | — | 200 | GW-008 |
| TC-GW-009 | P1 | GET /api/dealer-points | Bearer | — | 200 | GW-009 |
| TC-GW-010 | P1 | GET /api/work-orders | Bearer | — | 200 или 503 если WO down | GW-010 |
| TC-GW-011 | P1 | GET /api/movement-documents | Bearer | — | 200 | GW-011 |
| TC-GW-012 | P1 | GET /api/reviews | Bearer | — | 200 | GW-012 |
| TC-GW-013 | P1 | GET /api/stats/employee/overview | Bearer | — | 200, overview | GW-013 |
| TC-GW-014 | P1 | GET /api/stats/client/overview | Bearer | — | 200, overview | GW-014 |
| TC-GW-015 | P2 | GET /api/parts/{invalid-uuid} | Bearer | — | 400/404 | GW-015 |
| TC-GW-016 | P1 | GET /api/works | Bearer | — | 200 или 503 | GW-016 |
| TC-GW-017 | P1 | GET /api/employees | Bearer | — | 200 или 503 | GW-017 |
| TC-GW-018 | P2 | Authorization header | Bearer | Заголовок пробрасывается в gRPC | backend видит JWT | manual |

## Зарегистрированные backend-ы

auth, customers, vehicles, deals, parts, brands, dealer-points, employee-statistics, client-statistics, employee-reviews, work-orders, **works**, **employees**
