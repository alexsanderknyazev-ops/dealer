# Полное API-тестирование Dealer

**Мастер-план:** [`TEST-PLAN.md`](TEST-PLAN.md) — фазы, scope, P0, RBAC, расписание.

Структура тест-кейсов по микросервисам + автоматический прогон.

## Структура

```
qa/api-testing/
├── README.md
├── TEST-PLAN.md
├── results/                  ← результаты прогонов
│   ├── runs/                 ← каждый auto-run
│   ├── latest/               ← последний smoke
│   ├── manual/               ← протоколы фаз
│   └── templates/
├── fixtures/                 ← SQL INSERT-ы + manifest.json
├── scripts/
├── employee-auth/
├── gateway-employee/
├── customers/
├── vehicles/
├── brands/
├── dealer-points/
├── parts/
├── work-orders/
├── works/
├── employees/
├── deals/
├── employee-reviews/
├── employee-statistics/
├── client-statistics/
├── client-public-gateway/
├── client-auth/
├── client-registration/
├── client-reviews/
├── client-protected-gateway/
├── errors-ingest/
└── integration/
```

## SQL-фикстуры (тестовые данные)

```bash
make migrate && make full-seed
export POSTGRES_DSN='postgres://dealer:PASSWORD@127.0.0.1:5433/dealer?sslmode=disable'
./qa/api-testing/fixtures/apply-fixtures.sh
```

- [`fixtures/README.md`](fixtures/README.md) — пользователи, пароли, порядок
- [`fixtures/manifest.json`](fixtures/manifest.json) — все фиксированные UUID/VIN/SKU

## Два уровня документации

| Файл | Назначение |
|------|------------|
| `test-cases.md` | Краткий чеклист (smoke, матрица endpoint × код) |
| `detailed-test-cases.md` | **Полный сценарий:** curl → сохранить ID → SQL в Postgres → Redis/Kafka |

Общие гайды:
- [`_shared/db-verification.md`](_shared/db-verification.md) — схемы и паттерн проверки
- [`_shared/db-queries.sql`](_shared/db-queries.sql) — шаблоны SQL
- [`_shared/redis-kafka-checks.md`](_shared/redis-kafka-checks.md) — refresh tokens, async events
- [`_shared/test-users.md`](_shared/test-users.md) — роли admin/master для parts/WO

## Быстрый старт (после поднятия инфраструктуры)

```bash
export POSTGRES_DSN='postgres://dealer:PASSWORD@127.0.0.1:5433/dealer?sslmode=disable'

# Тестовые INSERT-ы (users, VIN, parts, WO, client)
make migrate && make full-seed
./qa/api-testing/fixtures/apply-fixtures.sh

# Smoke-прогон (HTTP) → отчёт в results/
./qa/api-testing/scripts/run-api-tests.sh
cat qa/api-testing/results/latest/smoke-report.md

# Проверка записи в БД после ручного API-шага
./qa/api-testing/scripts/db-check.sh customer <UUID>
./qa/api-testing/scripts/db-check.sh wo-movement <WO_UUID>

# E2E с БД — следовать integration/detailed-test-cases.md
```

## Точки входа (локально)

| Контур | URL | Порт |
|--------|-----|------|
| Employee (auth UI + proxy) | http://127.0.0.1:8080 | 8080 |
| Employee gateway | http://127.0.0.1:8090 | 8090 |
| Client public | http://127.0.0.1:8091 | 8091 |
| Client protected | http://127.0.0.1:8093 | 8093 |
| Errors ingest | http://127.0.0.1:8092 | 8092 |

## Формат ID тест-кейсов

`TC-{SERVICE}-{NNN}` — например `TC-CUST-001`.

Приоритеты: **P0** (critical path), **P1** (важный), **P2** (границы/регрессия).

## RBAC (employee)

| Роль | Write: customers, vehicles, deals, brands, dealer-points | Write: parts, work-orders |
|------|-------------------------------------------------------------|---------------------------|
| sales (дефолт при register) | ✅ | ❌ 403 |
| admin, manager, parts_manager, storekeeper, master, service_advisor | ✅ | ✅ (см. proto) |

Для parts/work-orders: `qa.admin@test.local` или `qa.master@test.local`, пароль `Test1234!` (см. `fixtures/`).
