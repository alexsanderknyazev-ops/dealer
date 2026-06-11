# QA fixtures — тестовые INSERT-ы

Фиксированные UUID и известные credentials для детальных тест-кейсов (`detailed-test-cases.md`).

## Порядок применения

```bash
export POSTGRES_DSN='postgres://dealer:PASSWORD@127.0.0.1:5433/dealer?sslmode=disable'

# 1. Миграции
make migrate

# 2. Базовый seed проекта (клиенты, бренды, склады, запчасти demo)
make full-seed

# 3. QA-фикстуры (идемпотентно)
./qa/api-testing/fixtures/apply-fixtures.sh

# 4. Admin из Makefile (опционально, если нужен admin@dealer.local)
make seed-admin
```

## Пароли

| Пользователь | Email | Password | Role |
|--------------|-------|----------|------|
| QA Admin | qa.admin@test.local | Test1234! | admin |
| QA Sales | qa.sales@test.local | Test1234! | sales |
| QA Master | qa.master@test.local | Test1234! | master |
| QA Parts Mgr | qa.parts@test.local | Test1234! | parts_manager |
| QA Storekeeper | qa.storekeeper@test.local | Test1234! | storekeeper |
| QA Client | qa.client@test.local | Test1234! | (clientauth) |

Bcrypt hash для `Test1234!` (cost 10):
`$2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy`

## Ключевые ID (см. `manifest.json`)

| Сущность | UUID | Ключ |
|----------|------|------|
| Vehicle (client reg) | a3300001-…-000000000001 | VIN `QAVINCLIENT001` |
| Customer | a2200001-…-000000000001 | — |
| Part (stock 50) | a4400001-…-000000000001 | SKU `QA-PART-001` |
| Part (stock 5) | a4400001-…-000000000002 | SKU `QA-PART-LOW` |
| Work order draft | a6600001-…-000000000001 | WO-QA-0001 |
| Work (diagnostic) | a8800001-…-000000000001 | LAB-QA-004 |
| Employee (master) | a9900001-…-000000000003 | QA Master |
| Deal draft | a5500001-…-000000000001 | — |

## Зависимости от `full-seed`

QA-фикстуры ссылаются на ID из `migrations/seed_dealer_brands.sql`:

- Dealer point Москва: `10000000-0000-4000-8000-000000000001`
- Warehouse parts Москва: `30000000-0000-4000-8000-000000000002`
- Brand Hyundai: `40000000-0000-4000-8000-000000000003`

## Очистка

```bash
psql "$POSTGRES_DSN" -f qa/api-testing/fixtures/99_cleanup.sql
```

Удаляет только записи с префиксом UUID `a1100001` / `a2200001` / … (QA namespace).

## Login для тестов

```bash
curl -s -X POST http://127.0.0.1:8090/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa.admin@test.local","password":"Test1234!"}'
```

Client (после apply `06_client.sql`):

```bash
curl -s -X POST http://127.0.0.1:8091/api/login \
  -d '{"email":"qa.client@test.local","password":"Test1234!"}'
```
