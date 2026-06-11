# employee-auth — тест-кейсы

**Сервис:** `auth-service`  
**HTTP:** `:8080` (auth + proxy на gateway)  
**gRPC:** `:50051`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-AUTH-001 | P0 | POST /api/register | — | email, password, name, phone уникальные | 200, user_id, access_token, refresh_token, expires_at | AUTH-001 |
| TC-AUTH-002 | P0 | POST /api/login | — | Валидные credentials | 200, tokens | AUTH-002 |
| TC-AUTH-003 | P0 | GET /api/me | Bearer | access_token из login | 200, user_id, email, valid=true | AUTH-003 |
| TC-AUTH-004 | P1 | POST /api/refresh | — | refresh_token | 200, новая пара tokens | AUTH-004 |
| TC-AUTH-005 | P1 | POST /api/logout | — | refresh_token | 200/204, refresh инвалидирован | AUTH-005 |
| TC-AUTH-006 | P1 | POST /api/register | — | Дубликат email | 4xx, ошибка «already exists» | AUTH-006 |
| TC-AUTH-007 | P1 | POST /api/login | — | Неверный password | 401/403 | AUTH-007 |
| TC-AUTH-008 | P1 | GET /api/me | — | Без Authorization | 401 | AUTH-008 |
| TC-AUTH-009 | P1 | GET /api/me | Bearer | Просроченный/битый token | 401 | manual |
| TC-AUTH-010 | P2 | POST /api/register | — | Пустой email или короткий password | 400 | AUTH-010 |
| TC-AUTH-011 | P1 | GET /healthz | — | — | 200 | AUTH-011 |
| TC-AUTH-012 | P1 | GET /readyz | — | Postgres + Redis up | 200 | AUTH-012 |
| TC-AUTH-013 | P2 | gRPC RegisterClient | internal | Только service-to-service | 200, role=client | manual |
| TC-AUTH-014 | P2 | Kafka auth.events | — | После register employee | event user.registered | manual |

## Proxy через auth (:8080)

| ID | P | Endpoint | Steps | Expected | Auto |
|----|---|----------|-------|----------|------|
| TC-AUTH-020 | P1 | GET /api/customers | Bearer через :8080 | 200 (proxy → gateway) | AUTH-020 |
| TC-AUTH-021 | P1 | POST /api/telemetry/events | JSON js_error | 204 (proxy → errors-ingest) | AUTH-021 |

## RBAC при регистрации

- Новый employee получает role **sales** (проверить payload JWT на jwt.io или decode)
