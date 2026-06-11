# client-public-gateway — тест-кейсы

**Сервис:** `client-public-gateway-service`  
**HTTP:** `:8091`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-CPG-001 | P0 | POST /api/client/register | — | email, password, full_name, phone, **vin** | 200, tokens + client_id | CPG-001 |
| TC-CPG-002 | P1 | POST /api/client/register | — | Без vin | 400 required fields | CPG-002 |
| TC-CPG-003 | P1 | POST /api/client/register | — | vin не существует | 4xx vehicle not found | CPG-003 |
| TC-CPG-004 | P0 | POST /api/login | — | client credentials | 200, tokens | CPG-004 |
| TC-CPG-005 | P1 | POST /api/refresh | — | refresh_token | 200 | CPG-005 |
| TC-CPG-006 | P1 | POST /api/logout | — | refresh_token | 200/204 | CPG-006 |
| TC-CPG-007 | P1 | OPTIONS /api/login | — | CORS | 204 | CPG-007 |
| TC-CPG-008 | P2 | POST /api/client/register | — | duplicate email | 4xx | manual |

## Backend routing

- `/api/client/register` → client-registration-service
- `/api/login|refresh|logout` → client-auth-service (public)
