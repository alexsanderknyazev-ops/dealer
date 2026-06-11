# customers — детальные тест-кейсы (API + БД)

**REST:** `/api/customers`  
**Схема:** `customers.customers`

---

## TC-CUST-D001 — Create individual (P0)

### До
```sql
SELECT COUNT(*) AS n FROM customers.customers;
```

### API
```bash
curl -s -X POST "$EMPLOYEE_API/api/customers" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "type": "individual",
    "full_name": "Иванов Иван",
    "phone": "+79001234567",
    "email": "ivanov@test.local",
    "address": "Москва",
    "notes": "QA create"
  }'
```

Сохранить: `CUSTOMER_ID` из `id`

### БД
```sql
SELECT id, name, email, phone, customer_type, address, notes
FROM customers.customers WHERE id = '<CUSTOMER_ID>';
```
| API | БД column |
|-----|-----------|
| full_name | name |
| type=individual | customer_type='individual' |

### Delta
- `COUNT(*)` + 1

---

## TC-CUST-D002 — Get by id (P0)

### API
```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/customers/$CUSTOMER_ID"
```

### Проверка
- Response `customer.id` = CUSTOMER_ID
- Поля совпадают с БД (SELECT выше)

---

## TC-CUST-D003 — List + pagination (P1)

```bash
curl -s -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/customers?limit=5&offset=0"
```

### БД cross-check
```sql
SELECT COUNT(*) FROM customers.customers;
-- response.total должен совпадать
```

---

## TC-CUST-D004 — Update (P1)

### API
```bash
curl -s -X PUT "$EMPLOYEE_API/api/customers/$CUSTOMER_ID" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+79009998877", "notes": "updated"}'
```

### БД
```sql
SELECT phone, notes, updated_at > created_at AS was_updated
FROM customers.customers WHERE id = '<CUSTOMER_ID>';
```
- phone = +79009998877
- was_updated = true

---

## TC-CUST-D005 — Delete (P1)

### До
```sql
SELECT COUNT(*) FROM customers.customers WHERE id = '<CUSTOMER_ID>';
-- 1
```

### API
```bash
curl -s -X DELETE -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  "$EMPLOYEE_API/api/customers/$CUSTOMER_ID"
```

### БД
```sql
SELECT COUNT(*) FROM customers.customers WHERE id = '<CUSTOMER_ID>';
-- 0
```

---

## TC-CUST-D006 — No auth (P1)

- POST/GET без Bearer → **401**
- COUNT в БД без изменений

---

## TC-CUST-D007 — Create legal entity (P1)

### API
```json
{
  "type": "legal",
  "full_name": "ООО Рога",
  "inn": "7707083893",
  "phone": "+7495",
  "email": "legal@test.local"
}
```

### БД
```sql
SELECT customer_type, inn, name FROM customers.customers WHERE inn = '7707083893';
-- customer_type = 'legal'
```

---

## TC-CUST-D008 — Referenced by deal (integration hint)

После create customer используется в deals — удаление customer с active deal:
- ожидание: delete **может** пройти (нет FK в deals) — зафиксировать фактическое поведение

```sql
SELECT COUNT(*) FROM deals.deals WHERE customer_id = '<CUSTOMER_ID>';
```
