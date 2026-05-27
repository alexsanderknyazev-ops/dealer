# Проверка выполнения backlog задач (Quick Wins)

Источник задач: `backlog/quick-wins/backend.md`, `backlog/quick-wins/frontend.md`.

Статусы:
- `Выполнено` — DoD подтверждается кодом/конфигом.
- `Частично` — есть реализация, но DoD закрыт не полностью.
- `Не выполнено` — подтверждений реализации нет.

## Backend

| ID | Статус | Подтверждение |
|---|---|---|
| BE-01 | Выполнено | Убраны небезопасные дефолты секретов (`services/*/internal/config/config.go`, `docker-compose.yml`, `k8s/dealer-stack.yaml`), сервисы валидируют обязательные `POSTGRES_DSN`/`JWT_SECRET` при старте (`services/*/main.go`), deploy-stage в CI требует `POSTGRES_PASSWORD`/`JWT_SECRET` (`Jenkinsfile`). |
| BE-02 | Выполнено | В `deals`, `customers`, `parts` для write-операций внедрена роль-ориентированная проверка и `403` при запрете (`services/*/internal/httpapi/handler.go`) + покрытие тестами (`handler_test.go`, `auth_test.go`). |
| BE-03 | Выполнено | Есть обязательный CI stage `Go lint (changed services)` + `.golangci.yml` baseline + блокировка пайплайна при ошибках lint. |
| BE-04 | Выполнено | Метрики `request_count`, `error_rate`, `p95` для бизнес-путей auth/deals формализованы PromQL-наборами в `docs/monitoring/auth-deals-promql.md`; `/metrics` публикуется через `pkg/metrics` + `observe.RegisterHTTP`. |
| BE-05 | Выполнено | При создании сделки добавлена проверка существования `customer_id`/`vehicle_id` (`services/deals/internal/service/deal_service.go`, `repository/deal_repository.go`), возвращаются понятные `4xx` ошибки и есть тесты (`deal_service_test.go`). |
| BE-06 | Выполнено | Ролевая политика распространена на write-операции `deals/customers/parts`, есть единая модель запрета (`403`) и аудит-логи отказов (`rbac deny ...` в `handler.go`). |
| BE-07 | Выполнено | Для `auth` и `deals` выставлено `replicas: 2`, добавлены `readiness/liveness` probes и rolling update стратегия без `maxUnavailable` (`services/auth/k8s/auth-deployment.yaml`, `services/deals/k8s/deals-deployment.yaml`). |
| BE-08 | Выполнено | В CI добавлен coverage threshold gate для `services/auth` и `services/deals` с параметрами `AUTH_COVERAGE_MIN`/`DEALS_COVERAGE_MIN` (`Jenkinsfile` stage `Coverage threshold (auth + deals)`), pipeline падает ниже порога. |

## Frontend

| ID | Статус | Подтверждение |
|---|---|---|
| FE-01 | Частично | Для `deals` реализован централизованный маппинг `401/403` + logout/redirect (`frontend/auth/src/dealsApi.ts`, `DealForm.tsx`, `Deals.tsx`), но не видно подтверждения покрытия всех ключевых экранов/сценариев тестами. |
| FE-02 | Выполнено | Есть клиентская валидация формы сделки, блокировка submit, консистентные сообщения (`frontend/auth/src/DealForm.tsx`). |
| FE-03 | Не выполнено | В `frontend/auth/package.json` нет `eslint`/lint-скрипта; в Jenkins нет frontend lint stage. |
| FE-04 | Не выполнено | E2E-инструменты/скрипты (`playwright`/`cypress`) и smoke-сценарий "Логин -> Создание сделки" в CI не найдены. |
| FE-05 | Не выполнено | Негативные e2e по токенам/ролям/API ошибкам не найдены. |
| FE-06 | Частично | Для `deals` есть нормализация некоторых бизнес-ошибок (`401/403`), но единой схемы по всем API/кодам нет. |
| FE-07 | Не выполнено | Централизованный сбор JS ошибок и отчеты/дашборды по frontend latency не найдены. |
