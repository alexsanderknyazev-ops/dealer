# customers — тест-кейсы

**Сервис:** `customers-service`  
**REST:** `/api/customers` (через gateway :8090)  
**gRPC:** `:50052`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-CUST-001 | P0 | POST /api/customers | Bearer sales+ | type=individual, full_name, phone, email | 200, customer.id | CUST-001 |
| TC-CUST-002 | P0 | GET /api/customers/{id} | Bearer | id из create | 200, customer | CUST-002 |
| TC-CUST-003 | P0 | GET /api/customers | Bearer | limit=10, offset=0 | 200, customers[], total | CUST-003 |
| TC-CUST-004 | P1 | PUT /api/customers/{id} | Bearer | Обновить phone | 200, обновлённые поля | CUST-004 |
| TC-CUST-005 | P1 | DELETE /api/customers/{id} | Bearer | id без активных deals | 200/204 | CUST-005 |
| TC-CUST-006 | P1 | POST /api/customers | — | Без JWT | 401 | CUST-006 |
| TC-CUST-007 | P1 | GET /api/customers/{uuid} | Bearer | Несуществующий id | 404 | CUST-007 |
| TC-CUST-008 | P2 | POST /api/customers | Bearer | Пустой full_name | 400 | CUST-008 |
| TC-CUST-009 | P1 | POST /api/customers | Bearer | type=legal_entity + company fields | 200 | manual |
| TC-CUST-010 | P2 | GET /api/customers | Bearer | search query param | 200, фильтрация | manual |

## Межсервисные зависимости

- **deals-service** вызывает `GetCustomer` при create/update deal
- **workorders-service** проверяет `CustomerExists` при create WO
