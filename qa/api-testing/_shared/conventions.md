# Соглашения по тест-кейсам

## Шаблон кейса

| Поле | Описание |
|------|----------|
| **ID** | Уникальный идентификатор |
| **Priority** | P0 / P1 / P2 |
| **Endpoint** | HTTP method + path |
| **Auth** | None / Bearer employee / Bearer client |
| **Preconditions** | Данные, роль, состояние системы |
| **Steps** | Пошаговые действия |
| **Expected** | HTTP-код, тело, side-effects (Kafka, БД) |
| **Auto** | ID в `run-api-tests.sh` или `manual` |
| **DB verify** | SQL + `scripts/db-check.sh` |
| **Async wait** | Kafka consumers: до **15 с** перед SQL |

Детальные сценарии: `detailed-test-cases.md` в каждой папке сервиса.

## Коды ответов (ожидаемые)

| Код | Когда |
|-----|-------|
| 200/201 | Успех |
| 204 | Logout, telemetry |
| 400 | Валидация / protobuf invalid |
| 401 | Нет или битый JWT |
| 403 | RBAC / недостаточно прав |
| 404 | Сущность не найдена |
| 409/500 | Конфликт уникальности, внутренняя ошибка |
| 503 | Backend недоступен (gateway → gRPC down / circuit breaker) |

## Проверка side-effects

- **Kafka** `client.registration.v1` → user появляется в client-auth (async, до ~5 с)
- **Kafka** `review.published.v1` → отзыв в employee-reviews
- **Kafka** `deal.completed.v1` → метрики employee-statistics
- **Movement confirm** → stock уменьшается, work order `parts_issued=true`
