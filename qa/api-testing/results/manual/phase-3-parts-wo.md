# Phase 3 — Parts, Works, Employees, Work orders (L2)

**Write-токены:** `qa.master@test.local` или `qa.admin@test.local` / `Test1234!`  
**Read-only smoke:** любой employee JWT (auto-run использует sales → 403 на write — это ожидаемо)

## Entry criteria

- [ ] Phase 0–1 passed
- [ ] `apply-fixtures.sh` (включая `08_works_employees.sql`)
- [ ] Health: `:8097` workorders, `:8098` works, `:8099` employees

## Ручные кейсы (detailed-test-cases.md)

| Test ID | Область | Status | DB / async | Notes |
|---------|---------|--------|------------|-------|
| TC-PRT-D002 | parts create (master) | ☐ | part_stock | |
| TC-PRT-D003 | movement draft | ☐ | movement_documents | |
| TC-PRT-D004 | confirm document | ☐ | stock ↓, stock_movements | |
| TC-PRT-D005 | insufficient stock | ☐ | draft unchanged | |
| TC-WRK-D001 | list works | ☐ | works.works | |
| TC-WRK-D003 | create work (master) | ☐ | COUNT +1 | |
| TC-WRK-D005 | WO labor auto-fill | ☐ | work_id join | |
| TC-EMP-D002 | get employee | ☐ | employees.employees | |
| TC-EMP-D005 | WO + advisor/executor | ☐ | service_advisor_name | |
| TC-WO-D001 | create WO | ☐ | labor.work_id | |
| TC-WO-D003 | move-parts-to-work | ☐ | movement doc draft | |
| TC-WO-D004 | confirm E2E | ☐ | parts_issued=true | |
| TC-WO-D007 | sales 403 | ☐ | COUNT=0 | |
| **E2E-003** | full chain | ☐ | `db-check.sh wo-movement a6600001-0000-4000-8000-000000000001` | |

## Быстрые команды

```bash
export POSTGRES_DSN='postgres://dealer:PASSWORD@127.0.0.1:5433/dealer?sslmode=disable'
export EMPLOYEE_API=http://127.0.0.1:8090

# Login master
MASTER_TOKEN=$(curl -s -X POST "$EMPLOYEE_API/api/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa.master@test.local","password":"Test1234!"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

# Проверка справочников
curl -s -H "Authorization: Bearer $MASTER_TOKEN" "$EMPLOYEE_API/api/works?limit=5"
curl -s -H "Authorization: Bearer $MASTER_TOKEN" "$EMPLOYEE_API/api/employees?limit=5"

# DB после шага
./qa/api-testing/scripts/db-check.sh work a8800001-0000-4000-8000-000000000001
./qa/api-testing/scripts/db-check.sh employee a9900001-0000-4000-8000-000000000003
./qa/api-testing/scripts/db-check.sh wo-movement a6600001-0000-4000-8000-000000000001
```

## Протокол E2E-003

Скопировать шаблон: `results/templates/e2e-protocol.md` → `results/runs/<run-id>/manual/e2e-003.md`

**Phase 3 result:** ☐ PASS · ☐ FAIL · ☐ BLOCKED (works/employees/workorders down)
