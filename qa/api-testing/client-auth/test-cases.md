# client-auth — тест-кейсы

**Сервис:** `client-auth-service`  
**gRPC:** `:50059`  
**HTTP (direct):** `:8088` health only; auth через gateway

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-CA-001 | P0 | POST /api/login | public gw | client email/password | 200, JWT role=client | CA-001 |
| TC-CA-002 | P0 | GET /api/me | protected gw | Bearer client token | 200, user_id, valid | CA-002 |
| TC-CA-003 | P0 | Kafka consumer | — | client.registration.v1 после register | user credentials в Postgres | manual |
| TC-CA-004 | P1 | POST /api/login | — | wrong password | 401 | CA-004 |
| TC-CA-005 | P1 | gRPC IssueTokens | internal | user_id после Kafka | tokens (registration retry) | manual |
| TC-CA-006 | P1 | POST /api/refresh | — | expired refresh | 401 | manual |
| TC-CA-007 | P2 | Redis session | — | logout → refresh invalid | 401 on refresh | manual |

## Public vs Session services

- **ClientAuthPublicService** — login, refresh, logout (8091)
- **ClientAuthSessionService** — Validate → GET /api/me (8093)
