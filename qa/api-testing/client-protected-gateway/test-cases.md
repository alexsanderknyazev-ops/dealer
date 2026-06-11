# client-protected-gateway — тест-кейсы

**Сервис:** `client-protected-gateway-service`  
**HTTP:** `:8093`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-CPP-001 | P0 | GET /api/me | Bearer client | — | 200 | CPP-001 |
| TC-CPP-002 | P0 | GET /api/client/profile | Bearer client | — | 200 | CPP-002 |
| TC-CPP-003 | P0 | GET /api/client/vehicles | Bearer client | — | 200 | CPP-003 |
| TC-CPP-004 | P0 | GET /api/client/reviews | Bearer client | — | 200 | CPP-004 |
| TC-CPP-005 | P1 | GET /api/me | Bearer **employee** token | — | 401/403 | CPP-005 |
| TC-CPP-006 | P1 | GET /api/client/profile | — | Без JWT | 401 | CPP-006 |
| TC-CPP-007 | P1 | Authorization header | Bearer | Проброс в gRPC metadata | backend auth OK | manual |

## Backends

- client-auth (session Validate)
- client-registration (account)
- client-reviews
