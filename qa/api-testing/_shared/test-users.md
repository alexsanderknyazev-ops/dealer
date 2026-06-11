# Тестовые пользователи и роли

## Тестовые пользователи

### Через SQL-фикстуры (рекомендуется для QA)

```bash
make migrate && make full-seed
export POSTGRES_DSN='...'
./qa/api-testing/fixtures/apply-fixtures.sh
```

См. [`fixtures/README.md`](fixtures/README.md) и [`fixtures/manifest.json`](fixtures/manifest.json).

| Email | Password | Role |
|-------|----------|------|
| qa.admin@test.local | Test1234! | admin |
| qa.sales@test.local | Test1234! | sales |
| qa.master@test.local | Test1234! | master |
| qa.parts@test.local | Test1234! | parts_manager |
| qa.client@test.local | Test1234! | client (B2C) |

### Вручную через API (role=sales по умолчанию)

```bash
curl -s -X POST http://127.0.0.1:8090/api/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa-sales@test.local","password":"Test1234!","name":"QA Sales","phone":"+7900"}'
```

Проверка роли в БД:

```sql
SELECT email, role FROM auth.users WHERE email = 'qa-sales@test.local';
-- ожидание: role = 'sales'
```

## Employee — admin / master (для parts, work-orders write)

RBAC write для parts/work-orders **не** даёт role=sales. Поднять вручную:

```sql
UPDATE auth.users SET role = 'admin' WHERE email = 'qa-admin@test.local';
-- или: master, parts_manager, storekeeper, service_advisor
```

Затем login и использовать token admin-пользователя.

## Client

Создаётся только через `POST /api/client/register` (8091) с валидным VIN.

```sql
SELECT c.id, c.email, c.full_name, u.id AS user_id
FROM clients.clients c
JOIN clientauth.users u ON u.id = c.user_id
WHERE c.email = '<CLIENT_EMAIL>';
```

## Переменные для curl-скриптов

```bash
export EMPLOYEE_API=http://127.0.0.1:8090
export CLIENT_PUBLIC=http://127.0.0.1:8091
export CLIENT_PROTECTED=http://127.0.0.1:8093
export QA_PASSWORD='Test1234!'
export EMPLOYEE_TOKEN='...'
export CLIENT_TOKEN='...'
```
