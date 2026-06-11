# employees — детальные тест-кейсы (API + БД)

**REST:** `/api/employees`  
**Схема:** `employees.employees`  
**Precondition:** fixtures `08_works_employees.sql`, admin token для write

IDs: `manifest.json` → `employees.master`, `employees.admin`

---

## TC-EMP-D001 — List employees (P0)

```bash
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$EMPLOYEE_API/api/employees?limit=20"
```

### БД
```sql
SELECT COUNT(*) FROM employees.employees WHERE active = true;
SELECT id, user_id, full_name, position FROM employees.employees
WHERE full_name LIKE 'QA %' ORDER BY full_name;
```

---

## TC-EMP-D002 — Get employee by id (P0)

```bash
EMP_ID="a9900001-0000-4000-8000-000000000003"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$EMPLOYEE_API/api/employees/$EMP_ID"
```

### БД
```sql
SELECT id, user_id, full_name, position, active
FROM employees.employees WHERE id = '$EMP_ID';
-- full_name = QA Master, user_id = a1100001-...-003
```

---

## TC-EMP-D003 — Create employee (P1, admin)

```bash
curl -s -X POST "$EMPLOYEE_API/api/employees" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "",
    "full_name": "QA Mechanic Temp",
    "position": "mechanic",
    "department": "СТО",
    "phone": "+79009998877",
    "active": true
  }'
```

### БД
```sql
SELECT * FROM employees.employees WHERE full_name = 'QA Mechanic Temp';
```

---

## TC-EMP-D004 — Master forbidden write (P1)

POST /api/employees с `MASTER_TOKEN` → **403**

---

## TC-EMP-D005 — Work order with service advisor + executor (P0)

### API — create WO
```bash
curl -s -X POST "$EMPLOYEE_API/api/work-orders" \
  -H "Authorization: Bearer $MASTER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "customer_id": "a2200001-0000-4000-8000-000000000001",
    "vehicle_id": "a3300001-0000-4000-8000-000000000001",
    "dealer_point_id": "10000000-0000-4000-8000-000000000001",
    "warehouse_id": "30000000-0000-4000-8000-000000000002",
    "service_advisor_id": "a9900001-0000-4000-8000-000000000003",
    "labor": [{
      "work_id": "a8800001-0000-4000-8000-000000000001",
      "executor_id": "a9900001-0000-4000-8000-000000000003",
      "sort_order": 1
    }],
    "parts": []
  }'
```

### HTTP response
- `service_advisor_name` = `QA Master` (не пусто)
- `labor[0].executor_name` если реализовано в proto (опционально)

### БД
```sql
SELECT service_advisor_id, order_number FROM workorders.work_orders WHERE id = '<WO_ID>';
SELECT work_id, executor_id FROM workorders.work_order_labor WHERE work_order_id = '<WO_ID>';
```

---

## TC-EMP-D006 — Invalid employee ref on WO (P1)

`service_advisor_id` = random uuid → **4xx** employee not found, без новой строки в work_orders.

---

## TC-EMP-D007 — ResolveRef by user_id (manual gRPC)

Internal: workorders вызывает `EmployeeExists` / `EmployeeFullName` с id мастера **или** user_id из auth — оба должны проходить валидацию.

```sql
-- employee row for qa.master user
SELECT e.id, e.user_id, u.email
FROM employees.employees e
JOIN auth.users u ON u.id = e.user_id
WHERE u.email = 'qa.master@test.local';
```

Создать WO с `service_advisor_id` = **user_id** (не employee.id) — ожидается 200.
