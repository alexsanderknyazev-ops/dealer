# Предусловия для API-тестирования

## Инфраструктура

- PostgreSQL `:5433`, Redis `:6379`, Kafka `:9092`
- Все сервисы из `docker-compose.yml` в статусе `healthy` / `running`
- Миграции применены (`make migrate` или аналог)

## Переменные окружения (скрипт)

| Переменная | По умолчанию |
|------------|--------------|
| `EMPLOYEE_API` | `http://127.0.0.1:8090` |
| `EMPLOYEE_AUTH` | `http://127.0.0.1:8080` |
| `CLIENT_PUBLIC` | `http://127.0.0.1:8091` |
| `CLIENT_PROTECTED` | `http://127.0.0.1:8093` |
| `ERRORS_INGEST` | `http://127.0.0.1:8092` |
| `QA_PASSWORD` | `Test1234!` |

## Тестовые данные

- При регистрации employee через `/api/register` роль по умолчанию: **sales**
- Для client registration обязателен **VIN** существующего авто в `vehicles-service`
- В seed-данных есть brands (Hyundai, Toyota и др.) — id вида `40000000-0000-4000-8000-000000000003`

## Health-check каждого сервиса

```bash
curl -sf http://127.0.0.1:8090/healthz   # gateway
curl -sf http://127.0.0.1:8080/healthz   # auth
curl -sf http://127.0.0.1:8091/healthz   # client-public-gateway (если есть)
```

## Заголовки

```http
Content-Type: application/json
Authorization: Bearer <access_token>
```
