# План тестирования — Dealer Platform

| | |
|---|---|
| **Версия документа** | 1.0 |
| **Дата** | 2026-06-12 |
| **Объект** | Микросервисная платформа автодилера (employee + client контуры) |
| **Артефакты** | `qa/api-testing/` — кейсы, фикстуры, скрипты |
| **Статус** | Draft — к исполнению после поднятия стенда |

---

## 1. Цели и задачи

### 1.1 Цели

- Подтвердить корректность **REST API** всех сервисов через grpc-gateway.
- Проверить **целостность данных** в PostgreSQL после операций (CREATE/UPDATE/DELETE, stock, movement).
- Проверить **межсервисное взаимодействие**: gRPC sync, Kafka async, Redis sessions.
- Подтвердить **RBAC** (роли employee) и **изоляцию контуров** (employee vs client JWT).
- Обеспечить **go/no-go** перед релизом по critical path сценариям (P0).

### 1.2 Задачи

| # | Задача | Результат |
|---|--------|-----------|
| T1 | Smoke всех gateway-маршрутов | `results/latest/smoke-report.md`, 0 P0 fail |
| T2 | Детальные кейсы с проверкой БД | Протокол по `detailed-test-cases.md` |
| T3 | E2E бизнес-потоки | E2E-001 … E2E-006 passed |
| T4 | Негативные и RBAC | 401/403/404/400 по матрице |
| T5 | Async (Kafka) | Events в consumer-таблицах ≤15 с |
| T6 | Регрессия после фиксов | Повтор P0 + затронутой области |

---

## 2. Область тестирования

### 2.1 In scope

| Область | Компоненты |
|---------|------------|
| **Employee HTTP** | auth :8080, gateway :8090 |
| **Client HTTP** | public :8091, protected :8093 |
| **Backend gRPC** | 17+ сервисов (см. §4) |
| **Данные** | PostgreSQL (все схемы из `pkg/dbschema/schemas.go`) |
| **Сессии** | Redis refresh tokens |
| **События** | Kafka: `client.registration.v1`, `review.published.v1`, `deal.completed.v1`, `platform.errors.v1` |
| **Observability** | `/healthz`, `/readyz`, `/metrics`, telemetry → errors-ingest |
| **Frontend API** | `frontend/auth`, `frontend/client` — контракт с gateway (ручная выборочная проверка) |

### 2.2 Out of scope (v1.0)

- Нагрузочное / stress-тестирование (отдельный план при необходимости).
- Penetration testing / OWASP полный аудит.
- Тестирование Jenkins/K8s деплоя (только smoke health после deploy).
- UI/E2E Playwright (есть `frontend/auth/tests/e2e/` — опциональная фаза 6).
- ClickHouse analytics — только smoke telemetry (если сервис поднят).

---

## 3. Архитектура и точки входа

```mermaid
flowchart LR
  subgraph employee
    FE1[frontend/auth] --> AUTH8080[auth :8080]
    AUTH8080 --> GW8090[gateway :8090]
  end
  subgraph client
    FE2[frontend/client] --> PUB8091[public :8091]
    FE2 --> PROT8093[protected :8093]
  end
  GW8090 --> SVC[Microservices gRPC]
  PUB8091 --> SVC
  PROT8093 --> SVC
  SVC --> PG[(PostgreSQL)]
  SVC --> Redis[(Redis)]
  SVC --> Kafka[(Kafka)]
```

| Контур | Base URL | Auth |
|--------|----------|------|
| Employee gateway | http://127.0.0.1:8090 | Bearer JWT employee |
| Employee auth UI | http://127.0.0.1:8080 | + proxy на gateway |
| Client public | http://127.0.0.1:8091 | нет |
| Client protected | http://127.0.0.1:8093 | Bearer JWT client |
| Errors ingest | http://127.0.0.1:8092 | нет |

---

## 4. Матрица сервисов и покрытие

| Сервис | gRPC | REST prefix | Папка кейсов | P0 кейсов | Auto smoke |
|--------|------|-------------|--------------|-----------|------------|
| employee-auth | :50051 | /api/register, login, me | `employee-auth/` | 5 | AUTH-* |
| gateway-employee | — | все /api/* | `gateway-employee/` | 4 | GW-* |
| customers | :50052 | /api/customers | `customers/` | 4 | CUST-* |
| vehicles | :50053 | /api/vehicles | `vehicles/` | 4 | VEH-* |
| brands | :50056 | /api/brands | `brands/` | 2 | BRD-* |
| dealer-points | :50057 | /api/dealer-points, legal-entities, warehouses | `dealer-points/` | 3 | DP-* |
| parts | :50055 | /api/parts, movement-documents | `parts/` | 4 | PRT-* |
| work-orders | :50064 | /api/work-orders | `work-orders/` | 5 | WO-* |
| works | :50065 | /api/works | `works/` | 3 | WRK-* |
| employees | :50066 | /api/employees | `employees/` | 3 | EMP-* |
| deals | :50054 | /api/deals | `deals/` | 4 | DEL-* |
| employee-reviews | :50063 | /api/reviews | `employee-reviews/` | 2 | EREV-* |
| employee-statistics | :50061 | /api/stats/employee/overview | `employee-statistics/` | 2 | EST-* |
| client-statistics | :50062 | /api/stats/client/overview | `client-statistics/` | 2 | CST-* |
| client-public-gateway | — | register, login | `client-public-gateway/` | 3 | CPG-* |
| client-auth | :50059 | login, me (via gw) | `client-auth/` | 3 | CA-* |
| client-registration | :50058 | profile, vehicles | `client-registration/` | 4 | CR-* |
| client-reviews | :50060 | /api/client/reviews | `client-reviews/` | 3 | CREV-* |
| client-protected-gateway | — | protected routes | `client-protected-gateway/` | 3 | CPP-* |
| errors-ingest | — | /api/telemetry/events | `errors-ingest/` | 1 | ERR-* |
| **Integration** | — | cross-service | `integration/` | 6 E2E | INT-* |

---

## 5. Стратегия тестирования

### 5.1 Уровни

| Уровень | Описание | Инструмент | Когда |
|---------|----------|------------|-------|
| **L0 Smoke** | HTTP 200/401, health, gateway alive | `scripts/run-api-tests.sh` | Каждый deploy, CI (целевое) |
| **L1 API functional** | CRUD, валидация, коды ответов | `test-cases.md` | Sprint / feature complete |
| **L2 API + DB** | L1 + SQL snapshot/delta | `detailed-test-cases.md`, `db-check.sh` | Перед релизом |
| **L3 Integration** | Несколько сервисов + Kafka/Redis | `integration/detailed-test-cases.md` | Перед релизом |
| **L4 RBAC / Security** | Роли, cross-token, 403 | §7 + fixtures users | Перед релизом |
| **L5 Regression** | P0 + область изменений | Полный прогон P0 | После bugfix |
| **L6 UI (optional)** | Playwright deals-flow | `frontend/auth/tests/e2e/` | По запросу |

### 5.2 Приоритеты кейсов

| Priority | Критерий | Блокирует релиз |
|----------|----------|-----------------|
| **P0** | Critical path, деньги, stock, auth | **Да** |
| **P1** | Важные негативы, RBAC, Kafka | Нет (major bug) |
| **P2** | Границы, pagination, CORS | Нет |

### 5.3 Метод «API → DB → Async»

Каждый L2/L3 кейс выполняется по шаблону:

1. **Snapshot до** — `COUNT(*)` / SELECT (`_shared/db-verification.md`)
2. **HTTP request** — curl, сохранить `id` из JSON
3. **Assert HTTP** — status, body fields
4. **Snapshot после** — row exists, fields match
5. **Delta** — count ±1 или 0 при ошибке
6. **Async wait** — до 15 с, проверка consumer-таблицы или Redis key

---

## 6. Тестовое окружение

### 6.1 Требования

| Компонент | Версия / порт | Обязательно |
|-----------|---------------|-------------|
| PostgreSQL | 16, :5433 | да |
| Redis | 7, :6379 | да |
| Kafka | :9092 | да (async кейсы) |
| Docker Compose | `docker compose up -d` | да |
| Миграции | `make migrate` | да |
| JWT_SECRET | одинаковый на всех сервисах | да |

### 6.2 Подготовка стенда (checklist)

```bash
# 1. Инфра
docker compose up -d
docker compose ps   # все healthy / running

# 2. Миграции + базовый seed
export POSTGRES_DSN='postgres://dealer:PASSWORD@127.0.0.1:5433/dealer?sslmode=disable'
make migrate
make full-seed

# 3. QA-фикстуры (фиксированные UUID, пользователи, VIN)
./qa/api-testing/fixtures/apply-fixtures.sh

# 4. Опционально: admin@dealer.local
make seed-admin

# 5. Health
curl -sf http://127.0.0.1:8090/healthz
curl -sf http://127.0.0.1:8080/healthz
curl -sf http://127.0.0.1:8091/healthz || true
curl -sf http://127.0.0.1:8093/healthz || true
```

### 6.3 Entry criteria (начало тестирования)

- [ ] Postgres, Redis, Kafka доступны
- [ ] Миграции применены без ошибок
- [ ] `make full-seed` + `apply-fixtures.sh` выполнены
- [ ] Gateway :8090 отвечает 200 на `/healthz`
- [ ] workorders-service :8097 up (для P0 WO — иначе зафиксировать 503 как blocker)
- [ ] Известен `POSTGRES_DSN` для SQL-проверок

### 6.4 Exit criteria (завершение / go-live)

- [ ] **100% P0** passed (или documented waiver с owner approval)
- [ ] **≥95% P1** passed
- [ ] Все E2E-001 … E2E-004 passed
- [ ] Нет открытых **Critical/Blocker** дефектов
- [ ] `results/latest/smoke-report.md` + `results/INDEX.md` приложены к релизу
- [ ] Известные риски (§12) приняты product owner

---

## 7. Тестовые данные

### 7.1 Источники

| Источник | Назначение |
|----------|------------|
| `migrations/seed_*.sql` | Бренды, склады, demo parts |
| `fixtures/*.sql` | QA namespace, предсказуемые ID |
| `fixtures/manifest.json` | Справочник UUID/VIN/SKU |
| API register | Динамические пользователи (негативы) |

### 7.2 Ключевые QA-аккаунты (пароль `Test1234!`)

| Email | Role | Использование |
|-------|------|---------------|
| qa.admin@test.local | admin | parts, WO, movement full write |
| qa.sales@test.local | sales | deals, customers; **403** на parts |
| qa.master@test.local | master | WO, СТО сценарии |
| qa.parts@test.local | parts_manager | parts CRUD |
| qa.client@test.local | client | protected API без register flow |

### 7.3 Ключевые сущности

| Key | Value | Сценарий |
|-----|-------|----------|
| VIN | `QAVINCLIENT001` | Client register E2E-002 |
| SKU | `QA-PART-001` | Stock 50, WO issue |
| SKU | `QA-PART-LOW` | Insufficient stock (qty 5) |
| WO | `WO-QA-0001` / id `a6600001-…-000000000001` | move-parts-to-work |
| Deal draft | `a5500001-…-000000000001` | complete → Kafka |

### 7.4 Очистка между прогонами

```bash
psql "$POSTGRES_DSN" -f qa/api-testing/fixtures/99_cleanup.sql
./qa/api-testing/fixtures/apply-fixtures.sh
```

---

## 8. RBAC и безопасность

### 8.1 Матрица ролей (employee write)

| Endpoint group | sales | admin / master / parts_manager |
|----------------|-------|--------------------------------|
| /api/customers, vehicles, deals, brands, dealer-points | ✅ | ✅ |
| /api/parts, /api/movement-documents | ❌ 403 | ✅ |
| /api/work-orders | ❌ 403 | ✅ |
| GET (read) | ✅ | ✅ |

**Кейсы:** TC-PRT-D001, TC-WO-D007, INT-006 — `qa.sales@test.local` vs `qa.admin@test.local`.

### 8.2 Изоляция контуров

| Test | Expected |
|------|----------|
| Employee JWT → :8093 /api/client/profile | 401/403 |
| Client JWT → :8090 /api/customers | 401/403 |
| Без JWT на protected | 401 |

**Кейсы:** E2E-004, TC-CPP-D003.

### 8.3 Token lifecycle

| Step | Redis key | API |
|------|-----------|-----|
| login | `auth:refresh:{token}` EXISTS | 200 |
| refresh | old DEL, new SET | 200 |
| logout | DEL | refresh → 401 |

**Кейсы:** TC-AUTH-D004/D005, E2E-005.

---

## 9. Фазы выполнения

### Фаза 0 — Подготовка (0.5 дня)

| # | Действие | Ответственный | Done |
|---|----------|---------------|------|
| 0.1 | Поднять docker compose | QA/DevOps | ☐ |
| 0.2 | migrate + full-seed + fixtures | QA | ☐ |
| 0.3 | Проверить health всех сервисов | QA | ☐ |
| 0.4 | Login qa.admin, qa.client — smoke | QA | ☐ |

### Фаза 1 — Smoke L0 (0.5 дня)

```bash
./qa/api-testing/scripts/run-api-tests.sh
```

| Критерий | Pass |
|----------|------|
| FAIL = 0 по P0 auto IDs | ☐ |
| Отчёт в `results/latest/smoke-report.md` | ☐ |

### Фаза 2 — Employee core L2 (1–1.5 дня)

**Порядок:** auth → customers → vehicles → brands → dealer-points → deals

| Сервис | Файл | Мин. кейсов |
|--------|------|-------------|
| employee-auth | `employee-auth/detailed-test-cases.md` | D001–D008 |
| customers | `customers/detailed-test-cases.md` | D001–D005 |
| vehicles | `vehicles/detailed-test-cases.md` | D001–D004 |
| brands | `brands/detailed-test-cases.md` | D001–D003 |
| dealer-points | `dealer-points/detailed-test-cases.md` | D001–D003 |
| deals | `deals/detailed-test-cases.md` | D001, D004 (Kafka) |

### Фаза 3 — Parts & Work orders L2 (1–1.5 дня)

**Precondition:** login as `qa.admin@test.local`

| Сервис | Ключевые кейсы |
|--------|----------------|
| parts | PRT-D002, D003, D004 (confirm), D005 (insufficient) |
| works | WRK-D001, D003, D005 |
| employees | EMP-D001, D005 |
| work-orders | WO-D001, D001b, D003, D004 |
| **E2E-003** | Full WO → movement → confirm → stock ↓ |

**Cross-DB check:**

```bash
./qa/api-testing/scripts/db-check.sh wo-movement a6600001-0000-4000-8000-000000000001
```

### Фаза 4 — Client contour L2 (1 день)

| Сервис | Ключевые кейсы |
|--------|----------------|
| client-registration | CR-D001 (register + VIN), D004 profile |
| client-auth | CA-D001, D002 |
| client-reviews | REV-D001, D004 (Kafka → employee_reviews) |
| **E2E-002** | Register → profile → review → employee list |

### Фаза 5 — Statistics, reviews, telemetry (0.5 дня)

| Область | Кейсы |
|---------|-------|
| employee-statistics | EST-D002 после deal complete |
| client-statistics | CST-D002, D003 |
| employee-reviews | EREV-D003, D004 |
| errors-ingest | ERR-D001 (если :8092 up) |

### Фаза 6 — RBAC, негативы, регрессия (0.5–1 день)

- INT-004, INT-006
- Все P1 негативы (401, 404, duplicate VIN/email)
- Повтор P0 после фиксов

### Фаза 7 — UI optional

- `frontend/auth` — login, deals list
- `frontend/client` — register, reviews
- Playwright: `frontend/auth/tests/e2e/deals-flow.spec.ts`

---

## 10. Critical path (P0) — минимальный прогон

Если время ограничено, выполнить **только** этот список:

| # | Сценарий | ID |
|---|----------|-----|
| 1 | Employee login + me | TC-AUTH-D002, D003 |
| 2 | CRUD customer | TC-CUST-D001 |
| 3 | Create vehicle | TC-VEH-D001 |
| 4 | Create + complete deal + stats event | TC-DEL-D004, EST-D002 |
| 5 | Client register by VIN | TC-CR-D001 |
| 6 | Client review → employee sees | REV-D001, EREV-D004 |
| 7 | WO move-parts + confirm stock | E2E-003 |
| 8 | Sales 403 on parts | TC-PRT-D001 |
| 9 | Gateway smoke all routes | TC-GW-D006 |
| 10 | Auto script green | `run-api-tests.sh` |

**Оценка:** ~4–6 часов ручной работы + 15 мин auto.

---

## 11. Межсервисные потоки (обязательные)

| ID | Поток | Проверки |
|----|-------|----------|
| **E2E-001** | Employee: customer → vehicle → deal → complete | deals.deals, employee_statistics.deal_events |
| **E2E-002** | Client: register(VIN) → profile | clients.*, clientauth.users, Redis |
| **E2E-003** | WO → movement draft → confirm | parts.stock, workorders.parts_issued |
| **E2E-004** | Token isolation employee/client | 401/403, no DB writes |
| **E2E-005** | Refresh rotation / logout | Redis keys |
| **E2E-006** | Telemetry (optional) | 204, ClickHouse |

Детали: [`integration/detailed-test-cases.md`](integration/detailed-test-cases.md)

---

## 12. Риски и митигация

| Risk | Impact | Mitigation |
|------|--------|------------|
| workorders-service down → 503 | P0 WO blocked | Проверить `curl :8097/healthz`, docker logs |
| parts ↔ workorders circular gRPC | confirm fails mid-flight | Retry; проверить оба сервиса up |
| Kafka lag >15s | false fail async | Увеличить wait; проверить consumer group |
| sales user для parts тестов | ложный pass write | Использовать fixtures admin/master |
| `auth.events` без consumer | orphan topic | Out of scope v1; не блокер |
| Duplicate /api/login paths | wrong gateway | Явно указывать порт 8090 vs 8091 |
| JWT_SECRET mismatch | mass 401 | Один secret в docker-compose |

---

## 13. Дефекты и отчётность

### 13.1 Severity

| Severity | Пример |
|----------|--------|
| **Blocker** | Gateway down, login broken, stock deduct wrong |
| **Critical** | Deal complete без Kafka event, client register fail |
| **Major** | 500 на valid input, RBAC bypass |
| **Minor** | Неточное сообщение ошибки, P2 pagination |
| **Trivial** | Typo в response field name (non-breaking) |

### 13.2 Шаблон баг-репорта

```
Title: [SERVICE] краткое описание
Environment: local docker / commit SHA
Preconditions: user role, fixture IDs
Steps: 1…n
Expected: HTTP + DB state
Actual: …
Logs: service name, correlation if any
Artifacts: curl, SQL output, TEST-RUN-REPORT row ID
```

### 13.3 Deliverables

| Артефакт | Путь |
|----------|------|
| План тестирования | `TEST-PLAN.md` (этот документ) |
| Авто-отчёт smoke | `results/runs/<id>/smoke-report.md`, `results/latest/` |
| Индекс прогонов | `results/INDEX.md` |
| Протоколы фаз | `results/manual/` |
| Go/No-Go | `results/manual/go-no-go.md` |
| Протокол ручного прогона | Confluence/Jira / markdown по фазам |
| Баги | Jira/GitHub Issues |
| Go/No-Go | Sign-off в релизном тикете |

---

## 14. Инструменты

| Инструмент | Назначение |
|------------|------------|
| curl / httpie | HTTP API |
| psql | SQL verification |
| redis-cli | Refresh tokens |
| `scripts/run-api-tests.sh` | Auto smoke |
| `scripts/db-check.sh` | Entity lookup |
| `fixtures/apply-fixtures.sh` | Seed QA data |
| kafka-console-consumer | Debug events (optional) |
| jwt.io / base64 decode | JWT payload |
| Playwright | UI E2E (optional) |

---

## 15. Трассировка: фаза → документы

| Фаза | test-cases.md | detailed-test-cases.md | Fixtures | Results |
|------|---------------|------------------------|----------|---------|
| 0 Setup | — | — | apply-fixtures | `results/manual/phase-0-setup.md` |
| 1 Smoke | все `*/test-cases.md` P0 | — | apply-fixtures | `results/latest/smoke-report.md` |
| 2 Employee | customers, vehicles, … | same + auth, deals | 01–04 | `results/manual/phase-2-employee.md` |
| 3 Parts/WO | parts, work-orders, works, employees | + integration E2E-003 | 03, 05, 08 | `results/manual/phase-3-parts-wo.md` |
| 4 Client | client-* | registration, reviews | 02, 06 | `results/manual/phase-4-client.md` |
| 5 Stats | *-statistics, reviews | EST, CST, EREV | 07 | `results/manual/phase-5-stats.md` |
| 6 Security | gateway, auth | E2E-004–006 | test-users | `results/manual/phase-6-regression.md` |

---

## 16. Расписание (ориентир)

| Фаза | Длительность | Зависимости |
|------|--------------|-------------|
| 0 Подготовка | 0.5 д | Dev: стенд up |
| 1 Smoke | 0.5 д | Фаза 0 |
| 2 Employee | 1.5 д | Фаза 1 pass |
| 3 Parts/WO | 1.5 д | workorders up |
| 4 Client | 1 д | Фаза 2 (vehicle VIN) |
| 5 Stats/telemetry | 0.5 д | Kafka |
| 6 Regression | 1 д | Bugfixes |
| **Итого** | **~6–7 д** | 1 QA FTE |

*Smoke-only (P0 list §10): 1 день.*

---

## 17. Ссылки

| Документ | Путь |
|----------|------|
| README QA | [`README.md`](README.md) |
| **Результаты** | [`results/README.md`](results/README.md) |
| Индекс прогонов | [`results/INDEX.md`](results/INDEX.md) |
| Предусловия | [`_shared/prerequisites.md`](_shared/prerequisites.md) |
| DB verification | [`_shared/db-verification.md`](_shared/db-verification.md) |
| Redis/Kafka | [`_shared/redis-kafka-checks.md`](_shared/redis-kafka-checks.md) |
| Fixtures | [`fixtures/README.md`](fixtures/README.md) |
| Manifest IDs | [`fixtures/manifest.json`](fixtures/manifest.json) |
| E2E flows | [`integration/detailed-test-cases.md`](integration/detailed-test-cases.md) |

---

*Документ поддерживается вместе с `qa/api-testing/`. При добавлении сервисов — обновить §4 и фазы.*
