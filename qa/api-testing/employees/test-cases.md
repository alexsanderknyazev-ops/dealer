# employees — тест-кейсы

**Сервис:** `employees-service`  
**REST:** `/api/employees`  
**gRPC:** `:50066` (GetEmployeeByUserID — только gRPC, без HTTP)  
**HTTP health:** `:8099/healthz`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-EMP-001 | P0 | GET /api/employees | Bearer | limit=20 | 200, employees[] | EMP-001 |
| TC-EMP-002 | P0 | GET /api/employees/{id} | Bearer | QA Master id | 200, full_name | EMP-002 |
| TC-EMP-003 | P1 | GET /api/employees | Bearer | search=Master | 200, filtered | manual |
| TC-EMP-004 | P1 | GET /api/employees | Bearer | active_only=true | 200 | manual |
| TC-EMP-005 | P0 | POST /api/employees | Bearer admin | user_id, full_name, position | 200 | manual |
| TC-EMP-006 | P1 | POST /api/employees | Bearer master | — | **403** | EMP-006 |
| TC-EMP-007 | P1 | PUT /api/employees/{id} | Bearer admin | phone, active | 200 | manual |
| TC-EMP-008 | P1 | DELETE /api/employees/{id} | Bearer admin | inactive test row | 200 | manual |
| TC-EMP-009 | P1 | GET /api/employees/{missing} | Bearer | random uuid | 404 | manual |
| TC-EMP-010 | P1 | GET /healthz :8099 | — | service up | 200 | EMP-010 |
| TC-EMP-011 | P0 | WO response fields | Bearer master | GET work-order with service_advisor_id | service_advisor_name заполнен | manual |

## Write roles

admin, manager

## Связь с work-orders

- `service_advisor_id` и `labor[].executor_id` валидируются через employees gRPC
- ResolveRef принимает **employee.id** или **auth.users.id** (user_id)
- В ответе WO поле `service_advisor_name` — из EmployeeFullName
